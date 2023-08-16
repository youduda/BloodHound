package pg_test

import (
	"context"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDriver_Run(t *testing.T) {
	driver, err := dawgs.Open(pg.DriverName, "user=bhe dbname=bhe password=bhe4eva host=localhost")
	require.Nil(t, err)

	require.Nil(t, pg.InitSchemaDown(context.Background(), driver))
	require.Nil(t, pg.InitSchemaUp(context.Background(), driver))

	schema := graph.NewDatabaseSchema()
	adGraph := schema.Graph("ad_graph")

	adGraph.DefineKinds(ad.NodeKinds()...)
	adGraph.DefineKinds(ad.Relationships()...)

	require.Nil(t, driver.AssertSchema(context.Background(), schema))
	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		// Scope to the AD graph
		tx = tx.WithGraph("ad_graph")

		_, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"name":      "user",
			"domainsid": "12345",
		}), ad.Entity, ad.User)

		return err
	}))
}
