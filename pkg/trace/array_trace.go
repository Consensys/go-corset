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
	"fmt"
	"math"
	"strings"

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace[T word.Word[T]] struct {
	// Internal memory pool
	pool word.Pool[uint, T]
	// Holds the height of each module in this trace.  The index of each
	// module in this array uniquely identifies it, and is referred to as the
	// "module index".
	modules []ArrayModule[T]
}

// NewArrayTrace constructs a trace from a given set of indexed modules and columns.
func NewArrayTrace[T word.Word[T]](pool word.Pool[uint, T], modules []ArrayModule[T]) *ArrayTrace[T] {
	return &ArrayTrace[T]{pool, modules}
}

// Column accesses a given column directly via a reference.
func (p *ArrayTrace[T]) Column(cref ColumnRef) Column[T] {
	return p.modules[cref.Module()].Column(cref.Column().index)
}

// HasModule determines whether this trace has a module with the given name and,
// if so, what its module index is.
func (p *ArrayTrace[T]) HasModule(module string) (uint, bool) {
	// Linea scan through list of modules
	for mid, mod := range p.modules {
		if mod.name == module {
			return uint(mid), true
		}
	}
	//
	return math.MaxUint, false
}

// Module returns a specific module in this trace.
func (p *ArrayTrace[T]) Module(module uint) Module[T] {
	return p.modules[module]
}

// RawModule returns a specific module in this trace.
func (p *ArrayTrace[T]) RawModule(module uint) *ArrayModule[T] {
	return &p.modules[module]
}

// Modules returns an iterator over the modules in this trace.
func (p *ArrayTrace[T]) Modules() iter.Iterator[Module[T]] {
	arr := iter.NewArrayIterator(p.modules)
	return iter.NewCastIterator[ArrayModule[T], Module[T]](arr)
}

// Width returns number of columns in this trace.
func (p *ArrayTrace[T]) Width() uint {
	return uint(len(p.modules))
}

// Pad prepends (front) and appends (back) a given module with a given number of
// padding rows.
func (p *ArrayTrace[T]) Pad(module uint, front uint, back uint) {
	p.modules[module].Pad(front, back)
}

// Pool implementation for Trace interface.
func (p *ArrayTrace[T]) Pool() word.Pool[uint, T] {
	return p.pool
}

func (p *ArrayTrace[T]) String() string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder

	id.WriteString("{")
	// Write each module in turn
	for i, m := range p.modules {
		if i != 0 {
			id.WriteString(", ")
		}
		//
		id.WriteString(m.String())
	}
	//
	id.WriteString("}")
	//
	return id.String()
}

// ----------------------------------------------------------------------------

// ArrayModule describes an individual module within a trace.
type ArrayModule[T fmt.Stringer] struct {
	// Holds the name of this module
	name string
	// Holds the height of all columns within this module.
	height uint
	// Holds the length multiplier of all columns in this module.  Specifically,
	// the length of all modules must be a multiple of this value.  This then
	// means, for example, that padding must padd in multiples of this to be
	// safe.
	multiplier uint
	// Holds the complete set of columns in this trace.  The index of each
	// column in this array uniquely identifies it, and is referred to as the
	// "column index".
	columns []ArrayColumn[T]
}

// NewArrayModule constructs a module with the given name and an (as yet)
// unspecified height.
func NewArrayModule[T word.Word[T]](name string, multiplier uint, columns []ArrayColumn[T]) ArrayModule[T] {
	var (
		height uint = 0
		first       = true
	)
	//
	for _, c := range columns {
		if first && c.data != nil {
			height = c.Height()
			first = false
		} else if c.data != nil && height != c.Height() {
			// NOTE: we ignore nil columns and assume they are computed columns
			// which are yet to be filled.
			panic(fmt.Sprintf("invalid column height (have %d, expected %d)", c.Height(), height))
		}
	}
	// Sanity check height is a multiple of the length multiplier
	if multiplier == 0 {
		panic(fmt.Sprintf("invalid length multiplier (%d)", multiplier))
	} else if height%multiplier != 0 {
		panic(fmt.Sprintf("invalid module height (have %d, expected multiple of %d)", height, multiplier))
	}
	//
	return ArrayModule[T]{name, height, multiplier, columns}
}

// Name returns the name of this module.
func (p ArrayModule[T]) Name() string {
	return p.name
}

// Column returns a specific column within this trace module.
func (p ArrayModule[T]) Column(id uint) Column[T] {
	return &p.columns[id]
}

// ColumnOf returns a specific column within this trace module.
func (p ArrayModule[T]) ColumnOf(name string) Column[T] {
	for _, c := range p.columns {
		if c.name == name {
			return &c
		}
	}
	//
	panic(fmt.Sprintf("unknown column \"%s\"", name))
}

// Height returns the height of this module, meaning the number of assigned
// rows.
func (p ArrayModule[T]) Height() uint {
	return p.height
}

// Width returns the number of columns in this module.
func (p ArrayModule[T]) Width() uint {
	return uint(len(p.columns))
}

