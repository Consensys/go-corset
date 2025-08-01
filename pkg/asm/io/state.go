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
package io

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"

	"github.com/consensys/go-corset/pkg/schema"
	util_math "github.com/consensys/go-corset/pkg/util/math"
)

// RETURN is used to signal that a given instruction returns from the enclosing
// function.
const RETURN uint = math.MaxUint

// State collects together local state necessary for executing a given
// instruction.  This contrasts with an I/O Map which represents the non-local
// state.
type State struct {
	// Program Counter position.
	pc uint
	// Values for each register in this state excluding the program counter
	// (since this is held above).  Thus, this array has one less item than
	// registers.
	state []big.Int
	// Registers referenced in this state.  This is necessary to determine
	// appropriate bitwidths for copying data, and also for debugging.
	registers []Register
	// Io subsystem is necessary for enabling reads / writes from I/O buses.
	io Map
}

// EmptyState constructs an initially empty state at the given PC value.  One
// can then set register values as needed via Store.
func EmptyState(pc uint, registers []schema.Register, io Map) State {
	var state = make([]big.Int, len(registers)-1)
	// Construct state
	return State{pc, state, registers, io}
}

// InitialState constructs a suitable initial state for executing a given
// function with the given arguments.
func InitialState[T Instruction[T]](arguments []big.Int, fn Function[T], io Map) State {
	var (
		state = EmptyState(0, fn.registers, io)
		index = 0
	)
	// Initialise arguments
	for i, reg := range fn.Registers() {
		if reg.IsInput() {
			var (
				rid = schema.NewRegisterId(uint(i))
				val = arguments[index]
				ith big.Int
			)
			// Clone big int
			ith.Set(&val)
			state.Store(rid, ith)
			//
			index = index + 1
		}
	}
	// Construct state
	return state
}

// Clone this state, producing a disjoint state.
func (p *State) Clone() State {
	return State{
		p.pc,
		slices.Clone(p.state),
		p.registers,
		p.io,
	}
}

// Goto updates the program counter for this state to a given value.
func (p *State) Goto(pc uint) {
	p.pc = pc
}

// In performs an I/O read across a given bus.  More specifically, it reads the
// value at a given address on the bus.
func (p *State) In(bus Bus) {
	var address = p.LoadN(bus.Address())
	// Read value from I/O bus
	values := p.io.Read(bus.BusId, address)
	// Write them back
	p.StoreN(bus.Data(), values)
}

// Outputs extracts values from output registers of the given state.
func (p *State) Outputs() []big.Int {
	// Construct outputs
	outputs := make([]big.Int, 0)
	//
	for i, reg := range p.registers {
		if reg.IsOutput() {
			outputs = append(outputs, p.state[i])
		}
	}
	//
	return outputs
}

// Load value of a given register from this state.
func (p *State) Load(reg RegisterId) *big.Int {
	index := reg.Unwrap()
	//
	if index == 0 {
		return big.NewInt(int64(p.pc))
	}
	//
	return &p.state[index-1]
}

// LoadN reads the values of zero or more registers from this state.
func (p *State) LoadN(registers []RegisterId) []big.Int {
	values := make([]big.Int, len(registers))
	//
	for i, src := range registers {
		values[i] = *p.Load(src)
	}
	//
	return values
}

// Out performs an I/O write across a given bus.  More specifically, it sets the
// value at a given address on the bus.
func (p *State) Out(bus Bus) {
	var (
		address = p.LoadN(bus.Address())
		data    = p.LoadN(bus.Data())
	)

	p.io.Write(bus.BusId, address, data)
}

// Pc returns the current program counter position.
func (p *State) Pc() uint {
	return p.pc
}

// Store value to a given register from this state.
func (p *State) Store(reg RegisterId, value big.Int) {
	index := reg.Unwrap()
	//
	if value.BitLen() > int(p.registers[index].Width) {
		panic("write exceeds register width")
	} else if index == 0 {
		p.pc = uint(value.Uint64())
	} else {
		// Write to normal register
		p.state[index-1] = value
	}
}

// StoreAcross a given value across a set of registers, splitting its bits as
// necessary.  The target registers are given with the least significant first.
// For example, consider writing 01100010 to registers [R1, R2] of type u4.
// Then, after the write, we have R1=0010 and R2=0110.
func (p *State) StoreAcross(value big.Int, registers ...RegisterId) {
	var (
		offset uint    = 0
		val    big.Int = value
		sign           = val.Sign() >= 0
	)
	// Check for negative values
	if !sign {
		val = big.Int{}
		// Clone value before mutating it
		val.Set(&value)
		//
		width := schema.WidthOfRegisters(p.registers, registers)
		// Normalise negative value
		val.Add(&val, util_math.Pow2(width))
	}
	//
	for _, id := range registers {
		width := p.registers[id.Unwrap()].Width
		p.Store(id, ReadBitSlice(offset, width, val, sign))
		offset += width
	}
}

// StoreN writes a set of zero or more values to a corresponding set of
// registers in this state.
func (p *State) StoreN(registers []RegisterId, values []big.Int) {
	for i, dst := range registers {
		p.Store(dst, values[i])
	}
}

// String produces a string representation of the given execution state.
func (p *State) String() string {
	var builder strings.Builder
	//
	if p.Terminated() {
		builder.WriteString("(pc=--) ")
	} else {
		pc := fmt.Sprintf("(pc=%02x) ", p.pc)
		builder.WriteString(pc)
	}
	//
	for i := range p.registers {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		val := p.Load(schema.NewRegisterId(uint(i))).Text(16)
		reg := p.registers[i].Name
		builder.WriteString(fmt.Sprintf("%s=0x%s", reg, val))
	}
	//
	return builder.String()
}

// Terminated determines whether this state represents a terminated function
// execution.
func (p *State) Terminated() bool {
	return p.pc == RETURN
}

// ReadBitSlice reads a slice of bits starting at a given offset in a give
// value.  For example, consider the value is 10111000 and we have offset=1 and
// width=4, then the result is 1100.
func ReadBitSlice(offset uint, width uint, value big.Int, sign bool) big.Int {
	var (
		slice big.Int
		bit   uint
		n     = int(offset + width)
		m     = value.BitLen()
		i     = int(offset)
		j     = 0
	)
	// Read bits upto end
	for ; i < min(n, m); i, j = i+1, j+1 {
		// Read appropriate bit
		bit = value.Bit(i)
		// set appropriate bit
		slice.SetBit(&slice, j, bit)
	}
	// Sign extend (negative values)
	if !sign {
		// Negative value
		for ; i < n; i, j = i+1, j+1 {
			// set appropriate bit
			slice.SetBit(&slice, j, 1)
		}
	}
	//
	return slice
}
