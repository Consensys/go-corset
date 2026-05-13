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
package function

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
)

// Function --- see documentation on vm.Function.
type Function[I instruction.Instruction] struct {
	// Unique name of this function.
	name string
	// Registers describes zero or more registers of a given width.  Each
	// register can be designated as an input / output or temporary.
	registers []register.Register
	// Number of input registers
	numInputs uint
	// Number of output registers
	numOutputs uint
	// Code defines the body of this function.
	code []instruction.Vector[I]
}

// New constructs a new function with the given components.
func New[I instruction.Instruction](name string, registers []register.Register,
	code []instruction.Vector[I]) *Function[I] {
	//
	var (
		numInputs  = array.CountMatching(registers, func(r register.Register) bool { return r.IsInput() })
		numOutputs = array.CountMatching(registers, func(r register.Register) bool { return r.IsOutput() })
	)
	// Check registers sorted as: inputs, outputs then internal.
	if !set.IsSorted(registers, func(r register.Register) register.Type { return r.Kind() }) {
		panic("function registers ordered incorrectly")
	}
	// All good
	return &Function[I]{name, registers, numInputs, numOutputs, code}
}

// CodeAt returns the ith instruction making up the body of this function.
func (p *Function[I]) CodeAt(i uint) instruction.Vector[I] {
	return p.code[i]
}

// Code returns the instructions making up the body of this function.
func (p *Function[I]) Code() []instruction.Vector[I] {
	return p.code
}

// IsAtomic determines whether or not this is a "one line function".  That is,
// where every instance of this function occupies exactly one line in the
// corresponding trace.  This is useful to know, as certain optimisations can be
// applied for one line functions (e.g. no PC register is required).
func (p *Function[I]) IsAtomic() bool {
	return len(p.code) == 1
}

// HasRegister checks whether a register with the given name exists and, if
// so, returns its register identifier.  Otherwise, it returns false.
func (p *Function[I]) HasRegister(name string) (register.Id, bool) {
	for i, r := range p.registers {
		if r.Name() == name {
			return register.NewId(uint(i)), true
		}
	}
	// Failed
	return register.UnusedId(), false
}

// Inputs returns the set of input registers for this function.
func (p *Function[I]) Inputs() []register.Register {
	return p.registers[:p.numInputs]
}

// NumInputs returns the number of input registers for this function.
func (p *Function[I]) NumInputs() uint {
	return p.numInputs
}

// NumOutputs returns the number of output registers for this function.
func (p *Function[I]) NumOutputs() uint {
	return p.numOutputs
}

// Name returns the name of this function.
func (p *Function[I]) Name() string {
	// Functions always have a multiplier of 1.
	return p.name
}

// Outputs returns the set of output registers for this function.
func (p *Function[I]) Outputs() []register.Register {
	return p.registers[p.numInputs : p.numInputs+p.numOutputs]
}

// Register returns the ith register used in this function.
func (p *Function[I]) Register(id register.Id) register.Register {
	return p.registers[id.Unwrap()]
}

// Registers returns the set of all registers used during execution of this
// function.
func (p *Function[I]) Registers() []register.Register {
	return p.registers
}

// Width returns the number of registers in this module.'
func (p *Function[I]) Width() uint {
	return uint(len(p.registers))
}
