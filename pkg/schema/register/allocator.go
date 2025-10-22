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
type Allocator interface {
	Map
	// Allocate a fresh register of the given width within the target module.
	// This is presumed to be a computed register, and automatically assigned a
	// unique name.  Furthermore, an optional
	Allocate(prefix string, width uint) Id
	// Assign a given register the outcome of evaluating a given polynomial,
	// shifted by a given amount.
	Assign(reg Id, shift uint, poly RelativePolynomial)
	// Assignments returns the list of carry assignments
	Assignments() []CarryAssignment
	// Reset back to a given number of registers.  This is essentially for
	// "undoing" allocations in algorithms that perform speculative allocation.
	Reset(uint)
}

// ============================================================================

type registerAllocator struct {
	mapping     Map
	assignments []CarryAssignment
	registers   []Register
}

// NewAllocator converts a mapping into a full allocator simply by wrapping the
// two fields.
func NewAllocator(mapping Map) Allocator {
	registers := slices.Clone(mapping.Registers())
	return &registerAllocator{mapping, nil, registers}
}

// Allocate implementation for the RegisterAllocator interface
func (p *registerAllocator) Allocate(prefix string, width uint) Id {
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
	//
	return NewId(index)
}

// Assign implementation for the RegisterAllocator interface
func (p *registerAllocator) Assign(target Id, shift uint, poly RelativePolynomial) {
	p.assignments = append(p.assignments, CarryAssignment{target, shift, poly})
}

// Assign implementation for the RegisterAllocator interface
func (p *registerAllocator) Assignments() []CarryAssignment {
	return p.assignments
}

// Name implementation for RegisterMapping interface
func (p *registerAllocator) Name() string {
	return p.mapping.Name()
}

// HasRegister implementation for RegisterMap interface.
func (p *registerAllocator) HasRegister(name string) (Id, bool) {
	for i, reg := range p.registers {
		if reg.Name == name {
			return NewId(uint(i)), true
		}
	}
	//
	return UnusedId(), false
}

// Register implementation for RegisterMap interface.
func (p *registerAllocator) Register(rid Id) Register {
	return p.registers[rid.Unwrap()]
}

// Registers implementation for RegisterMap interface.
func (p *registerAllocator) Registers() []Register {
	return p.registers
}

// Reset implementation for RegisterAllocator interface.
func (p *registerAllocator) Reset(n uint) {
	p.registers = p.registers[:n]
}

func (p *registerAllocator) String() string {
	return MapToString(p)
}
