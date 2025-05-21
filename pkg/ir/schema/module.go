// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package schema

import "github.com/consensys/go-corset/pkg/util/collection/iter"

// Module represents a "table" within a schema which contains zero or more rows
// for a given set of columns.
type Module interface {
	// Module name
	Name() string

	// Access a given column in this module.
	Column(uint) Column

	// Columns returns an iterator over the underlying columns of this schema.
	// Specifically, the index of a column in this array is its column index.
	Columns() iter.Iterator[Column]

	// Returns the number of columns in this module.
	Width() uint
}

// ============================================================================
//
// ============================================================================

type Table struct {
}

// Module name
func (p *Table) Name() string {
	panic("todo")
}

// Access a given column in this Table.
func (p *Table) Column(uint) Column {
	panic("todo")
}

// Columns returns an iterator over the underlying columns of this schema.
// Specifically, the index of a column in this array is its column index.
func (p *Table) Columns() iter.Iterator[Column] {
	panic("todo")
}

// Returns the number of columns in this Table.
func (p *Table) Width() uint {
	panic("todo")
}

func (p *Table) New(column Column) uint {
	panic("todo")
}
