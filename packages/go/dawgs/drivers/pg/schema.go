package pg

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/graph"
	"strconv"
	"strings"
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
create table graph (
    id serial,
    name varchar(256) not null,

    primary key (id),
    unique (name)
);

create table kind (
	id smallserial,
    name varchar(256) not null,

	primary key (id),
    unique (name)
);

create table node (
    id serial,
	graph_id integer not null references graph(id),
	kind_ids smallint[8] not null,
	properties jsonb not null,

	primary key (id, graph_id)
) partition by list (graph_id);

alter table node alter column properties set storage main;

create index node_graph_id_index on node using btree (graph_id);
create index node_kind_ids_index on node using gin (kind_ids);

create table edge (
    id serial,
    graph_id integer not null references graph(id),
    start_id integer not null,
    end_id integer not null,
	kind_id smallint not null,
	properties jsonb not null,

	primary key (id, graph_id)
) partition by list (graph_id);

alter table edge alter column properties set storage main;

create index edge_graph_id_index on edge using btree (graph_id);
create index edge_start_id_index on edge using btree (start_id);
create index edge_end_id_index on edge using btree (end_id);
create index edge_kind_index on edge using btree (kind_id);
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

func NewGraph(id int32, name string) Graph {
	return Graph{
		ID:            id,
		Name:          name,
		NodePartition: formatPartitionTableName(nodeTableName, id),
		EdgePartition: formatPartitionTableName(edgeTableName, id),
	}
}

type Schema struct {
	Graphs map[string]Graph
	Kinds  map[graph.Kind]int16
}

func (s *Schema) fetchGraphs(tx graph.Transaction) error {
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

	s.Graphs = graphs
	return result.Error()
}

func (s *Schema) validatePartitions(tx graph.Transaction) error {
	return nil
}

func (s *Schema) fetchKinds(tx graph.Transaction) error {
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

	s.Kinds = kinds
	return result.Error()
}

func (s *Schema) Fetch(tx graph.Transaction) error {
	if err := s.fetchGraphs(tx); err != nil {
		return err
	}

	if err := s.validatePartitions(tx); err != nil {
		return err
	}

	if err := s.fetchKinds(tx); err != nil {
		return err
	}

	return nil
}

func (s *Schema) DefineKind(tx graph.Transaction, kind graph.Kind) (int16, error) {
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

func (s *Schema) createGraphPartition(tx graph.Transaction, parent string, graphID int32) error {
	result := tx.Run(formatPartitionTableSQL(parent, graphID), nil)
	defer result.Close()

	return result.Error()
}

func (s *Schema) createGraphPartitions(tx graph.Transaction, graphID int32) error {
	if err := s.createGraphPartition(tx, nodeTableName, graphID); err != nil {
		return err
	}

	return s.createGraphPartition(tx, edgeTableName, graphID)
}

func (s *Schema) createGraph(tx graph.Transaction, graphSchema *graph.GraphSchema) (Graph, error) {
	var (
		graphID int32
		result  = tx.Run(`insert into graph (name) values (@name) returning id`, map[string]any{
			"name": graphSchema.Name,
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
		Name:          graphSchema.Name,
		NodePartition: formatPartitionTableName(nodeTableName, graphID),
		EdgePartition: formatPartitionTableName(edgeTableName, graphID),
	}, nil
}

func (s *Schema) DefineGraph(tx graph.Transaction, graphSchema *graph.GraphSchema) error {
	for kind := range graphSchema.Kinds {
		if _, isDefined := s.Kinds[kind]; !isDefined {
			if kindID, err := s.DefineKind(tx, kind); err != nil {
				return err
			} else {
				s.Kinds[kind] = kindID
			}
		}
	}

	if _, isDefined := s.Graphs[graphSchema.Name]; !isDefined {
		if newGraphInst, err := s.createGraph(tx, graphSchema); err != nil {
			return err
		} else if err := s.createGraphPartitions(tx, newGraphInst.ID); err != nil {
			return err
		} else {
			s.Graphs[newGraphInst.Name] = newGraphInst
		}
	}

	return nil
}

func (s *Schema) Define(tx graph.Transaction, databaseSchema *graph.DatabaseSchema) error {
	if err := s.Fetch(tx); err != nil {
		return err
	}

	for _, graphSchema := range databaseSchema.Graphs {
		if err := s.DefineGraph(tx, graphSchema); err != nil {
			return err
		}
	}

	return s.Fetch(tx)
}

func (s *Schema) KindIDs(kinds ...graph.Kind) ([]int16, bool) {
	ids := make([]int16, len(kinds))

	for idx, kind := range kinds {
		if id, hasID := s.Kinds[kind]; !hasID {
			return nil, false
		} else {
			ids[idx] = id
		}
	}

	return ids, true
}
