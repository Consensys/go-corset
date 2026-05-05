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
package instruction

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/word"
	vm_word "github.com/consensys/go-corset/pkg/zkc/vm/internal/word"
)

// Word captures the subset of all instructions which can be executed
// by a word machine.
type Word interface {
	Instruction
	// IsWord demarcates word instructions
	IsWord() bool
}

// vmWord is a convenient alias
type vmWord[W any] = vm_word.Word[W]

// ============================================================================
// Word Instructions
// ============================================================================

// Destruct represents an instruction of the following form:
//
// tn::t0 := r0
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, r0 is the source register which are.
type Destruct = word.Destruct

// NewDestruct constructs a new concatenation instruction which concatenates the
// source registers and writes them into the target register.  Observe that we
// have a little endian ordering here for the target registers.  That is, the
// value of the register targets[0] will be assigned the least significant bits of
// the source value.
func NewDestruct(targets []register.Id, source register.Id) *Destruct {
	return &Destruct{Targets: targets, Source: source}
}

// ============================================================================

// Cast represents a truncating cast instruction of the following form:
//
//	t := (uN)s
//
// Here, t is the target register, s is the source register, and N is the cast
// bit width.  The N low-order bits of s are retained and written to t.
type Cast = word.Cast

// NewCast constructs a new truncating cast instruction.
func NewCast(target register.Id, source register.Id, width uint) *Cast {
	return &Cast{Target: target, Source: source, Width: width}
}

// ============================================================================
// Integer Addition
// ============================================================================

