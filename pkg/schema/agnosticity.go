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
	"slices"
)

// Limb is just an alias for Register, but it helps to clarify when we are
// referring to a register after subdivision.
type Limb = Register

// LimbId is just an alias for RegisterId, but it helps to clarify when we are
// referring to a register after subdivision.
type LimbId = RegisterId

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
	Subdivide(RegisterMappings) T
}

// RegisterMappings provides a high-level mapping of all registers before and after
// subdivision occurs.
type RegisterMappings interface {
	// Field returns the underlying field configuration used for this mapping.
	// This includes the field bandwidth (i.e. number of bits available in
	// underlying field) and the maximum register width (i.e. width at which
	// registers are capped).
	Field() FieldConfig
	// Module returns register mapping information for the given module.
	Module(ModuleId) RegisterMapping
	// ModuleOf returns register mapping information for the given module.
	ModuleOf(string) RegisterMapping
}

// RegisterMapping provides a high-level mapping of all registers before and
// after subdivision occurs in a given module.
type RegisterMapping interface {
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
	// RegisterOf determines a register's ID based on its name.
	RegisterOf(string) RegisterId
}

// RegisterAllocator extends a register mapping with the ability to allocate new
// registers as necessary.  This is useful, for example,  in the context of
// register splitting for introducing new carry registers.
type RegisterAllocator interface {
	RegisterMapping
	// AllocateCarry a fresh register of the given width within the target module.
	// This is presumed to be a computed register, and automatically assigned a
	// unique name.
	AllocateCarry(prefix string, width uint) RegisterId
}

// ============================================================================

type registerAllocator struct {
	mapping RegisterMapping
	limbs   []Register
}

// NewAllocator converts a mapping into a full allocator simply by wrapping the
// two fields.
func NewAllocator(mapping RegisterMapping) RegisterAllocator {
	limbs := slices.Clone(mapping.Limbs())
	return &registerAllocator{mapping, limbs}
}

// AllocateCarry implementation for the schema.RegisterAllocator interface
func (p *registerAllocator) AllocateCarry(prefix string, width uint) RegisterId {
	var (
		// Determine index for new register
		index = uint(len(p.limbs))
		// Determine unique name for new register
		name = fmt.Sprintf("%s$%d", prefix, index)
	)
	// Allocate a new computed register.
	p.limbs = append(p.limbs, NewComputedRegister(name, width))
	//
	return NewRegisterId(index)
}

// BandWidth implementation for schema.RegisterMapping interface
func (p *registerAllocator) Field() FieldConfig {
	return p.mapping.Field()
}

// Limbs implementation for the schema.RegisterMapping interface
func (p *registerAllocator) LimbIds(reg RegisterId) []LimbId {
	return p.mapping.LimbIds(reg)
}

// Limb implementation for the schema.RegisterMapping interface
func (p *registerAllocator) Limb(reg LimbId) Limb {
	return p.limbs[reg.Unwrap()]
}

// Limbs implementation for the schema.RegisterMapping interface
func (p *registerAllocator) Limbs() []Limb {
	return p.limbs
}

// RegisterOf implementation for the schema.RegisterMapping interface
func (p *registerAllocator) RegisterOf(name string) RegisterId {
	return p.mapping.RegisterOf(name)
}
