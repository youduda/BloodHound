package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type transaction struct {
	schema         Schema
	ctx            context.Context
	tx             pgx.Tx
	targetGraph    string
	hasTargetGraph bool
}

func newTransaction(ctx context.Context, conn *pgxpool.Conn, options pgx.TxOptions, currentSchema Schema) (*transaction, error) {
	if pgxTx, err := conn.BeginTx(ctx, options); err != nil {
		return nil, err
	} else {
		return &transaction{
			schema:         currentSchema,
			ctx:            ctx,
			tx:             pgxTx,
			hasTargetGraph: false,
		}, nil
	}
}

func (s *transaction) WithGraph(graphName string) graph.Transaction {
	s.targetGraph = graphName
	s.hasTargetGraph = true

	return s
}

func (s *transaction) Close() {
	if s.tx != nil {
		s.tx.Rollback(s.ctx)
		s.tx = nil
	}
}

func (s *transaction) getTargetGraph() (Graph, error) {
	if !s.hasTargetGraph {
		return Graph{}, fmt.Errorf("postgresql driver requires a graph target to be set")
	} else if targetGraph, hasGraph := s.schema.Graphs[s.targetGraph]; !hasGraph {
		return Graph{}, fmt.Errorf("unknown graph: %s", s.targetGraph)
	} else {
		return targetGraph, nil
	}
}

func (s *transaction) CreateNode(properties *graph.Properties, kinds ...graph.Kind) (*graph.Node, error) {
	if graphTarget, err := s.getTargetGraph(); err != nil {
		return nil, err
	} else if kindIDSlice, hasAllIDs := s.schema.KindIDs(kinds...); !hasAllIDs {
		return nil, fmt.Errorf("unable to map all kinds: %v", kinds)
	} else {
		var (
			nodeID int32
			result = s.tx.QueryRow(s.ctx, `insert into node (graph_id, kind_ids, properties) values (@graph_id, @kind_ids, @properties) returning id`, pgx.NamedArgs{
				"graph_id":   graphTarget.ID,
				"kind_ids":   kindIDSlice,
				"properties": properties.Map,
			})
		)

		if err := result.Scan(&nodeID); err != nil {
			return nil, err
		}

		return graph.NewNode(graph.ID(nodeID), properties, kinds...), err
	}
}

func (s *transaction) UpdateNode(node *graph.Node) error {
	//TODO implement me
	panic("implement me")
}

func (s *transaction) UpdateNodeBy(update graph.NodeUpdate) error {
	//TODO implement me
	panic("implement me")
}

func (s *transaction) Nodes() graph.NodeQuery {
	//TODO implement me
	panic("implement me")
}

func (s *transaction) CreateRelationship(startNode, endNode *graph.Node, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	return nil, fmt.Errorf("unsupported")
}

func (s *transaction) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	if graphTarget, err := s.getTargetGraph(); err != nil {
		return nil, err
	} else if kindID, hasKind := s.schema.Kinds[kind]; !hasKind {
		return nil, fmt.Errorf("unable to map all kind: %s", kind)
	} else {
		var (
			edgeID int32
			result = s.tx.QueryRow(s.ctx, `insert into edge (graph_id, start_id, end_id, kind_id, properties) values (@graph_id, @start_id, @end_id, @kind_id, @properties) returning id`, pgx.NamedArgs{
				"graph_id":   graphTarget.ID,
				"start_id":   startNodeID,
				"end_id":     endNodeID,
				"kind_id":    kindID,
				"properties": properties.MapOrEmpty(),
			})
		)

		if err := result.Scan(&edgeID); err != nil {
			return nil, err
		}

		return graph.NewRelationship(graph.ID(edgeID), startNodeID, endNodeID, properties, kind), err
	}
}

func (s *transaction) UpdateRelationship(relationship *graph.Relationship) error {
	//TODO implement me
	panic("implement me")
}

func (s *transaction) UpdateRelationshipBy(update graph.RelationshipUpdate) error {
	//TODO implement me
	panic("implement me")
}

func (s *transaction) Relationships() graph.RelationshipQuery {
	//TODO implement me
	panic("implement me")
}

func (s *transaction) query(query string, parameters map[string]any) (pgx.Rows, error) {
	if parameters == nil || len(parameters) == 0 {
		return s.tx.Query(s.ctx, query)
	}

	return s.tx.Query(s.ctx, query, pgx.NamedArgs(parameters))
}

func (s *transaction) Run(query string, parameters map[string]any) graph.Result {
	if rows, err := s.query(query, parameters); err != nil {
		return queryError{
			err: err,
		}
	} else {
		return &queryResult{
			rows: rows,
		}
	}
}

func (s *transaction) Commit() error {
	return s.tx.Commit(s.ctx)
}
