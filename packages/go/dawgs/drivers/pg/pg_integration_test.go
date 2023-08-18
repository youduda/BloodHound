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

	require.Nil(t, driver.AssertSchema(context.Background(), graph.Schema{
		Kinds: append(ad.Nodes(), ad.Relationships()...),
		Graphs: []graph.Graph{{
			Name: "ad_graph",
		}},
	}))

	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		// Scope to the AD graph
		tx = tx.WithGraph("ad_graph")

		if domainNode, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"name":      "user",
			"objectid":  "12345",
			"domainsid": "12345",
		}), ad.Entity, ad.User); err != nil {
			return err
		} else if userNode, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"name":      "user",
			"objectid":  "12345",
			"domainsid": "12345",
		}), ad.Entity, ad.User); err != nil {
			return err
		} else if _, err := tx.CreateRelationshipByIDs(domainNode.ID, userNode.ID, ad.Contains, nil); err != nil {
			return err
		}

		return nil
	}))
}
