package pg

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/graph"
	"strconv"
	"strings"
	"sync"
)

const (
	nodeTableName = "node"
	edgeTableName = "edge"
)

const fetchPartitionSQL = `
select inhrelid::regclass as child from pg_catalog.pg_inherits
where inhparent = @parent::regclass
`

const graphSchemaSQLUp = `
create table if not exists graph (
    id serial,
    name varchar(256) not null,

    primary key (id),
    unique (name)
);

create table if not exists kind (
	id smallserial,
    name varchar(256) not null,

	primary key (id),
    unique (name)
);

create table if not exists node (
    id serial,
	graph_id integer not null references graph(id),
	kind_ids smallint[8] not null,
	properties jsonb not null,

	primary key (id, graph_id)
) partition by list (graph_id);

alter table node alter column properties set storage main;

create index if not exists node_graph_id_index on node using btree (graph_id);
create index if not exists node_kind_ids_index on node using gin (kind_ids);

create table if not exists edge (
    id serial,
    graph_id integer not null references graph(id),
    start_id integer not null,
    end_id integer not null,
	kind_id smallint not null,
	properties jsonb not null,

	primary key (id, graph_id)
) partition by list (graph_id);

alter table edge alter column properties set storage main;

create index if not exists edge_graph_id_index on edge using btree (graph_id);
create index if not exists edge_start_id_index on edge using btree (start_id);
create index if not exists edge_end_id_index on edge using btree (end_id);
create index if not exists edge_kind_index on edge using btree (kind_id);
`

const graphSchemaSQLDown = `
drop table if exists node;
drop table if exists edge;
drop table if exists kind;
drop table if exists graph;
`

func formatPartitionTableName(parent string, graphID int32) string {
	return parent + "_" + strconv.FormatInt(int64(graphID), 10)
}

func formatPartitionTableSQL(parent string, graphID int32) string {
	var (
		graphIDStr = strconv.FormatInt(int64(graphID), 10)
		builder    = strings.Builder{}
	)

	builder.WriteString("create table ")
	builder.WriteString(parent)
	builder.WriteString("_")
	builder.WriteString(graphIDStr)
	builder.WriteString(" partition of ")
	builder.WriteString(parent)
	builder.WriteString(" for values in (")
	builder.WriteString(graphIDStr)
	builder.WriteString(")")

	return builder.String()
}

type Graph struct {
	ID            int32
	Name          string
	NodePartition string
	EdgePartition string
}

type SchemaManager struct {
	graphs map[string]Graph
	kinds  map[graph.Kind]int16
	lock   *sync.RWMutex
}

func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		graphs: map[string]Graph{},
		kinds:  map[graph.Kind]int16{},
		lock:   &sync.RWMutex{},
	}
}

func (s *SchemaManager) fetchGraphs(tx graph.Transaction) error {
	var (
		graphID   int32
		graphName string

		graphs = map[string]Graph{}
		result = tx.Run(`select id, name from graph`, nil)
	)

	defer result.Close()

	for result.Next() {
		if err := result.Scan(&graphID, &graphName); err != nil {
			return err
		} else {
			graphs[graphName] = Graph{
				ID:            graphID,
				Name:          graphName,
				NodePartition: formatPartitionTableName(nodeTableName, graphID),
				EdgePartition: formatPartitionTableName(edgeTableName, graphID),
			}
		}
	}

	s.graphs = graphs
	return result.Error()
}

func (s *SchemaManager) validatePartitions(tx graph.Transaction) error {
	return nil
}

func (s *SchemaManager) fetchKinds(tx graph.Transaction) error {
	var (
		kindID   int16
		kindName string

		kinds  = map[graph.Kind]int16{}
		result = tx.Run(`select id, name from kind`, nil)
	)

	defer result.Close()

	for result.Next() {
		if err := result.Scan(&kindID, &kindName); err != nil {
			return err
		}

		kinds[graph.StringKind(kindName)] = kindID
	}

	s.kinds = kinds
	return result.Error()
}

func (s *SchemaManager) fetch(tx graph.Transaction) error {
	if err := s.fetchGraphs(tx); err != nil {
		return err
	}

	if err := s.validatePartitions(tx); err != nil {
		return err
	}

	return s.fetchKinds(tx)
}

func (s *SchemaManager) defineKind(tx graph.Transaction, kind graph.Kind) (int16, error) {
	var (
		kindID int16
		result = tx.Run(`insert into kind (name) values (@name) returning id`, map[string]any{
			"name": kind.String(),
		})
	)

	defer result.Close()

	if !result.Next() {
		return -1, pgx.ErrNoRows
	}

	if err := result.Scan(&kindID); err != nil {
		return -1, err
	}

	return kindID, result.Error()
}

func (s *SchemaManager) createGraphPartition(tx graph.Transaction, parent string, graphID int32) error {
	result := tx.Run(formatPartitionTableSQL(parent, graphID), nil)
	defer result.Close()

	return result.Error()
}