// IntAdd computes the integer sum of the source registers plus a constant
// and writes the result into the target register.  Specifically, the value
// assigned is sources[0] + ... + sources[n-1] + constant, evaluated within
// the bit-width of the target register.  Overflow at runtime aborts
// execution with an arithmetic-overflow error.  The source slice may be
// empty, in which case the instruction simply loads the constant.
type IntAdd[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntAdd constructs a new addition instruction
func NewIntAdd[W vmWord[W]](target register.Id, sources []register.Id, constant W) *IntAdd[W] {
	return &IntAdd[W]{word.NewOpArith(opcode.INT_ADD, target, sources, constant)}
}

// ============================================================================
// Integer Subtraction
// ============================================================================

// IntSub computes a chained subtraction of the source registers and a
// constant, assigning the result to the target register.  The value assigned
// is sources[0] - sources[1] - ... - sources[n-1] - constant, evaluated
// within the bit-width of the target register.  Underflow at runtime aborts
// execution with an arithmetic-underflow error.
type IntSub[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntSub constructs a new subtraction instruction
func NewIntSub[W vmWord[W]](target register.Id, sources []register.Id, constant W) *IntSub[W] {
	return &IntSub[W]{word.NewOpArith(opcode.INT_SUB, target, sources, constant)}
}

// ============================================================================
// Integer Multiplication
// ============================================================================

// IntMul computes the integer product of the source registers and a
// constant, assigning the result to the target register.  The value assigned
// is constant * sources[0] * ... * sources[n-1], evaluated within the
// bit-width of the target register.  Overflow at runtime aborts execution
// with an arithmetic-overflow error.
type IntMul[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntMul constructs a new multiplication instruction
func NewIntMul[W vmWord[W]](target register.Id, sources []register.Id, constant W) *IntMul[W] {
	return &IntMul[W]{word.NewOpArith(opcode.INT_MUL, target, sources, constant)}
}

// ============================================================================
// Integer Division
// ============================================================================

// IntDiv computes the (truncated) integer quotient of two source registers,
// assigning the result to the target register.  Specifically, sources[0] is
// the dividend and sources[1] is the divisor; division by zero aborts
// execution with a division-by-zero error.  The constant operand is unused.
type IntDiv[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntDiv constructs a new division instruction.
func NewIntDiv[W vmWord[W]](target, dividend, divisor register.Id) *IntDiv[W] {
	var zero W
	return &IntDiv[W]{word.NewOpArith[W](opcode.INT_DIV, target, []register.Id{dividend, divisor}, zero)}
}

// ============================================================================
// Integer Remainder
// ============================================================================

// IntRem computes the remainder of the integer division of two source
// registers, assigning the result to the target register.  Specifically,
// sources[0] is the dividend and sources[1] is the divisor; division by zero
// aborts execution with a division-by-zero error.  The constant operand is
// unused.
type IntRem[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntRem constructs a new remainder instruction.
func NewIntRem[W vmWord[W]](target, dividend, divisor register.Id) *IntRem[W] {
	var zero W
	return &IntRem[W]{word.NewOpArith(opcode.INT_REM, target, []register.Id{dividend, divisor}, zero)}
}

// ============================================================================
// Field Addition
// ============================================================================

// IntAddModP computes the sum of the source registers and a constant within
// the prime field of the surrounding machine, assigning the result to the
// target register.  The value assigned is sources[0] + ... + sources[n-1] +
// constant, reduced modulo the field's prime characteristic.  The source
// slice may be empty, in which case the instruction simply loads the
// constant.
type IntAddModP[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntAddModP constructs a new field addition instruction
func NewIntAddModP[W vmWord[W]](target register.Id, sources []register.Id, constant W) *IntAddModP[W] {
	return &IntAddModP[W]{word.NewOpArith(opcode.INT_ADDMOD_P, target, sources, constant)}
}

// ============================================================================
// Field Subtraction
// ============================================================================

// IntSubModP computes a chained subtraction of the source registers and a
// constant within the prime field of the surrounding machine, assigning the
// result to the target register.  The value assigned is sources[0] -
// sources[1] - ... - sources[n-1] - constant, reduced modulo the field's
// prime characteristic.
type IntSubModP[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntSubModP constructs a new field subtraction instruction
func NewIntSubModP[W vmWord[W]](target register.Id, sources []register.Id, constant W) *IntSubModP[W] {
	return &IntSubModP[W]{word.NewOpArith(opcode.INT_SUBMOD_P, target, sources, constant)}
}

// ============================================================================
// Field Multiplication
// ============================================================================

// IntMulModP computes the product of the source registers and a constant
// within the prime field of the surrounding machine, assigning the result
// to the target register.  The value assigned is constant * sources[0] *
// ... * sources[n-1], reduced modulo the field's prime characteristic.
type IntMulModP[W vmWord[W]] struct{ word.OpArith[W] }

// NewIntMulModP constructs a new field multiplication instruction
func NewIntMulModP[W vmWord[W]](target register.Id, sources []register.Id, constant W) *IntMulModP[W] {
	return &IntMulModP[W]{word.NewOpArith(opcode.INT_MULMOD_P, target, sources, constant)}
}

// ============================================================================
// Bitwise And
// ============================================================================

// BitAnd computes the bitwise AND of the source registers and a constant,
// assigning the result to the target register.  The value assigned is
// constant & sources[0] & ... & sources[n-1].  Callers needing AND with no
// constant contribution should pass the AND identity (all-ones within the
// target bit-width) as the constant.
type BitAnd[W vmWord[W]] struct{ word.OpArith[W] }

// NewBitAnd constructs a new bitwise AND instruction.
func NewBitAnd[W vmWord[W]](target register.Id, sources []register.Id, constant W) *BitAnd[W] {
	return &BitAnd[W]{word.NewOpArith(opcode.BIT_AND, target, sources, constant)}
}

// ============================================================================
// Bitwise Not
// ============================================================================

// BitNot computes the bitwise complement of a single source register and
// assigns the result to the target register.  The complement is taken within
// the bit-width of the target register.  The constant operand is unused.
type BitNot[W vmWord[W]] struct{ word.OpArith[W] }

// NewBitNot constructs a new bitwise NOT instruction.
func NewBitNot[W vmWord[W]](target, source register.Id) *BitNot[W] {
	var zero W
	return &BitNot[W]{word.NewOpArith(opcode.BIT_NOT, target, []register.Id{source}, zero)}
}

// ============================================================================
// Bitwise Or
// ============================================================================

// BitOr computes the bitwise OR of the source registers and a constant,
// assigning the result to the target register.  The value assigned is
// constant | sources[0] | ... | sources[n-1].
type BitOr[W vmWord[W]] struct{ word.OpArith[W] }

// NewBitOr constructs a new bitwise OR instruction.
func NewBitOr[W vmWord[W]](target register.Id, sources []register.Id, constant W) *BitOr[W] {
	return &BitOr[W]{word.NewOpArith(opcode.BIT_OR, target, sources, constant)}
}

// ============================================================================
// Bitwise Xor
// ============================================================================

// BitXor computes the bitwise exclusive-OR of the source registers and a
// constant, assigning the result to the target register.  The value assigned
// is constant ^ sources[0] ^ ... ^ sources[n-1].
type BitXor[W vmWord[W]] struct{ word.OpArith[W] }

// NewBitXor constructs a new bitwise XOR instruction.
func NewBitXor[W vmWord[W]](target register.Id, sources []register.Id, constant W) *BitXor[W] {
	return &BitXor[W]{word.NewOpArith(opcode.BIT_XOR, target, sources, constant)}
}

// ============================================================================
// Bitwise Shift Left
// ============================================================================

// BitShl computes the bitwise left-shift of one source register by another,
// assigning the result to the target register.  Specifically, sources[0] is
// the value to be shifted and sources[1] is the shift amount, with the
// result evaluated within the bit-width of the target register.  The
// constant operand is unused.
type BitShl[W vmWord[W]] struct{ word.OpArith[W] }

// NewBitShl constructs a new bitwise left-shift instruction.
func NewBitShl[W vmWord[W]](target, value, amount register.Id) *BitShl[W] {
	var zero W
	return &BitShl[W]{word.NewOpArith[W](opcode.BIT_SHL, target, []register.Id{value, amount}, zero)}
}

// ============================================================================

// BitShr computes the bitwise (logical) right-shift of one source register
// by another, assigning the result to the target register.  Specifically,
// sources[0] is the value to be shifted and sources[1] is the shift amount.
// The constant operand is unused.
type BitShr[W vmWord[W]] struct{ word.OpArith[W] }

// NewBitShr constructs a new bitwise right-shift instruction.
func NewBitShr[W vmWord[W]](target, value, amount register.Id) *BitShr[W] {
	var zero W
	return &BitShr[W]{word.NewOpArith(opcode.BIT_SHR, target, []register.Id{value, amount}, zero)}
}

// ============================================================================

// BitConcat concatenates the source registers and writes the joined value
// into the target register.  The source ordering is little-endian: the value
// in sources[0] occupies the least-significant bits of the result, sources[1]
// the next-least-significant bits, and so on.  The constant operand is
// unused.
type BitConcat[W vmWord[W]] struct{ word.OpArith[W] }

// NewBitConcat constructs a new concatenation instruction which concatenates
// the source registers and writes them into the target register.  Observe
// that we have a little endian ordering here for the source registers.  That
// is, the value of the register sources[0] will occupy the least significant
// bits of the result.
func NewBitConcat[W vmWord[W]](target register.Id, sources []register.Id) *BitConcat[W] {
	var zero W
	return &BitConcat[W]{word.NewOpArith(opcode.BIT_CONCAT, target, sources, zero)}
}
