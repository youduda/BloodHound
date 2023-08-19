// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package neo4j

import (
	"context"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
)

const (
	nativeBTreeIndexProvider  = "native-btree-1.0"
	nativeLuceneIndexProvider = "lucene+native-3.0"

	dropPropertyIndexStatement        = "drop index $name;"
	dropPropertyConstraintStatement   = "drop constraint $name;"
	createPropertyIndexStatement      = "call db.createIndex($name, $labels, $properties, $provider);"
	createPropertyConstraintStatement = "call db.createUniquePropertyConstraint($name, $labels, $properties, $provider);"
)

type neo4jIndex struct {
	graph.Index

	kind graph.Kind
}

type neo4jConstraint struct {
	graph.Constraint

	kind graph.Kind
}

func parseProviderType(provider string) graph.IndexType {
	switch provider {
	case nativeBTreeIndexProvider:
		return graph.BTreeIndex
	case nativeLuceneIndexProvider:
		return graph.FullTextSearchIndex
	default:
		return graph.UnsupportedIndex
	}
}

func indexTypeProvider(indexType graph.IndexType) string {
	switch indexType {
	case graph.BTreeIndex:
		return nativeBTreeIndexProvider
	case graph.FullTextSearchIndex:
		return nativeLuceneIndexProvider
	default:
		return ""
	}
}

func allIndexes(graphSchema graph.Graph, dbKinds graph.Kinds) map[string]neo4jIndex {
	indexes := map[string]neo4jIndex{}

	for _, index := range graphSchema.Indexes {
		for _, kind := range dbKinds {
			indexName := index.Name

			if indexName == "" {
				indexName = strings.ToLower(kind.String()) + "_" + strings.ToLower(index.Field) + "_index"
			}

			indexes[indexName] = neo4jIndex{
				Index: index,
				kind:  kind,
			}
		}
	}

	return indexes
}

func allConstraints(graphSchema graph.Graph, dbKinds graph.Kinds) map[string]neo4jConstraint {
	constraints := map[string]neo4jConstraint{}

	for _, constraint := range graphSchema.Constraints {
		for _, kind := range dbKinds {
			constraintName := constraint.Name

			if constraintName == "" {
				constraintName = strings.ToLower(kind.String()) + "_" + strings.ToLower(constraint.Field) + "_constraint"
			}

			constraints[constraintName] = neo4jConstraint{
				Constraint: constraint,
				kind:       kind,
			}
		}
	}

	return constraints
}

