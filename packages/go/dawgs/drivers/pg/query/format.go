package query

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"strconv"
	"strings"
)

func postgresIndexType(indexType graph.IndexType) string {
	switch indexType {
	case graph.BTreeIndex:
		return pgIndexTypeBTree
	case graph.TextSearchIndex:
		return pgIndexTypeGIN
	default:
		return "NOT SUPPORTED"
	}
}

func parsePostgresIndexType(pgType string) graph.IndexType {
	switch strings.ToLower(pgType) {
	case pgIndexTypeBTree:
		return graph.BTreeIndex
	case pgIndexTypeGIN:
		return graph.TextSearchIndex
	default:
		return graph.UnsupportedIndex
	}
}

func join(values ...string) string {
	builder := strings.Builder{}

	for idx := 0; idx < len(values); idx++ {
		builder.WriteString(values[idx])
	}

	return builder.String()
}

func formatDropPropertyIndex(indexName string) string {
	return join("drop index if exists ", indexName, ";")
}

func formatCreatePropertyIndex(indexName, tableName, fieldName string, indexType graph.IndexType) string {
	var (
		pgIndexType  = postgresIndexType(indexType)
		queryPartial = join("create index ", indexName, " on ", tableName, " using ",
			pgIndexType, " ((", tableName, ".", pgPropertiesColumn, " ->> '", fieldName)
	)

	if indexType == graph.TextSearchIndex {
		// GIN text search requires the column to be typed and to contain the tri-gram operation extension
		return join(queryPartial, "::text'::text) gin_trgm_ops);")
	} else {
		return join(queryPartial, "'));")
	}
}

func formatCreatePartitionTable(name, parent string, graphID int32) string {
	builder := strings.Builder{}

	builder.WriteString("create table ")
	builder.WriteString(name)
	builder.WriteString(" partition of ")
	builder.WriteString(parent)
	builder.WriteString(" for values in (")
	builder.WriteString(strconv.FormatInt(int64(graphID), 10))
	builder.WriteString(")")

	return builder.String()
}
