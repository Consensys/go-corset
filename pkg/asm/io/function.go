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

// Function defines a distinct functional entity within the system.  Functions
// accepts zero or more inputs and produce zero or more outputs.  Functions
// declare zero or more internal registers for use, and their interpretation is
// given by a sequence of zero or more instructions.
type Function[T any] struct {
	// Unique name of this function.
	name string
	// Registers describes zero or more registers of a given width.  Each
	// register can be designated as an input / output or temporary.
	registers []Register
	// Code defines the body of this function.
	code []T
}

// NewFunction constructs a new function with the given components.
func NewFunction[T any](name string, registers []Register, code []T) Function[T] {
	return Function[T]{name, registers, code}
}

// CodeAt returns the ith instruction making up the body of this function.
func (p *Function[T]) CodeAt(i uint) T {
	return p.code[i]
}

// Code returns the instructions making up the body of this function.
func (p *Function[T]) Code() []T {
	return p.code
}

// Inputs returns the set of input registers for this function.
func (p *Function[T]) Inputs() []Register {
	var inputs []Register
	//
	for _, r := range p.registers {
		if r.IsInput() {
			inputs = append(inputs, r)
		}
	}
	//
	return inputs
}

// Name returns the name of this function.
func (p *Function[T]) Name() string {
	return p.name
}

// Outputs returns the set of output registers for this function.
func (p *Function[T]) Outputs() []Register {
	var outputs []Register
	//
	for _, r := range p.registers {
		if r.IsOutput() {
			outputs = append(outputs, r)
		}
	}
	//
	return outputs
}

// Register returns the ith register used in this function.
func (p *Function[T]) Register(i uint) Register {
	return p.registers[i]
}

// Registers returns the set of all registers used during execution of this
// function.
func (p *Function[T]) Registers() []Register {
	return p.registers
}

// AllocateRegister allocates a new register of the given kind, name and width
// into this function.
func (p *Function[T]) AllocateRegister(kind uint8, name string, width uint) uint {
	index := uint(len(p.registers))
	p.registers = append(p.registers, NewRegister(kind, name, width))
	// Done
	return index
}
