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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type driver struct {
	pool                      *pgxpool.Pool
	schemaManager             *SchemaManager
	defaultTransactionTimeout time.Duration
}

func (s *driver) SetBatchWriteSize(size int) {
}

func (s *driver) SetWriteFlushSize(size int) {
}

func (s *driver) BatchOperation(ctx context.Context, batchDelegate graph.BatchDelegate) error {
	return nil
}

func (s *driver) Close(ctx context.Context) error {
	s.pool.Close()
	return nil
}

var (
	readOnlyTxOptions = pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	}

	readWriteTxOptions = pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
	}
)

func (s *driver) transaction(ctx context.Context, txDelegate graph.TransactionDelegate, pgxOptions pgx.TxOptions, dawgsOptions graph.TransactionConfig) error {
	if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		if tx, err := newTransaction(ctx, conn, pgxOptions, s.schemaManager); err != nil {
			return err
		} else {
			defer tx.Close()

			if err := txDelegate(tx); err != nil {
				return err
			}

			return tx.Commit()
		}
	}
}

func (s *driver) ReadTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	return s.transaction(ctx, txDelegate, readOnlyTxOptions, graph.TransactionConfig{})
}

func (s *driver) WriteTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	return s.transaction(ctx, txDelegate, readWriteTxOptions, graph.TransactionConfig{})
}

func (s *driver) FetchSchema(ctx context.Context) (graph.Schema, error) {
	return graph.Schema{}, nil
}

func (s *driver) AssertSchema(ctx context.Context, graphSchema graph.Schema) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return s.schemaManager.AssertSchema(tx, graphSchema)
	})
}

func (s *driver) Run(ctx context.Context, query string, parameters map[string]any) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		result := tx.Run(query, parameters)
		defer result.Close()

		return result.Error()
	})
}
