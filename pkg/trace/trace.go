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
package trace

import (
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace[F any] interface {
	// Provides agnostic mechanism for constructing arrays
	Builder() array.Builder[F]
	// Access a given column difrtly via a reference.
	Column(ColumnRef) Column[F]
	// Determine whether this trace has a module with the given name and, if so,
	// what its module index is.
	HasModule(name string) (uint, bool)
	// Access a given module in this trace.
	Module(ModuleId) Module[F]
	// Returns an iterator over the contained modules
	Modules() iter.Iterator[Module[F]]
	// Returns the number of modules in this trace.
	Width() uint
}

// Module describes a module within the trace.  Every module is composed of some
// number of columns, and has a specific height.
type Module[T any] interface {
	// Module name
	Name() string
	// Access a given column in this module.
	Column(uint) Column[T]
	// Access a given column by its name.
	ColumnOf(string) Column[T]
	// Returns the number of columns in this module.
	Width() uint
	// Returns the height of this module.
	Height() uint
}

// Column describes an individual column of data within a trace table.
type Column[T any] interface {
	// Holds the name of this column
	Name() string
	// Get the value at a given row in this column.  If the row is
	// out-of-bounds, then the column's padding value is returned instead.
	// Thus, this function always succeeds.
	Get(row int) T
	// Access the underlying data array for this column.  This is useful in
	// situations where we want to clone the entire column, etc.
	Data() array.Array[T]
	// Padding returns the value which will be used for padding this column.
	Padding() T
}
