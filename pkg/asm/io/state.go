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
)

// RETURN is used to signal that a given instruction returns from the enclosing
// function.
const RETURN uint = math.MaxUint

// State collects together local state necessary for executing a given
// instruction.  This contrasts with an I/O Map which represents the non-local
// state.
type State struct {
	// Program Counter position.
	Pc uint
	// Values for each register in this state.
	State []big.Int
	// Registers referenced in this state.  This is necessary to determine
	// appropriate bitwidths for copying data, and also for debugging.
	Registers []Register
	// Io subsystem is necessary for enabling reads / writes from I/O buses.
	Io Map
}

// InitialState constructs a suitable initial state for executing a given
// function with the given arguments.
func InitialState[T Instruction[T]](arguments []big.Int, fn Function[T], io Map) State {
	var (
		state = make([]big.Int, len(fn.Registers()))
		index = 0
	)
	// Initialise arguments
	for i, reg := range fn.Registers() {
		if reg.IsInput() {
			var (
				val = arguments[index]
				ith big.Int
			)
			// Clone big int
			ith.Set(&val)
			//
			state[i] = ith
			index = index + 1
		}
	}
	// Construct state
	return State{0, state, fn.Registers(), io}
}

// Clone this state, producing a disjoint state.
func (p *State) Clone() State {
	return State{
		p.Pc,
		slices.Clone(p.State),
		p.Registers,
		p.Io,
	}
}

// In performs an I/O read across a given bus.  More specifically, it reads the
// value at a given address on the bus.
func (p *State) In(bus Bus) {
	var address = p.LoadN(bus.Address())
	// Read value from I/O bus
	values := p.Io.Read(bus.BusId, address)
	// Write them back
	p.StoreN(bus.Data(), values)
}

// Outputs extracts values from output registers of the given state.
func (p *State) Outputs() []big.Int {
	// Construct outputs
	outputs := make([]big.Int, 0)
	//
	for i, reg := range p.Registers {
		if reg.IsOutput() {
			outputs = append(outputs, p.State[i])
		}
	}
	//
	return outputs
}

// Load value of a given register from this state.
func (p *State) Load(reg RegisterId) *big.Int {
	return &p.State[reg.Unwrap()]
}

// LoadN reads the values of zero or more registers from this state.
func (p *State) LoadN(registers []RegisterId) []big.Int {
	values := make([]big.Int, len(registers))
	//
	for i, src := range registers {
		values[i] = p.State[src.Unwrap()]
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

	p.Io.Write(bus.BusId, address, data)
}

// Next returns the program counter for the following instruction.
func (p *State) Next() uint {
	return p.Pc + 1
}

// Store a given value across a set of registers, splitting its bits as
// necessary.  The target registers are given with the least significant first.
// For example, consider writing 01100010 to registers [R1, R2] of type u4.
// Then, after the write, we have R1=0010 and R2=0110.
func (p *State) Store(value big.Int, registers ...RegisterId) {
	var offset uint = 0
	//
	for _, id := range registers {
		reg := id.Unwrap()
		width := p.Registers[reg].Width
		p.State[reg] = ReadBitSlice(offset, width, value)
		offset += width
	}
}

// StoreN writes a set of zero or more values to a corresponding set of
// registers in this state.
func (p *State) StoreN(registers []RegisterId, values []big.Int) {
	for i, dst := range registers {
		p.State[dst.Unwrap()] = values[i]
	}
}

// String produces a string representation of the given execution state.
func (p *State) String() string {
	var builder strings.Builder
	//
	if p.Terminated() {
		builder.WriteString("(pc=--) ")
	} else {
		pc := fmt.Sprintf("(pc=%02x) ", p.Pc)
		builder.WriteString(pc)
	}
	//
	for i := 0; i != len(p.Registers); i++ {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		val := p.State[i].Text(16)
		reg := p.Registers[i].Name
		builder.WriteString(fmt.Sprintf("%s=0x%s", reg, val))
	}
	//
	return builder.String()
}

// Terminated determines whether this state represents a terminated function
// execution.
func (p *State) Terminated() bool {
	return p.Pc == RETURN
}

// ReadBitSlice reads a slice of bits starting at a given offset in a give
// value.  For example, consider the value is 10111000 and we have offset=1 and
// width=4, then the result is 1100.
func ReadBitSlice(offset uint, width uint, value big.Int) big.Int {
	var slice big.Int
	//
	for i := 0; uint(i) < width; i++ {
		// Read appropriate bit
		bit := value.Bit(i + int(offset))
		// set appropriate bit
		slice.SetBit(&slice, i, bit)
	}
	//
	return slice
}
