package model

import (
	_ "embed"
	"github.com/specterops/bloodhound/dawgs/graph"
	"strconv"
	"strings"
)

var (
	// go:embed sql/schema_up.sql
	pgSchemaUp string

	// go:embed sql/schema_down.sql
	pgSchemaDown string

	// go:embed sql/select_table_indexes.sql
	pgSelectTableIndexes string
)

func GraphSchemaSQLUp() string {
	return pgSchemaUp
}

func GraphSchemaSQLDown() string {
	return pgSchemaDown
}

func formatPartitionTableName(parent string, graphID int32) string {
	return parent + "_" + strconv.FormatInt(int64(graphID), 10)
}

func formatPartitionTableSQL(parent string, graphID int32) string {
	var (
		graphIDStr = strconv.FormatInt(int64(graphID), 10)
		builder    = strings.Builder{}
	)

	builder.WriteString("create table ")
	builder.WriteString(parent)
	builder.WriteString("_")
	builder.WriteString(graphIDStr)
	builder.WriteString(" partition of ")
	builder.WriteString(parent)
	builder.WriteString(" for values in (")
	builder.WriteString(graphIDStr)
	builder.WriteString(")")

	return builder.String()
}

func formatIndexName(partitionName string, index graph.Index) string {
	stringBuilder := strings.Builder{}

	stringBuilder.WriteString(partitionName)
	stringBuilder.WriteString("_")
	stringBuilder.WriteString(index.Field)
	stringBuilder.WriteString("_index")

	return stringBuilder.String()
}

func formatConstraintName(partitionName string, constraint graph.Constraint) string {
	stringBuilder := strings.Builder{}

	stringBuilder.WriteString(partitionName)
	stringBuilder.WriteString("_")
	stringBuilder.WriteString(constraint.Field)
	stringBuilder.WriteString("_constraint")

	return stringBuilder.String()
}
