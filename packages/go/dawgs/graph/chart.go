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

package graph

type IndexType int

const (
	BTreeIndex    IndexType = 1
	FullTextIndex IndexType = 2
)

func (s IndexType) String() string {
	switch s {
	case BTreeIndex:
		return "btree"

	case FullTextIndex:
		return "fts"

	default:
		return "invalid"
	}
}

type Constraint struct {
	Field     string
	IndexType IndexType
}

func (s Constraint) Name() string {
	return s.Field + "_" + s.IndexType.String() + "_constraint"
}

type Index struct {
	Field string
	Type  IndexType
}

func (s Index) Name() string {
	return s.Field + "_" + s.Type.String() + "_index"
}

type Graph struct {
	Name        string
	Constraints []Constraint
	Indexes     []Index
}

type Schema struct {
	Kinds  []Kind
	Graphs []Graph
}