// FillColumn sets the data and padding for the given column.  This will panic
// if the data is already set.  Also, if the module height is updated then this
// returns true to signal a height recalculation is required.
func (p *ArrayModule[T]) FillColumn(cid uint, data array.Array[T], padding T) bool {
	// Find column to fill
	col := &p.columns[cid]
	// Sanity check this column has not already been filled.
	if p.height == math.MaxUint {
		// Initialise column height
		p.height = data.Len()
	} else if data.Len() != p.Height() {
		// The height of this module maybe changing as a result of this
		// operation.  Therefore, we temporarily set it to an invalid value
		// under the expectation at the height will be subsequently recalculated.
		p.height = math.MaxUint
	}
	// Fill the column
	col.data = data
	col.padding = padding
	//
	return p.height == math.MaxUint
}

// Resize the height of this module on the assumption it has been reset whilst
// filling a column (as above).  This will panic if either: the module height
// was not previously reset; or, if the column heights are inconsistent.
func (p *ArrayModule[T]) Resize() {
	var (
		nsize uint = math.MaxUint
		first bool = true
	)
	//
	for i := 0; i != len(p.columns); i++ {
		data := p.columns[i].Data()
		//
		if data == nil {
			// skip
		} else if first {
			nsize = data.Len()
			first = false
		} else if nsize != data.Len() {
			panic(fmt.Sprintf("incompatible column heights (%d vs %d)", nsize, data.Len()))
		}
	}
	// Sanity check height is a multiple of the length multiplier
	if nsize%p.multiplier != 0 {
		panic(fmt.Sprintf("invalid module height (have %d, expected multiple of %d)", nsize, p.multiplier))
	}
	// Done
	p.height = nsize
}

// Pad prepends (front) and appends (back) all columns in this module with a
// given number of padding rows.
func (p *ArrayModule[T]) Pad(front uint, back uint) {
	if front%p.multiplier != 0 {
		panic(fmt.Sprintf("invalid front padding (have %d, expected multiple of %d)", front, p.multiplier))
	} else if back%p.multiplier != 0 {
		panic(fmt.Sprintf("invalid back padding (have %d, expected multiple of %d)", front, p.multiplier))
	}
	// Update height accordingly
	p.height += front + back
	// Padd each column contained within this module.
	for i := 0; i < len(p.columns); i++ {
		p.columns[i].pad(front, back)
	}
}

func (p *ArrayModule[T]) String() string {
	var id strings.Builder
	//
	if p.name == "" {
		id.WriteString("∅")
	} else {
		id.WriteString(p.name)
	}

	id.WriteString("={")
	//
	for i, c := range p.columns {
		if i != 0 {
			id.WriteString(", ")
		}
		//
		id.WriteString(c.String())
	}
	//
	id.WriteString("}")
	// Done
	return id.String()
}

// ----------------------------------------------------------------------------

// ArrayColumn describes an individual column of data within a trace table.
type ArrayColumn[T fmt.Stringer] struct {
	// Holds the name of this column
	name string
	// Holds the raw data making up this column
	data array.Array[T]
	// Value to be used when padding this column
	padding T
}

// NewArrayColumn constructs a with the give name, data and padding.  The given
// data is permitted to be nil, and this is used to signal a computed column.
func NewArrayColumn[T fmt.Stringer](name string, data array.Array[T], padding T) ArrayColumn[T] {
	return ArrayColumn[T]{name, data, padding}
}

// Name returns the name of the given column.
func (p *ArrayColumn[T]) Name() string {
	return p.name
}

// Height determines the height of this column.
func (p *ArrayColumn[T]) Height() uint {
	return p.data.Len()
}

// Padding returns the value which will be used for padding this column.
func (p *ArrayColumn[T]) Padding() T {
	return p.padding
}

// Data provides access to the underlying data of this column
func (p *ArrayColumn[T]) Data() array.Array[T] {
	return p.data
}

// Get the value at a given row in this column.  If the row is
// out-of-bounds, then the column's padding value is returned instead.
// Thus, this function always succeeds.
func (p *ArrayColumn[T]) Get(row int) T {
	if row < 0 || uint(row) >= p.data.Len() {
		// out-of-bounds access
		return p.padding
	}
	// in-bounds access
	return p.data.Get(uint(row))
}

func (p *ArrayColumn[T]) String() string {
	var id strings.Builder
	// Write column name
	id.WriteString(p.name)
	//
	if p.data == nil {
		id.WriteString("=⊥")
	} else {
		id.WriteString("={")
		// Print out each element
		for j := uint(0); j < p.Height(); j++ {
			jth := p.Get(int(j))

			if j != 0 {
				id.WriteString(",")
			}

			id.WriteString(jth.String())
		}
		//
		id.WriteString("}")
	}
	//
	return id.String()
}

func (p *ArrayColumn[T]) pad(front uint, back uint) {
	// if p.data != nil {
	// 	// Pad front of array
	// 	p.data = p.data.Pad(front, back, p.padding)
	// }
	panic("todo")
}
