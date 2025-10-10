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

import (
	"fmt"
	"math/big"
	"slices"

	"github.com/consensys/go-corset/pkg/util/poly"
)

// Limb is just an alias for Register, but it helps to clarify when we are
// referring to a register after subdivision.
type Limb = Register

// LimbId is just an alias for RegisterId, but it helps to clarify when we are
// referring to a register after subdivision.
type LimbId = RegisterId

// Polynomial defines the type of polynomials over which packets (and register
// splitting in general) operate.
type Polynomial = *poly.ArrayPoly[RegisterId]

// CarryAssignment captures information required to compute the value for a
// given carry line.
type CarryAssignment struct {
	// Register being assigned
	LeftHandSide RegisterId
	// Shift amount applied to result of rhs
	Shift uint
	// Value being calculated
	RightHandSide Polynomial
}

// FieldAgnostic captures the notion of an entity (e.g. module, constraint or
// assignment) which is agnostic to the underlying field being used.  More
// specificially, any registers used within (and constraints, etc) can be
// subdivided as necessary to ensure a maximum bandwidth requirement is met.
// Here, bandwidth refers to the maximum number of data bits which can be stored
// in the underlying field. As a simple example, the prime field F_7 has a
// bandwidth of 2bits.  To target a specific prime field, two parameters are
// used: the maximum bandwidth (as determined by the prime); the maximum
// register width (which should be smaller than the bandwidth).  The maximum
// register width determines the maximum permitted width of any register after
// subdivision.  Since every register value will be stored as a field element,
// it follows that the maximum width cannot be greater than the bandwidth.
// However, in practice, we want it to be marginally less than the bandwidth to
// ensure there is some capacity for calculations involving registers.
type FieldAgnostic[T any] interface {
	// Subdivide for a given bandwidth and maximum register width. This will
	// split all registers wider than the maximum permitted width into two or
	// more "limbs" (i.e. subregisters which do not exceeded the permitted
	// width).  For example, consider a register "r" of width u32. Subdividing
	// this register into registers of at most 8bits will result in four limbs:
	// r'0, r'1, r'2 and r'3 where (by convention) r'0 is the least significant.
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
	Subdivide(RegisterAllocator) T
}

// LimbsMap provides a high-level mapping of all registers across all
// modules before and after subdivision occurs.
type LimbsMap = ModuleMap[RegisterLimbsMap]

// RegisterLimbsMap provides a high-level mapping of all registers before and
// after subdivision occurs within a given module.  That is, it maps a given
// register to those limbs into which it was subdivided.
type RegisterLimbsMap interface {
	RegisterMap
	// Field returns the underlying field configuration used for this mapping.
	// This includes the field bandwidth (i.e. number of bits available in
	// underlying field) and the maximum register width (i.e. width at which
	// registers are capped).
	Field() FieldConfig
	// Limbs identifies the limbs into which a given register is divided.
	// Observe that limbs are ordered by their position in the original
	// register.  In particular, the first limb (i.e. at index 0) is always
	// least significant limb, and the last always most significant.
	LimbIds(RegisterId) []LimbId
	// Limbs returns information about a given limb (i.e. a register which
	// exists after the split).
	Limb(LimbId) Limb
	// Limbs returns all limbs in the mapping.
	Limbs() []Limb
	// LimbsMap returns a register map for the limbs themselves.  This is useful
	// where we need a register map over the limbs, rather than the original
	// registers.
	LimbsMap() RegisterMap
}

// RegisterAllocator extends a register mapping with the ability to allocate new
// registers as necessary.  This is useful, for example,  in the context of
// register splitting for introducing new carry registers.
type RegisterAllocator interface {
	RegisterLimbsMap
	// Allocate a fresh register of the given width within the target module.
	// This is presumed to be a computed register, and automatically assigned a
	// unique name.  Furthermore, an optional
	Allocate(prefix string, width uint) RegisterId
	// Assign a given register the outcome of evaluating a given polynomial,
	// shifted by a given amount.
	Assign(reg RegisterId, shift uint, poly Polynomial)
	// Assignments returns the list of carry assignments
	Assignments() []CarryAssignment
}

// ============================================================================

type registerAllocator struct {
	mapping     RegisterLimbsMap
	assignments []CarryAssignment
	limbs       []Register
}

// NewAllocator converts a mapping into a full allocator simply by wrapping the
// two fields.
func NewAllocator(mapping RegisterLimbsMap) RegisterAllocator {
	limbs := slices.Clone(mapping.Limbs())
	return &registerAllocator{mapping, nil, limbs}
}

// Allocate implementation for the RegisterAllocator interface
func (p *registerAllocator) Allocate(prefix string, width uint) RegisterId {
	var (
		// Determine index for new register
		index = uint(len(p.limbs))
		// Determine unique name for new register
		name = fmt.Sprintf("%s$%d", prefix, index)
		// Default padding (for now)
		zero big.Int
	)
	// Allocate a new computed register.
	p.limbs = append(p.limbs, NewComputedRegister(name, width, zero))
	//
	return NewRegisterId(index)
}

// Assign implementation for the RegisterAllocator interface
func (p *registerAllocator) Assign(target RegisterId, shift uint, poly Polynomial) {
	p.assignments = append(p.assignments, CarryAssignment{target, shift, poly})
}

// Assign implementation for the RegisterAllocator interface
func (p *registerAllocator) Assignments() []CarryAssignment {
	return p.assignments
}

// BandWidth implementation for RegisterMapping interface
func (p *registerAllocator) Field() FieldConfig {
	return p.mapping.Field()
}

// Limbs implementation for the RegisterMapping interface
func (p *registerAllocator) LimbIds(reg RegisterId) []LimbId {
	return p.mapping.LimbIds(reg)
}

// Limb implementation for the RegisterMapping interface
func (p *registerAllocator) Limb(reg LimbId) Limb {
	return p.limbs[reg.Unwrap()]
}

// Limbs implementation for the RegisterMapping interface
func (p *registerAllocator) Limbs() []Limb {
	return p.limbs
}

// LimbsMap implementation for the RegisterMapping interface
func (p *registerAllocator) LimbsMap() RegisterMap {
	return p.mapping.LimbsMap()
}

// Name implementation for RegisterMapping interface
func (p *registerAllocator) Name() string {
	return p.mapping.Name()
}

// HasRegister implementation for RegisterMap interface.
func (p *registerAllocator) HasRegister(name string) (RegisterId, bool) {
	return p.mapping.HasRegister(name)
}

// Register implementation for RegisterMap interface.
func (p *registerAllocator) Register(rid RegisterId) Register {
	return p.mapping.Register(rid)
}

// Registers implementation for RegisterMap interface.
func (p *registerAllocator) Registers() []Register {
	return p.mapping.Registers()
}

func (p *registerAllocator) String() string {
	return p.mapping.String()
}
