package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type transaction struct {
	schema Schema
	ctx    context.Context
	tx     pgx.Tx
}

func newTransaction(ctx context.Context, conn *pgxpool.Conn, options pgx.TxOptions, currentSchema Schema) (*transaction, error) {
	if pgxTx, err := conn.BeginTx(ctx, options); err != nil {
		return nil, err
	} else {
		return &transaction{
			schema: currentSchema,
			ctx:    ctx,
			tx:     pgxTx,
		}, nil
	}
}

func (s *transaction) Close() {
	if s.tx != nil {
		s.tx.Rollback(s.ctx)
		s.tx = nil
	}
}

func (s *transaction) CreateNode(properties *graph.Properties, kinds ...graph.Kind) (*graph.Node, error) {
	if kindIDSlice, hasAllIDs := s.schema.IDs(kinds...); !hasAllIDs {
		return nil, fmt.Errorf("unable to map all kinds to IDs")
	} else {
		_, err := s.tx.Exec(s.ctx, `insert into node (graph_id, kinds) values (@graph_id, @kinds)`, map[string]any{
			"graph_id": 1,
			"kinds":    kindIDSlice,
		})

		return nil, err
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
	//TODO implement me
	panic("implement me")
}

func (s *transaction) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	//TODO implement me
	panic("implement me")
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
