package pg

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_formatPartitionTableSQL(t *testing.T) {
	require.Equal(t, "create table node_1 partition of node for values in (1)", formatPartitionTableSQL(nodeTableName, 1))
	require.Equal(t, "create table edge_1 partition of edge for values in (1)", formatPartitionTableSQL(edgeTableName, 1))
}
