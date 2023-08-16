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

func stringBuilderAppend(builder *strings.Builder, strings ...string) {
	for idx := 0; idx < len(strings); idx++ {
		builder.WriteString(strings[idx])
	}
}

func formatPartitionTableName(parent string, graphID int32) string {
	return parent + "_" + strconv.FormatInt(int64(graphID), 10)
}

func formatPartitionTableSQL(parent string, graphID int32) string {
	var (
		graphIDStr = strconv.FormatInt(int64(graphID), 10)
		builder    = strings.Builder{}
	)

	stringBuilderAppend(
		&builder,
		"create table ", parent, "_", graphIDStr, " partition of ", parent,
		" for values in (", graphIDStr, ")")

	return builder.String()
}

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
	graph_id integer not null,
	kinds smallint[8],

	primary key (id, graph_id)
)

partition by list (graph_id);

create table edge (
    id serial,
	graph_id integer not null,
	kind smallint references kind(id),

	primary key (id, graph_id)
)

partition by list (graph_id);
`

const graphSchemaSQLDown = `
drop table if exists node;
drop table if exists edge;
drop table if exists kind;
drop table if exists graph;
`

type Schema struct {
	Graphs map[string]int32
	Kinds  map[graph.Kind]int16
}

func (s *Schema) fetchGraphs(tx graph.Transaction) error {
	var (
		graphID int32
		name    string

		graphs = map[string]int32{}
		result = tx.Run(`select id, name from graph`, nil)
	)

	defer result.Close()

	for result.Next() {
		if err := result.Scan(&graphID, &name); err != nil {
			return err
		} else {
			graphs[name] = graphID
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

func (s *Schema) createGraph(tx graph.Transaction, graphSchema *graph.GraphSchema) (int32, error) {
	var (
		graphID int32
		result  = tx.Run(`insert into graph (name) values (@name) returning id`, map[string]any{
			"name": graphSchema.Name,
		})
	)

	defer result.Close()

	if !result.Next() {
		return -1, fmt.Errorf("no ID returned from graph entry creation")
	}

	if err := result.Scan(&graphID); err != nil {
		return 1, fmt.Errorf("failed mapping ID from graph entry creation: %w", err)
	}

	s.Graphs[graphSchema.Name] = graphID
	return graphID, nil
}

func (s *Schema) getOrCreateGraph(tx graph.Transaction, graphSchema *graph.GraphSchema) (int32, error) {
	if graphID, hasGraph := s.Graphs[graphSchema.Name]; hasGraph {
		return graphID, nil
	}

	if graphID, err := s.createGraph(tx, graphSchema); err != nil {
		return -1, err
	} else {
		return graphID, s.createGraphPartitions(tx, graphID)
	}
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

	if _, err := s.getOrCreateGraph(tx, graphSchema); err != nil {
		return err
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

func (s *Schema) IDs(kinds ...graph.Kind) ([]int16, bool) {
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
