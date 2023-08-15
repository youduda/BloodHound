package pg

import (
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/graph"
)

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
	Graphs         map[string]int16
	Kinds          map[graph.Kind]int16
	NodePartitions map[string]int16
	EdgePartitions map[string]int16
}

func (s *Schema) fetchGraphs(tx graph.Transaction) error {
	var (
		graphID int16
		name    string

		graphs = map[string]int16{}
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

func (s *Schema) Define(tx graph.Transaction, graphSchema *graph.Schema) error {
	if err := s.Fetch(tx); err != nil {
		return err
	}

	for kind := range graphSchema.Kinds {
		if _, isDefined := s.Kinds[kind]; !isDefined {
			if kindID, err := s.DefineKind(tx, kind); err != nil {
				return err
			} else {
				s.Kinds[kind] = kindID
			}
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
