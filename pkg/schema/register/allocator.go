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
package register

import (
	"fmt"
	"math/big"
	"slices"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// CarryAssignment captures information required to compute the value for a
// given carry line.
type CarryAssignment struct {
	// Register being assigned
	LeftHandSide Id
	// Shift amount applied to result of rhs
	Shift uint
	// Value being calculated
	RightHandSide RelativePolynomial
}

// Polynomial defines the type of polynomials over which packets (and register
// splitting in general) operate.
type Polynomial[T util.Comparable[T]] = *poly.ArrayPoly[T]

// RelativePolynomial represents a polynomial over "relative registers".  That
// is, it can refer to registers on the current row or on a row relative to the
// current row (e.g. the next row, or the previous row, etc).
type RelativePolynomial = Polynomial[RelativeId]

// Allocator extends a register mapping with the ability to allocate new
// registers as necessary.  This is useful, for example,  in the context of
// register splitting for introducing new carry registers.
type Allocator[T any] interface {
	Map
	// Allocate a fresh register of the given width within the target module.
	// This is presumed to be a computed register, and automatically assigned a
	// unique name.  Furthermore, optional metadata can for the column.
	Allocate(prefix string, width uint, metadata util.Option[T]) Id
	// Assignments returns any metadata assigned to an allocated register.
	Assignments() []util.Pair[Id, T]
	// Reset back to a given number of registers.  This is essentially for
	// "undoing" allocations in algorithms that perform speculative allocation.
	Reset(uint)
}

// ============================================================================

type registerAllocator[T any] struct {
	mapping     Map
	assignments []util.Option[T]
	registers   []Register
}

// NewAllocator converts a mapping into a full allocator simply by wrapping the
// two fields.
func NewAllocator[T any](mapping Map) Allocator[T] {
	var (
		registers   = slices.Clone(mapping.Registers())
		assignments = make([]util.Option[T], len(registers))
	)
	//
	return &registerAllocator[T]{mapping, assignments, registers}
}

// Allocate implementation for the RegisterAllocator interface
func (p *registerAllocator[T]) Allocate(prefix string, width uint, filler util.Option[T]) Id {
	var (
		// Determine index for new register
		index = uint(len(p.registers))
		// Determine unique name for new register
		name = fmt.Sprintf("%s$%d", prefix, index)
		// Default padding (for now)
		zero big.Int
	)
	// Allocate a new computed register.
	p.registers = append(p.registers, NewComputed(name, width, zero))
	// record assignment
	p.assignments = append(p.assignments, filler)
	//
	return NewId(index)
}

// Assign implementation for the RegisterAllocator interface
func (p *registerAllocator[T]) Assignments() []util.Pair[Id, T] {
	var assignments []util.Pair[Id, T]
	//
	for i, data := range p.assignments {
		if data.HasValue() {
			// Construct assignment
			assignment := util.NewPair(NewId(uint(i)), data.Unwrap())
			// Include it
			assignments = append(assignments, assignment)
		}
	}
	//
	return assignments
}

// Name implementation for RegisterMapping interface
func (p *registerAllocator[T]) Name() string {
	return p.mapping.Name()
}

// HasRegister implementation for RegisterMap interface.
func (p *registerAllocator[T]) HasRegister(name string) (Id, bool) {
	for i, reg := range p.registers {
		if reg.Name == name {
			return NewId(uint(i)), true
		}
	}
	//
	return UnusedId(), false
}

// Register implementation for RegisterMap interface.
func (p *registerAllocator[T]) Register(rid Id) Register {
	return p.registers[rid.Unwrap()]
}

// Registers implementation for RegisterMap interface.
func (p *registerAllocator[T]) Registers() []Register {
	return p.registers
}

// Reset implementation for RegisterAllocator interface.
func (p *registerAllocator[T]) Reset(n uint) {
	if n < uint(len(p.mapping.Registers())) {
		panic("cannot reset pre-existing registers")
	}
	// Reset registers
	p.registers = p.registers[:n]
	// Reset metadata
	p.assignments = p.assignments[:n]
}

func (p *registerAllocator[T]) String() string {
	return MapToString(p)
}