func (s *SchemaManager) createGraphPartitions(tx graph.Transaction, graphID int32) error {
	if err := s.createGraphPartition(tx, nodeTableName, graphID); err != nil {
		return err
	}

	return s.createGraphPartition(tx, edgeTableName, graphID)
}

func (s *SchemaManager) createGraph(tx graph.Transaction, graphName string, graphSchema graph.Graph) (Graph, error) {
	var (
		graphID int32
		result  = tx.Run(`insert into graph (name) values (@name) returning id`, map[string]any{
			"name": graphName,
		})
	)

	defer result.Close()

	if !result.Next() {
		return Graph{}, fmt.Errorf("no ID returned from graph entry creation")
	}

	if err := result.Scan(&graphID); err != nil {
		return Graph{}, fmt.Errorf("failed mapping ID from graph entry creation: %w", err)
	}

	return Graph{
		ID:            graphID,
		Name:          graphName,
		NodePartition: formatPartitionTableName(nodeTableName, graphID),
		EdgePartition: formatPartitionTableName(edgeTableName, graphID),
	}, nil
}

func (s *SchemaManager) defineKinds(tx graph.Transaction, kinds graph.Kinds) error {
	for _, kind := range kinds {
		if kindID, err := s.defineKind(tx, kind); err != nil {
			return err
		} else {
			s.kinds[kind] = kindID
		}
	}

	return nil
}

func (s *SchemaManager) missingKinds(kinds graph.Kinds) graph.Kinds {
	var missingKinds graph.Kinds

	for _, kind := range kinds {
		if _, isDefined := s.kinds[kind]; !isDefined {
			missingKinds = append(missingKinds, kind)
		}
	}

	return missingKinds
}

func (s *SchemaManager) defineGraphKinds(tx graph.Transaction, graphSchemas []graph.Graph) error {
	for _, graphSchema := range graphSchemas {
		if err := s.defineKinds(tx, s.missingKinds(graphSchema.Kinds)); err != nil {
			return err
		}
	}

	return nil
}

func (s *SchemaManager) defineGraph(tx graph.Transaction, graphName string, graphSchema graph.Graph) error {
	if graphDefinition, err := s.createGraph(tx, graphName, graphSchema); err != nil {
		return err
	} else {
		s.graphs[graphName] = graphDefinition
		return s.createGraphPartitions(tx, graphDefinition.ID)
	}
}

func (s *SchemaManager) kindIDs(kinds graph.Kinds) []int16 {
	ids := make([]int16, 0, len(kinds))

	for _, kind := range kinds {
		if id, hasID := s.kinds[kind]; hasID {
			ids = append(ids, id)
		}
	}

	return ids
}

func (s *SchemaManager) AssertKinds(tx graph.Transaction, kinds graph.Kinds) ([]int16, error) {
	// Acquire a read-lock first to fast-pass validate if we're missing any kind definitions
	s.lock.RLock()

	if missingKinds := s.missingKinds(kinds); len(missingKinds) == 0 {
		// All kinds are defined. Release the read-lock here before returning
		s.lock.RUnlock()
		return s.kindIDs(kinds), nil
	}

	// Release the read-lock here so that we can acquire a write-lock
	s.lock.RUnlock()

	// Acquire a write-lock and release on-exit
	s.lock.Lock()
	defer s.lock.Unlock()

	// We have to re-acquire the missing kinds since there's a potential for another writer to acquire the write-lock
	// inbetween release of the read-lock and acquisition of the write-lock for this operation
	if err := s.defineKinds(tx, s.missingKinds(kinds)); err != nil {
		return nil, err
	}

	return s.kindIDs(kinds), nil
}

func (s *SchemaManager) AssertGraph(tx graph.Transaction, graphName string, graphSchema graph.Graph) (Graph, error) {
	// Acquire a read-lock first to fast-pass validate if we're missing the graph definitions
	s.lock.RLock()

	if graphInstance, isDefined := s.graphs[graphName]; isDefined {
		// The graph is defined. Release the read-lock here before returning
		s.lock.RUnlock()
		return graphInstance, nil
	}

	// Release the read-lock here so that we can acquire a write-lock
	s.lock.RUnlock()

	// Acquire a write-lock and create the graph definition
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.defineGraph(tx, graphName, graphSchema); err != nil {
		return Graph{}, err
	}

	return s.graphs[graphName], nil
}

func (s *SchemaManager) AssertSchema(tx graph.Transaction, dbSchema graph.Schema) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.fetch(tx); err != nil {
		return err
	}

	for _, graphSchema := range dbSchema.Graphs {
		if missingKinds := s.missingKinds(graphSchema.Kinds); len(missingKinds) > 0 {
			if err := s.defineKinds(tx, missingKinds); err != nil {
				return err
			}
		}
	}

	return nil
}
