package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type transaction struct {
	schemaManager     *model.SchemaManager
	ctx               context.Context
	tx                pgx.Tx
	targetGraphSet    bool
	targetGraphName   string
	targetGraphSchema graph.Graph
}

func newTransaction(ctx context.Context, conn *pgxpool.Conn, options pgx.TxOptions, schemaManager *model.SchemaManager) (*transaction, error) {
	if pgxTx, err := conn.BeginTx(ctx, options); err != nil {
		return nil, err
	} else {
		return &transaction{
			schemaManager:  schemaManager,
			ctx:            ctx,
			tx:             pgxTx,
			targetGraphSet: false,
		}, nil
	}
}

func (s *transaction) WithGraph(graphName string, graphSchema graph.Graph) graph.Transaction {
	s.targetGraphSet = true
	s.targetGraphName = graphName
	s.targetGraphSchema = graphSchema

	return s
}

func (s *transaction) Close() {
	if s.tx != nil {
		s.tx.Rollback(s.ctx)
		s.tx = nil
	}
}

func (s *transaction) getTargetGraph() (model.Graph, error) {
	if !s.targetGraphSet {
		return model.Graph{}, fmt.Errorf("driver operation requires a graph target to be set")
	}

	return s.schemaManager.AssertGraph(s, s.targetGraphName, s.targetGraphSchema)
}

func (s *transaction) CreateNode(properties *graph.Properties, kinds ...graph.Kind) (*graph.Node, error) {
	if graphTarget, err := s.getTargetGraph(); err != nil {
		return nil, err
	} else if kindIDSlice, err := s.schemaManager.AssertKinds(s, kinds); err != nil {
		return nil, err
	} else {
		var (
			nodeID int32
			result = s.tx.QueryRow(s.ctx, `insert into node (graph_id, kind_ids, properties) values (@graph_id, @kind_ids, @properties) returning id`, pgx.NamedArgs{
				"graph_id":   graphTarget.ID,
				"kind_ids":   kindIDSlice,
				"properties": properties.MapOrEmpty(),
			})
		)

		if err := result.Scan(&nodeID); err != nil {
			return nil, err
		}

		return graph.NewNode(graph.ID(nodeID), properties, kinds...), err
	}
}

func (s *transaction) UpdateNode(node *graph.Node) error {
	if kindIDSlice, err := s.schemaManager.AssertKinds(s, node.Kinds); err != nil {
		return err
	} else {
		_, err := s.tx.Exec(s.ctx, `update node set kind_ids = @kind_ids, properties = @properties where id = @node_id`, pgx.NamedArgs{
			"node_id":    node.ID,
			"kind_ids":   kindIDSlice,
			"properties": node.Properties.MapOrEmpty(),
		})

		return err
	}
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
	} else if kindIDSlice, err := s.schemaManager.AssertKinds(s, graph.Kinds{kind}); err != nil {
		return nil, err
	} else {
		var (
			edgeID int32
			result = s.tx.QueryRow(s.ctx, `insert into edge (graph_id, start_id, end_id, kind_id, properties) values (@graph_id, @start_id, @end_id, @kind_id, @properties) returning id`, pgx.NamedArgs{
				"graph_id":   graphTarget.ID,
				"start_id":   startNodeID,
				"end_id":     endNodeID,
				"kind_id":    kindIDSlice[0],
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
	if kindIDSlice, err := s.schemaManager.AssertKinds(s, graph.Kinds{relationship.Kind}); err != nil {
		return err
	} else {
		_, err := s.tx.Exec(s.ctx, `update edge set kind_id = @kind_id, properties = @properties where id = @edge_id`, pgx.NamedArgs{
			"edge_id":    relationship.ID,
			"kind_id":    kindIDSlice[0],
			"properties": relationship.Properties.MapOrEmpty(),
		})

		return err
	}
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