func assertIndexes(ctx context.Context, db graph.Database, indexesToRemove []string, indexesToAdd map[string]neo4jIndex) error {
	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		nameMap := map[string]any{}

		for _, indexToRemove := range indexesToRemove {
			nameMap["name"] = indexToRemove

			result := tx.Run(dropPropertyIndexStatement, nameMap)
			result.Close()

			if err := result.Error(); err != nil {
				return err
			}
		}
		for indexName, indexToAdd := range indexesToAdd {
			if err := db.Run(ctx, createPropertyIndexStatement, map[string]interface{}{
				"name":       indexName,
				"labels":     []string{indexToAdd.kind.String()},
				"properties": []string{indexToAdd.Field},
				"provider":   indexTypeProvider(indexToAdd.Type),
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func assertConstraints(ctx context.Context, db graph.Database, constraintsToRemove []string, constraintsToAdd map[string]neo4jConstraint) error {
	nameMap := map[string]any{}

	for _, constraintToRemove := range constraintsToRemove {
		nameMap["name"] = constraintToRemove

		if err := db.Run(ctx, dropPropertyConstraintStatement, nameMap); err != nil {
			return err
		}
	}

	for constraintName, constraintToAdd := range constraintsToAdd {
		if err := db.Run(ctx, createPropertyConstraintStatement, map[string]interface{}{
			"name":       constraintName,
			"labels":     []string{constraintToAdd.kind.String()},
			"properties": []string{constraintToAdd.Field},
			"provider":   indexTypeProvider(constraintToAdd.Type),
		}); err != nil {
			return err
		}
	}

	return nil
}

func assertGraphSchema(ctx context.Context, db graph.Database, graphSchema graph.Graph, presentIndexes) error {
	var (
		presentIndexes      = allIndexes(presentSchema, presentSchema.Kinds)
		presentConstraints  = allConstraints(presentSchema, presentSchema.Kinds)
		requiredIndexes     = allIndexes(graphSchema, graphSchema.Kinds)
		requiredConstraints = allConstraints(graphSchema, graphSchema.Kinds)

		indexesToRemove     []string
		constraintsToRemove []string
		indexesToAdd        = map[string]neo4jIndex{}
		constraintsToAdd    = map[string]neo4jConstraint{}
	)

	for existingIndexName := range presentIndexes {
		if _, hasMatchingDefinition := requiredIndexes[existingIndexName]; !hasMatchingDefinition {
			indexesToRemove = append(indexesToRemove, existingIndexName)
		}
	}

	for existingConstraintName := range presentConstraints {
		if _, hasMatchingDefinition := requiredConstraints[existingConstraintName]; !hasMatchingDefinition {
			constraintsToRemove = append(constraintsToRemove, existingConstraintName)
		}
	}

	for requiredIndexName, requiredIndex := range requiredIndexes {
		if existingIndex, hasMatchingDefinition := presentIndexes[requiredIndexName]; !hasMatchingDefinition {
			indexesToAdd[requiredIndexName] = requiredIndex
		} else if requiredIndex.Type != existingIndex.Type {
			indexesToRemove = append(indexesToRemove, requiredIndexName)
			indexesToAdd[requiredIndexName] = requiredIndex
		}
	}

	for requiredConstraintName, requiredConstraint := range requiredConstraints {
		if existingConstraint, hasMatchingDefinition := presentConstraints[requiredConstraintName]; !hasMatchingDefinition {
			constraintsToAdd[requiredConstraintName] = requiredConstraint
		} else if requiredConstraint.Type != existingConstraint.Type {
			constraintsToRemove = append(constraintsToRemove, requiredConstraintName)
			constraintsToAdd[requiredConstraintName] = requiredConstraint
		}
	}

	if err := assertConstraints(ctx, db, constraintsToRemove, constraintsToAdd); err != nil {
		return err
	}

	return assertIndexes(ctx, db, indexesToRemove, indexesToAdd)
}

func presentIndexesAndConstraints(ctx context.Context, db graph.Database) ([]graph.Index, []graph.Constraint, error) {
	var (
		indexes     []graph.Index
		constraints []graph.Constraint
	)

	return indexes, constraints, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if result := tx.Run("call db.indexes() yield name, uniqueness, provider, labelsOrTypes, properties;", nil); result.Error() != nil {
			return result.Error()
		} else {
			defer result.Close()

			var (
				name       string
				uniqueness string
				provider   string
				labels     []string
				properties []string
			)

			for result.Next() {
				if err := result.Scan(&name, &uniqueness, &provider, &labels, &properties); err != nil {
					return err
				}

				// Need this for neo4j 4.4+ which creates a weird index by default
				if len(labels) == 0 {
					continue
				}

				if len(labels) > 1 || len(properties) > 1 {
					return fmt.Errorf("composite index types are currently not supported")
				}

				if uniqueness == "UNIQUE" {
					constraints = append(constraints, graph.Constraint{
						Name:  name,
						Field: properties[0],
						Type:  parseProviderType(provider),
					})
				} else {
					indexes = append(indexes, graph.Index{
						Name:  name,
						Field: properties[0],
						Type:  parseProviderType(provider),
					})
				}
			}

			return result.Error()
		}
	})
}

func assertSchema(ctx context.Context, db graph.Database, dbSchema graph.Schema) error {
	if presentIndexes, presentConstraints, err := presentIndexesAndConstraints(ctx, db); err != nil {
		return err
	} else {
		for _, requiredGraph := range requiredSchema.Graphs {

		}
	}
}
