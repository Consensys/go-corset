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

// Module represents a "table" within a schema which contains zero or more rows
// for a given set of columns.
type Module interface {
	// Module name
	Name() string

	// Access a given column in this module.
	Column(uint) Column

	// Columns providers access to the underlying columns of this schema.
	Columns() []Column

	// Returns the number of columns in this module.
	Width() uint
}

// FieldAgnosticModule captures the notion of a module which is agnostic to the
// underlying field being used.  More specificially, it is a module whose
// registers (and constraints) can be subdivided as necessary to ensure a
// maximum bandwidth requirement is met.  Here, bandwidth refers to the maximum
// number of data bits which can be stored in the underlying field.  As a simple
// example, the prime field F_7 has a bandwidth of 2bits.  To target a specific
// prime field, two parameters are used: the maximum bandwidth (as determined by
// the prime); the maximum register width (which should be smaller than the
// bandwidth).  The maximum register width determines the maximum permitted
// width of any register in the module.  Since every register value will be
// stored as a field element, it follows that the maximum width cannot be
// greater than the bandwidth.  However, in practice, we want it to be
// marginally less than the bandwidth to ensure there is some capacity for
// calculations involving registers.
type FieldAgnosticModule[T any] interface {
	Module
	// Subdivide this module for a given bandwidth and maximum register width.
	// This will split all registers wider than the maximum permitted width into
	// two or more "limbs" (i.e. subregisters which do not exceeded the
	// permitted width).  For example, consider a register "r" of width u32.
	// Subdividing this register into registers of at most 8bits will result in
	// four limbs: r'0, r'1, r'2 and r'3 where (by convention) r'0 is the least
	// significant.
	//
	// As part of the subdivision process, constraints may also need to be
	// divided when they exceed the maximum permitted bandwidth.  For example,
	// consider a simple constraint such as "x = y + 1" using 16bit registers
	// x,y.  Subdividing for a bandwidth of 10bits and a maximum register width
	// of 8bits means splitting each register into two limbs, and transforming
	// our constraint into:
	//
	// 256*x'1 + x'0 = 256*y'1 + y'0 + 1
	//
	// However, as it stands, this constraint exceeds our bandwidth requirement
	// since it requires at least 17bits of information to safely evaluate each
	// side.  Thus, the constraint itself must be subdivided into two parts:
	//
	// 256*c + x'0 = y'0 + 1  // lower
	//
	//         x'1 = y'1 + c  // upper
	//
	// Here, c is a 1bit register introduced as part of the transformation to
	// act as a "carry" between the two constraints.
	Subdivide(bandwidth uint, maxRegisterWidth uint) T
}

// ============================================================================
//
// ============================================================================

// Table provides a straightforward, reusable module implementation.  There is
// nothing fancy here: we simply have a set of registers, constraints and
// assignments.  A table is a field agnostic module with a simple strategy of
// subdividing registers "in place".  For example, suppose we have registers X
// and Y (in that order) where both are to be halfed.  Then, the result is X'0,
// X'1, Y'0. Y'1 (in that order).  Hence, predicting the new register indices is
// relatively straightforward.
type Table[C Constraint] struct {
	name        string
	registers   []Column
	constraints []C
}

// NewTable constructs an empty table module.
func NewTable[C Constraint](name string, registers []Column, constraints []C) Table[C] {
	return Table[C]{name, registers, constraints}
}

// Module name
func (p Table[C]) Name() string {
	return p.name
}

// Access a given column in this Table.
func (p Table[C]) Column(index uint) Column {
	return p.registers[index]
}

// Columns returns an iterator over the underlying columns of this schema.
// Specifically, the index of a column in this array is its column index.
func (p Table[C]) Columns() []Column {
	panic("todo")
}

// Returns the number of columns in this Table.
func (p Table[C]) Width() uint {
	return uint(len(p.registers))
}

func (p *Table[C]) New(column Column) uint {
	panic("todo")
}
