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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace struct {
	// Holds the height of each module in this trace.  The index of each
	// module in this array uniquely identifies it, and is referred to as the
	// "module index".
	modules []ArrayModule
}

// NewArrayTrace constructs a trace from a given set of indexed modules and columns.
func NewArrayTrace(modules []ArrayModule) *ArrayTrace {
	return &ArrayTrace{modules}
}

// Module returns a specific module in this trace.
func (p *ArrayTrace) Module(module uint) Module {
	return p.modules[module]
}

// Modules returns an iterator over the modules in this trace.
func (p *ArrayTrace) Modules() iter.Iterator[Module] {
	arr := iter.NewArrayIterator(p.modules)
	return iter.NewCastIterator[ArrayModule, Module](arr)
}

// Width returns number of columns in this trace.
func (p *ArrayTrace) Width() uint {
	return uint(len(p.modules))
}

// Height returns the height of a given context (i.e. module) in the trace.
func (p *ArrayTrace) Height(ctx Context) uint {
	return p.modules[ctx.Module()].height * ctx.Multiplier
}

// Pad prepends (front) and appends (back) a given module with a given number of
// padding rows.
func (p *ArrayTrace) Pad(module uint, front uint, back uint) {
	p.modules[module].Pad(front, back)
}

func (p *ArrayTrace) String() string {
	panic("todo")
}

// ----------------------------------------------------------------------------

// ArrayModule describes an individual module within a trace.
type ArrayModule struct {
	// Holds the name of this module
	name string
	// Holds the height of all columns within this module.
	height uint
	// Holds the complete set of columns in this trace.  The index of each
	// column in this array uniquely identifies it, and is referred to as the
	// "column index".
	columns []ArrayColumn
}

// NewArrayModule constructs a module with the given name and an (as yet)
// unspecified height.
func NewArrayModule(name string, columns []ArrayColumn) ArrayModule {
	var height uint = math.MaxUint
	//
	for i, c := range columns {
		if i == 0 {
			height = c.Height()
		} else if height != c.Height() {
			panic("invalid column height")
		}
	}
	//
	return ArrayModule{name, height, columns}
}

// Name returns the name of this module.
func (p ArrayModule) Name() string {
	return p.name
}

// Column returns a specific column within this trace module.
func (p ArrayModule) Column(id uint) Column {
	return &p.columns[id]
}

// Height returns the height of this module, meaning the number of assigned
// rows.
func (p ArrayModule) Height() uint {
	return p.height
}

// Width returns the number of columns in this module.
func (p ArrayModule) Width() uint {
	return uint(len(p.columns))
}

// Pad prepends (front) and appends (back) all columns in this module with a
// given number of padding rows.
func (p *ArrayModule) Pad(front uint, back uint) {
	p.height += front + back
	// Padd each column contained within this module.
	for i := 0; i < len(p.columns); i++ {
		p.columns[i].pad(front, back)
	}
}

// ----------------------------------------------------------------------------

// ArrayColumn describes an individual column of data within a trace table.
type ArrayColumn struct {
	// Evaluation context of this column
	context Context
	// Holds the name of this column
	name string
	// Holds the raw data making up this column
	data field.FrArray
	// Value to be used when padding this column
	padding fr.Element
}

// NewArrayColumn constructs a  with the give name, data and padding.
func NewArrayColumn(context Context, name string, data field.FrArray,
	padding fr.Element) ArrayColumn {
	col := EmptyArrayColumn(context, name)
	col.fill(data, padding)
	//
	return col
}

// EmptyArrayColumn constructs a  with the give name, data and padding.
func EmptyArrayColumn(context Context, name string) ArrayColumn {
	return ArrayColumn{context, name, nil, fr.NewElement(0)}
}

// Context returns the evaluation context this column provides.
func (p *ArrayColumn) Context() Context {
	return p.context
}

// Name returns the name of the given column.
func (p *ArrayColumn) Name() string {
	return p.name
}

// Height determines the height of this column.
func (p *ArrayColumn) Height() uint {
	return p.data.Len()
}

// Padding returns the value which will be used for padding this column.
func (p *ArrayColumn) Padding() fr.Element {
	return p.padding
}

// Data provides access to the underlying data of this column
func (p *ArrayColumn) Data() field.FrArray {
	return p.data
}

// Get the value at a given row in this column.  If the row is
// out-of-bounds, then the column's padding value is returned instead.
// Thus, this function always succeeds.
func (p *ArrayColumn) Get(row int) fr.Element {
	if row < 0 || uint(row) >= p.data.Len() {
		// out-of-bounds access
		return p.padding
	}
	// in-bounds access
	return p.data.Get(uint(row))
}

func (p *ArrayColumn) fill(data field.FrArray, padding fr.Element) {
	// Sanity check this column has not already been filled.
	if p.data != nil {
		panic(fmt.Sprintf("computed column %s has already been filled", p.name))
	} else if data.Len()%p.context.LengthMultiplier() != 0 {
		panic(fmt.Sprintf("computed column %s filling has invalid length multiplier", p.name))
	}
	// Fill the column
	p.data = data
	p.padding = padding
}

func (p *ArrayColumn) pad(front uint, back uint) {
	if p.data != nil {
		// Apply the length multiplier
		front = front * p.context.LengthMultiplier()
		back = back * p.context.LengthMultiplier()
		// Pad front of array
		p.data = p.data.Pad(front, back, p.padding)
	}
}
