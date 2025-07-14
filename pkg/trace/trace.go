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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bytes"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace interface {
	// Access a given column directly via a reference.
	Column(ColumnRef) Column
	// Access a given module in this trace.
	Module(ModuleId) Module
	// Determine whether this trace has a module with the given name and, if so,
	// what its module index is.
	HasModule(name string) (uint, bool)
	// Returns the number of modules in this trace.
	Width() uint
	// Returns an iterator over the contained modules
	Modules() iter.Iterator[Module]
}

// Module describes a module within the trace.  Every module is composed of some
// number of columns, and has a specific height.
type Module interface {
	// Module name
	Name() string
	// Access a given column in this module.
	Column(uint) Column
	// Access a given column by its name.
	ColumnOf(string) Column
	// Returns the number of columns in this module.
	Width() uint
	// Returns the height of this module.
	Height() uint
}

// Column describes an individual column of data within a trace table.
type Column interface {
	// Holds the name of this column
	Name() string
	// Get the value at a given row in this column.  If the row is
	// out-of-bounds, then the column's padding value is returned instead.
	// Thus, this function always succeeds.
	Get(row int) fr.Element
	// Access the underlying data array for this column.  This is useful in
	// situations where we want to clone the entire column, etc.
	Data() field.FrArray
	// Value to be used when padding this column
	Padding() fr.Element
}

// RawFrColumn is a temporary alias which should be deprecated shortly.
type RawFrColumn = RawColumn[fr.Element]

// BigEndianColumn captures the notion of a raw column holding the bytes of an
// unsigned integer in big endian form.
type BigEndianColumn = RawColumn[bytes.BigEndian]

// RawColumn represents a raw column of data which has not (yet) been indexed as
// part of a trace, etc.  Raw columns are typically read directly from trace
// files, and subsequently indexed into a trace during the expansion process.
type RawColumn[T any] struct {
	// Name of the enclosing module
	Module string
	// Name of the column
	Name string
	// Data held in the column
	Data array.Array[T]
}

// QualifiedName returns the fully qualified name of this column.
func (p *RawColumn[T]) QualifiedName() string {
	return QualifiedColumnName(p.Module, p.Name)
}
