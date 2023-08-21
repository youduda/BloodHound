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

package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"time"
)

const (
	DriverName = "pg"

	poolInitConnectionTimeout = time.Second * 10
	defaultTransactionTimeout = time.Minute * 15
)

func newDatabase(connectionString string) (graph.Database, error) {
	poolCtx, done := context.WithTimeout(context.Background(), poolInitConnectionTimeout)
	defer done()

	if poolCfg, err := pgxpool.ParseConfig(connectionString); err != nil {
		return nil, err
	} else if pool, err := pgxpool.NewWithConfig(poolCtx, poolCfg); err != nil {
		return nil, err
	} else {
		return &driver{
			pool:                      pool,
			schemaManager:             model.NewSchemaManager(),
			defaultTransactionTimeout: defaultTransactionTimeout,
		}, nil
	}
}

func init() {
	dawgs.Register(DriverName, func(cfg any) (graph.Database, error) {
		if connectionString, typeOK := cfg.(string); !typeOK {
			return nil, fmt.Errorf("expected string for configuration type but got %T", cfg)
		} else {
			return newDatabase(connectionString)
		}
	})
}
