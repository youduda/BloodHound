// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package harness

import (
	"context"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/query"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDriver_Run(t *testing.T) {
	driver, err := dawgs.Open(pg.DriverName, "user=bhe dbname=bhe password=bhe4eva host=localhost")
	require.Nil(t, err)

	//require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
	//	return query.On(tx).DropSchema()
	//}))

	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		return query.On(tx).CreateSchema()
	}))

	require.Nil(t, driver.AssertSchema(context.Background(), CurrentSchema()))

	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		// Scope to an AD graph
		tx = tx.WithGraph(ActiveDirectoryGraphSchema("ad_graph"))

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
		} else if edge, err := tx.CreateRelationshipByIDs(domainNode.ID, userNode.ID, ad.Contains, graph.NewProperties()); err != nil {
			return err
		} else {
			domainNode.Properties.Set("other_prop", "lol")
			userNode.Properties.Set("is_bad", true)
			edge.Properties.Set("thing", "yes")

			require.Nil(t, tx.UpdateNode(domainNode))
			require.Nil(t, tx.UpdateNode(userNode))
			require.Nil(t, tx.UpdateRelationship(edge))
		}

		return nil
	}))
}
