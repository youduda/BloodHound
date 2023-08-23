package query

import (
	"embed"
	"fmt"
	"path"
)

var (
	//go:embed sql
	queryFS embed.FS
)

func readFile(name string) string {
	content, err := queryFS.ReadFile(name)

	if err != nil {
		panic(fmt.Sprintf("Unable to find embedded query file %s: %v", name, err))
	}

	return string(content)
}

func loadSQL(name string) string {
	return readFile(path.Join("sql", name))
}

var (
	sqlSchemaUp           = loadSQL("schema_up.sql")
	sqlSchemaDown         = loadSQL("schema_down.sql")
	sqlSelectTableIndexes = loadSQL("select_table_indexes.sql")
	sqlSelectKindID       = loadSQL("select_table_indexes.sql")
	sqlSelectGraphs       = loadSQL("select_graphs.sql")
	sqlInsertGraph        = loadSQL("insert_graph.sql")
	sqlInsertKind         = loadSQL("insert_kind.sql")
	sqlSelectKinds        = loadSQL("select_kinds.sql")
	sqlSelectGraphByName  = loadSQL("select_graph_by_name.sql")
)
