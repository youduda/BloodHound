package pg

import (
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/query"
	"github.com/specterops/bloodhound/dawgs/graph"
	"sync"
)

type SchemaManager struct {
	graphs map[string]model.Graph
	kinds  map[graph.Kind]int16
	lock   *sync.RWMutex
}

func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		graphs: map[string]model.Graph{},
		kinds:  map[graph.Kind]int16{},
		lock:   &sync.RWMutex{},
	}
}

func (s *SchemaManager) fetch(tx graph.Transaction) error {
	if kinds, err := query.On(tx).SelectKinds(); err != nil {
		return err
	} else {
		s.kinds = kinds
	}

	return nil
}

func (s *SchemaManager) defineKinds(tx graph.Transaction, kinds graph.Kinds) error {
	for _, kind := range kinds {
		if kindID, err := query.On(tx).InsertKind(kind); err != nil {
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

func (s *SchemaManager) defineGraphKinds(tx graph.Transaction, schemas []graph.Graph) error {
	for _, schema := range schemas {
		if err := s.defineKinds(tx, s.missingKinds(schema.Nodes)); err != nil {
			return err
		}

		if err := s.defineKinds(tx, s.missingKinds(schema.Edges)); err != nil {
			return err
		}
	}

	return nil
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
		kindIDs := s.kindIDs(kinds)
		s.lock.RUnlock()

		return kindIDs, nil
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

func (s *SchemaManager) AssertGraph(tx graph.Transaction, schema graph.Graph) (model.Graph, error) {
	// Acquire a read-lock first to fast-pass validate if we're missing the graph definitions
	s.lock.RLock()

	if graphInstance, isDefined := s.graphs[schema.Name]; isDefined {
		// The graph is defined. Release the read-lock here before returning
		s.lock.RUnlock()
		return graphInstance, nil
	}

	// Release the read-lock here so that we can acquire a write-lock
	s.lock.RUnlock()

	// Acquire a write-lock and create the graph definition
	s.lock.Lock()
	defer s.lock.Unlock()

	if graphInstance, isDefined := s.graphs[schema.Name]; isDefined {
		// The graph was defined by a different actor between the read unlock and the write lock.
		return graphInstance, nil
	}

	// Validate the schema if the graph already exists in the database
	if definition, err := query.On(tx).SelectGraphByName(schema.Name); err != nil {
		// ErrNoRows signifies that this graph must be created
		if err != pgx.ErrNoRows {
			return model.Graph{}, err
		}
	} else if definition, err := query.On(tx).AssertGraph(schema, definition); err != nil {
		return model.Graph{}, err
	} else {
		s.graphs[schema.Name] = definition
		return definition, nil
	}

	// Create the graph
	if definition, err := query.On(tx).CreateGraph(schema); err != nil {
		return model.Graph{}, err
	} else {
		s.graphs[schema.Name] = definition
		return definition, nil
	}
}

func (s *SchemaManager) AssertSchema(tx graph.Transaction, schema graph.Schema) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.fetch(tx); err != nil {
		return err
	}

	for _, graphSchema := range schema.Graphs {
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
