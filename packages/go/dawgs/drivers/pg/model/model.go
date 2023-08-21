package model

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"strings"
)

type IndexChangeSet struct {
	NodeIndexesToRemove     []string
	NodeConstraintsToRemove []string
	NodeIndexesToAdd        map[string]graph.Index
	NodeConstraintsToAdd    map[string]graph.Constraint
}

func NewIndexChangeSet() IndexChangeSet {
	return IndexChangeSet{
		NodeIndexesToAdd:     map[string]graph.Index{},
		NodeConstraintsToAdd: map[string]graph.Constraint{},
	}
}

type GraphPartition struct {
	Name        string
	Indexes     map[string]graph.Index
	Constraints map[string]graph.Constraint
}

func NewGraphPartition() GraphPartition {
	return GraphPartition{
		Indexes:     map[string]graph.Index{},
		Constraints: map[string]graph.Constraint{},
	}
}

func NewGraphPartitionWithName(name string) GraphPartition {
	return GraphPartition{
		Name:        name,
		Indexes:     map[string]graph.Index{},
		Constraints: map[string]graph.Constraint{},
	}
}

func NewGraphPartitionFromSchema(name string, indexes []graph.Index, constraints []graph.Constraint) GraphPartition {
	graphPartition := GraphPartition{
		Name:        name,
		Indexes:     make(map[string]graph.Index, len(indexes)),
		Constraints: make(map[string]graph.Constraint, len(constraints)),
	}

	for _, index := range indexes {
		graphPartition.Indexes[formatIndexName(name, index)] = index
	}

	for _, constraint := range constraints {
		graphPartition.Constraints[formatConstraintName(name, constraint)] = constraint
	}

	return graphPartition
}

type Graph struct {
	ID            int32
	Name          string
	NodePartition GraphPartition
	EdgePartition GraphPartition
}

func pgIndexTypeToString(indexType graph.IndexType) string {
	switch indexType {
	case graph.BTreeIndex:
		return pgIndexTypeBTree
	case graph.TextSearchIndex:
		return pgIndexTypeGIN
	default:
		return "NOT SUPPORTED"
	}
}

func pgParseIndexType(pgType string) graph.IndexType {
	switch strings.ToLower(pgType) {
	case pgIndexTypeBTree:
		return graph.BTreeIndex
	case pgIndexTypeGIN:
		return graph.TextSearchIndex
	default:
		return graph.UnsupportedIndex
	}
}
