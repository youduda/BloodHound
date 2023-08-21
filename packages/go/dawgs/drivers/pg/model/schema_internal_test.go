package model

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_PGIndexRegex(t *testing.T) {
	captureGroups := pgIndexRegex.FindStringSubmatch("CREATE INDEX edge_1_kind_id_idx ON public.edge_1 USING btree (kind_id)")

	require.Equal(t, pgIndexRegexNumExpectedGroups, len(captureGroups))
	require.Equal(t, "", captureGroups[pgIndexRegexGroupUnique])
	require.Equal(t, "edge_1_kind_id_idx", captureGroups[pgIndexRegexGroupName])
	require.Equal(t, "btree", captureGroups[pgIndexRegexGroupIndexType])
	require.Equal(t, "kind_id", captureGroups[pgIndexRegexGroupFields])

	captureGroups = pgIndexRegex.FindStringSubmatch("create UNIQUE index edge_1_unique_col_constraint ON public.edge_1 USING btree (unique_col)")

	require.Equal(t, pgIndexRegexNumExpectedGroups, len(captureGroups))
	require.Equal(t, "UNIQUE", captureGroups[pgIndexRegexGroupUnique])
	require.Equal(t, "edge_1_unique_col_constraint", captureGroups[pgIndexRegexGroupName])
	require.Equal(t, "btree", captureGroups[pgIndexRegexGroupIndexType])
	require.Equal(t, "unique_col", captureGroups[pgIndexRegexGroupFields])
}

func Test_formatPartitionTableSQL(t *testing.T) {
	require.Equal(t, "create table node_1 partition of node for values in (1)", formatPartitionTableSQL(pgNodeTableName, 1))
	require.Equal(t, "create table edge_1 partition of edge for values in (1)", formatPartitionTableSQL(pgEdgeTableName, 1))
}
