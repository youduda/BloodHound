package model

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/graph"
	"sync"
)

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
				NodePartition: NewGraphPartitionWithName(formatPartitionTableName(pgNodeTableName, graphID)),
				EdgePartition: NewGraphPartitionWithName(formatPartitionTableName(pgEdgeTableName, graphID)),
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
	if err := s.createGraphPartition(tx, pgNodeTableName, graphID); err != nil {
		return err
	}

	return s.createGraphPartition(tx, pgEdgeTableName, graphID)
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
		NodePartition: NewGraphPartitionWithName(formatPartitionTableName(pgNodeTableName, graphID)),
		EdgePartition: NewGraphPartitionWithName(formatPartitionTableName(pgEdgeTableName, graphID)),
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
		if err := s.defineKinds(tx, s.missingKinds(graphSchema.Nodes)); err != nil {
			return err
		}

		if err := s.defineKinds(tx, s.missingKinds(graphSchema.Edges)); err != nil {
			return err
		}
	}

	return nil
}

func (s *SchemaManager) defineGraph(tx graph.Transaction, graphName string, graphSchema graph.Graph) error {
	if graphDefinition, err := s.createGraph(tx, graphName, graphSchema); err != nil {
		return err
	} else if err := s.createGraphPartitions(tx, graphDefinition.ID); err != nil {
		return err
	} else if err := assertGraphPartitions(tx, graphSchema, graphDefinition); err != nil {
		return err
	} else {
		s.graphs[graphName] = graphDefinition

		return nil
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
		if missingKinds := s.missingKinds(graphSchema.Nodes); len(missingKinds) > 0 {
			if err := s.defineKinds(tx, missingKinds); err != nil {
				return err
			}
		}

		if missingKinds := s.missingKinds(graphSchema.Edges); len(missingKinds) > 0 {
			if err := s.defineKinds(tx, missingKinds); err != nil {
				return err
			}
		}
	}

	return nil
}
