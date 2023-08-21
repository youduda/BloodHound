package model

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"strings"
)

func fetchGraphPartition(tx graph.Transaction, partitionTableName string) (GraphPartition, error) {
	var (
		indexName       string
		indexDefinition string
		graphPartition  = NewGraphPartition()
		result          = tx.Run(pgSelectTableIndexes, map[string]any{
			"tablename": partitionTableName,
		})
	)

	defer result.Close()

	for result.Next() {
		if err := result.Scan(&indexDefinition); err != nil {
			return graphPartition, err
		}

		if captureGroups := pgIndexRegex.FindStringSubmatch(indexDefinition); captureGroups == nil {
			log.Infof("Potential regex mis-match with PG schema definition for index. Source definition: %s", indexDefinition)
		} else if indexFields := captureGroups[pgIndexRegexGroupFields]; strings.Contains(indexFields, pgPropertiesColumn) {
			indexName = captureGroups[pgIndexRegexGroupName]

			if captureGroups[pgIndexRegexGroupUnique] == pgIndexUniqueStr {
				graphPartition.Constraints[indexName] = graph.Constraint{
					Name:  indexName,
					Field: indexFields,
					Type:  pgParseIndexType(captureGroups[pgIndexRegexGroupIndexType]),
				}
			} else {
				graphPartition.Indexes[indexName] = graph.Index{
					Name:  indexName,
					Field: indexFields,
					Type:  pgParseIndexType(captureGroups[pgIndexRegexGroupIndexType]),
				}
			}
		}
	}

	return graphPartition, result.Error()
}

func assertGraphPartitionIndexes(tx graph.Transaction, partitionTableName string, indexChanges IndexChangeSet) error {
	var (
		builder      = strings.Builder{}
		runQueryFunc = func() error {
			var (
				result = tx.Run(builder.String(), nil)
				err    = result.Error()
			)

			result.Close()
			builder.Reset()

			return err
		}
	)

	for _, indexToRemove := range indexChanges.NodeIndexesToRemove {
		builder.WriteString(`drop index if exists `)
		builder.WriteString(indexToRemove)

		if err := runQueryFunc(); err != nil {
			return err
		}
	}

	for indexName, index := range indexChanges.NodeIndexesToAdd {
		builder.WriteString("create index ")
		builder.WriteString(indexName)
		builder.WriteString(" on ")
		builder.WriteString(partitionTableName)
		builder.WriteString(" using ")
		builder.WriteString(pgIndexTypeToString(index.Type))
		builder.WriteString(" ((")
		builder.WriteString(partitionTableName)
		builder.WriteString(".properties->>'")
		builder.WriteString(index.Field)

		// TODO: Inform schema of column type?

		if index.Type == graph.TextSearchIndex {
			// GIN text search requires the column to be typed and to contain the tri-gram operation extension
			builder.WriteString("::text') gin_trgm_ops")
		} else {
			builder.WriteString("')")
		}

		builder.WriteString(")")

		if err := runQueryFunc(); err != nil {
			return err
		}
	}

	return nil
}

func assertGraphPartitions(tx graph.Transaction, graphSchema graph.Graph, graphDefinition Graph) error {
	var (
		requiredNodePartition = NewGraphPartitionFromSchema(graphDefinition.NodePartition.Name, graphSchema.NodeIndexes, graphSchema.NodeConstraints)
		indexChangeSet        = NewIndexChangeSet()
	)

	if presentNodePartition, err := fetchGraphPartition(tx, graphDefinition.NodePartition.Name); err != nil {
		return err
	} else {
		for presentNodeIndexName := range presentNodePartition.Indexes {
			if _, hasMatchingDefinition := requiredNodePartition.Indexes[presentNodeIndexName]; !hasMatchingDefinition {
				indexChangeSet.NodeIndexesToRemove = append(indexChangeSet.NodeIndexesToRemove, presentNodeIndexName)
			}
		}

		for presentNodeConstraintName := range presentNodePartition.Constraints {
			if _, hasMatchingDefinition := requiredNodePartition.Constraints[presentNodeConstraintName]; !hasMatchingDefinition {
				indexChangeSet.NodeConstraintsToRemove = append(indexChangeSet.NodeConstraintsToRemove, presentNodeConstraintName)
			}
		}

		for requiredNodeIndexName, requiredNodeIndex := range requiredNodePartition.Indexes {
			if presentNodeIndex, hasMatchingDefinition := presentNodePartition.Indexes[requiredNodeIndexName]; !hasMatchingDefinition {
				indexChangeSet.NodeIndexesToAdd[requiredNodeIndexName] = requiredNodeIndex
			} else if requiredNodeIndex.Type != presentNodeIndex.Type {
				indexChangeSet.NodeIndexesToRemove = append(indexChangeSet.NodeIndexesToRemove, requiredNodeIndexName)
				indexChangeSet.NodeIndexesToAdd[requiredNodeIndexName] = requiredNodeIndex
			}
		}

		for requiredNodeConstraintName, requiredNodeConstraint := range requiredNodePartition.Constraints {
			if presentNodeConstraint, hasMatchingDefinition := presentNodePartition.Constraints[requiredNodeConstraintName]; !hasMatchingDefinition {
				indexChangeSet.NodeConstraintsToAdd[requiredNodeConstraintName] = requiredNodeConstraint
			} else if requiredNodeConstraint.Type != presentNodeConstraint.Type {
				indexChangeSet.NodeConstraintsToRemove = append(indexChangeSet.NodeConstraintsToRemove, requiredNodeConstraintName)
				indexChangeSet.NodeConstraintsToAdd[requiredNodeConstraintName] = requiredNodeConstraint
			}
		}
	}

	return assertGraphPartitionIndexes(tx, graphDefinition.NodePartition.Name, indexChangeSet)
}
