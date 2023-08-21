package model

import "regexp"

var (
	pgIndexRegex = regexp.MustCompile(`(?i)^CREATE (UNIQUE)? ?INDEX ([^ ]+).+USING ([^ ]+) \(([^)]+)\)$`)
)

const (
	pgIndexRegexGroupUnique       = 1
	pgIndexRegexGroupName         = 2
	pgIndexRegexGroupIndexType    = 3
	pgIndexRegexGroupFields       = 4
	pgIndexRegexNumExpectedGroups = 5

	pgIndexTypeBTree   = "btree"
	pgIndexTypeGIN     = "gin"
	pgIndexUniqueStr   = "unique"
	pgPropertiesColumn = "properties"

	pgNodeTableName = "node"
	pgEdgeTableName = "edge"
)
