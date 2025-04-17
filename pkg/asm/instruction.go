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
package asm

// Instruction provides an abstract notion of a "machine instruction".
type Instruction interface {
}

// Add two source registers of a given bitwidth together, writing the result to
// a given output register.  The result is written modulo the given bitwidth,
// with the carry flag set accordingly.
type Add struct {
	// Width determines the bitwidth of this operation.  Both registers must
	// have matching bitwidths.
	Width uint
	// Destination register where the outcome is written.
	Rdest uint
	// Left source operand
	Rsrcl uint
	// Right source operand
	Rsrcr uint
}

// Sub subtracts one source register from another, writing the result to a given
// output register.  The result is written modulo the given bitwidth, with the
// borrow flag set accordingly.
type Sub struct {
	// Width determines the bitwidth of this operation.  Both registers must
	// have matching bitwidths.
	Width uint
	// Destination register where the outcome is written.
	Rdest uint
	// Left source operand
	Rsrcl uint
	// Right source operand
	Rsrcr uint
}

// Jmp provides an unconditional branching instruction to a given instructon.
type Jmp struct {
	Target uint
}

// Jcond describes the family of conditional branching instructions.  For
// example, "jz" jumps when the zero flag is set whilst "jc" when the carry flag
// is set, etc.
type Jcond struct {
	Target uint
}
