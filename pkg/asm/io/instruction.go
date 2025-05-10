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
	"math"
	"math/big"
)

// RETURN is used to signal that a given instruction returns from the enclosing
// function.
const RETURN = math.MaxUint

// UNKNOWN_BUS signals a bus which is unknown.
const UNKNOWN_BUS = math.MaxUint

// Environment captures all things necessary to validate a particular
// instruction is well-formed.
type Environment[T any] struct {
	// Maximum width of the underlying field
	FieldWidth uint
	// Id of enclosing function
	Function uint
	// Enclosing program
	Program Program[T]
}

// Enclosing returns the enclosing component for this environment.
func (p *Environment[T]) Enclosing() Function[T] {
	return p.Program.Function(p.Function)
}

// Instruction provides an abstract notion of an executable "machine instruction".
type Instruction[T any] interface {
	// Execute a given instruction at a given program counter position, using a
	// given set of register values.  This may update the register values, and
	// returns the next program counter position.  If the program counter is
	// math.MaxUint then a return is signaled.
	Execute(pc uint, state []big.Int, regs []Register) uint
	// Registers returns the set of registers read this micro instruction.
	RegistersRead() []uint
	// Registers returns the set of registers written by this micro instruction.
	RegistersWritten() []uint
	// Validate that this instruction is well-formed.  For example, that it is
	// balanced, that there are no conflicting writes, that all temporaries have
	// been allocated, etc.  The maximum bit capacity of the underlying field is
	// needed for this calculation, so as to allow an instruction to check it
	// does not overflow the underlying field.
	Validate(env Environment[T]) error
	// Produce a suitable string representation of this instruction.  This is
	// primarily used for debugging.
	String(env Environment[T]) string
}
