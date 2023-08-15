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

	schema := graph.NewSchema()
	schema.DefineKinds(ad.NodeKinds()...)
	schema.DefineKinds(ad.Relationships()...)

	require.Nil(t, driver.AssertSchema(context.Background(), schema))
	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		_, err := tx.CreateNode(nil, ad.Entity, ad.User)
		return err
	}))
}
