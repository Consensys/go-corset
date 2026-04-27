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
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// ============================================================================
// Integer Addition
// ============================================================================

// IntAdd computes the integer sum of the source registers plus a constant
// and writes the result into the target register.  Specifically, the value
// assigned is sources[0] + ... + sources[n-1] + constant, evaluated within
// the bit-width of the target register.  Overflow at runtime aborts
// execution with an arithmetic-overflow error.  The source slice may be
// empty, in which case the instruction simply loads the constant.
type IntAdd[W word.Word[W]] struct{ OpArith[W] }

// NewIntAdd constructs a new addition instruction
func NewIntAdd[W word.Word[W]](target register.Id, sources []register.Id, constant W) *IntAdd[W] {
	return &IntAdd[W]{OpArith[W]{INT_ADD, target, sources, constant}}
}

// ============================================================================
// Integer Subtraction
// ============================================================================

// IntSub computes a chained subtraction of the source registers and a
// constant, assigning the result to the target register.  The value assigned
// is sources[0] - sources[1] - ... - sources[n-1] - constant, evaluated
// within the bit-width of the target register.  Underflow at runtime aborts
// execution with an arithmetic-underflow error.
type IntSub[W word.Word[W]] struct{ OpArith[W] }

// NewIntSub constructs a new subtraction instruction
func NewIntSub[W word.Word[W]](target register.Id, sources []register.Id, constant W) *IntSub[W] {
	return &IntSub[W]{OpArith[W]{INT_SUB, target, sources, constant}}
}

// ============================================================================
// Integer Multiplication
// ============================================================================

// IntMul computes the integer product of the source registers and a
// constant, assigning the result to the target register.  The value assigned
// is constant * sources[0] * ... * sources[n-1], evaluated within the
// bit-width of the target register.  Overflow at runtime aborts execution
// with an arithmetic-overflow error.
type IntMul[W word.Word[W]] struct{ OpArith[W] }

// NewIntMul constructs a new multiplication instruction
func NewIntMul[W word.Word[W]](target register.Id, sources []register.Id, constant W) *IntMul[W] {
	return &IntMul[W]{OpArith[W]{INT_MUL, target, sources, constant}}
}

// ============================================================================
// Integer Division
// ============================================================================

// IntDiv computes the (truncated) integer quotient of two source registers,
// assigning the result to the target register.  Specifically, sources[0] is
// the dividend and sources[1] is the divisor; division by zero aborts
// execution with a division-by-zero error.  The constant operand is unused.
type IntDiv[W word.Word[W]] struct{ OpArith[W] }

// NewIntDiv constructs a new division instruction.
func NewIntDiv[W word.Word[W]](target, dividend, divisor register.Id) *IntDiv[W] {
	var zero W
	return &IntDiv[W]{OpArith[W]{INT_DIV, target, []register.Id{dividend, divisor}, zero}}
}

// ============================================================================
// Integer Remainder
// ============================================================================

// IntRem computes the remainder of the integer division of two source
// registers, assigning the result to the target register.  Specifically,
// sources[0] is the dividend and sources[1] is the divisor; division by zero
// aborts execution with a division-by-zero error.  The constant operand is
// unused.
type IntRem[W word.Word[W]] struct{ OpArith[W] }

// NewIntRem constructs a new remainder instruction.
func NewIntRem[W word.Word[W]](target, dividend, divisor register.Id) *IntRem[W] {
	var zero W
	return &IntRem[W]{OpArith[W]{INT_REM, target, []register.Id{dividend, divisor}, zero}}
}

// ============================================================================
// Bitwise And
// ============================================================================

// BitAnd computes the bitwise AND of the source registers and a constant,
// assigning the result to the target register.  The value assigned is
// constant & sources[0] & ... & sources[n-1].  Callers needing AND with no
// constant contribution should pass the AND identity (all-ones within the
// target bit-width) as the constant.
type BitAnd[W word.Word[W]] struct{ OpArith[W] }

// NewBitAnd constructs a new bitwise AND instruction.
func NewBitAnd[W word.Word[W]](target register.Id, sources []register.Id, constant W) *BitAnd[W] {
	return &BitAnd[W]{OpArith[W]{BIT_AND, target, sources, constant}}
}

// ============================================================================
// Bitwise Not
// ============================================================================

// BitNot computes the bitwise complement of a single source register and
// assigns the result to the target register.  The complement is taken within
// the bit-width of the target register.  The constant operand is unused.
type BitNot[W word.Word[W]] struct{ OpArith[W] }

// NewBitNot constructs a new bitwise NOT instruction.
func NewBitNot[W word.Word[W]](target, source register.Id) *BitNot[W] {
	var zero W
	return &BitNot[W]{OpArith[W]{BIT_NOT, target, []register.Id{source}, zero}}
}

// ============================================================================
// Bitwise Or
// ============================================================================

// BitOr computes the bitwise OR of the source registers and a constant,
// assigning the result to the target register.  The value assigned is
// constant | sources[0] | ... | sources[n-1].
type BitOr[W word.Word[W]] struct{ OpArith[W] }

// NewBitOr constructs a new bitwise OR instruction.
func NewBitOr[W word.Word[W]](target register.Id, sources []register.Id, constant W) *BitOr[W] {
	return &BitOr[W]{OpArith[W]{BIT_OR, target, sources, constant}}
}

// ============================================================================
// Bitwise Xor
// ============================================================================

// BitXor computes the bitwise exclusive-OR of the source registers and a
// constant, assigning the result to the target register.  The value assigned
// is constant ^ sources[0] ^ ... ^ sources[n-1].
type BitXor[W word.Word[W]] struct{ OpArith[W] }

// NewBitXor constructs a new bitwise XOR instruction.
func NewBitXor[W word.Word[W]](target register.Id, sources []register.Id, constant W) *BitXor[W] {
	return &BitXor[W]{OpArith[W]{BIT_XOR, target, sources, constant}}
}

// ============================================================================
// Bitwise Shift Left
// ============================================================================

// BitShl computes the bitwise left-shift of one source register by another,
// assigning the result to the target register.  Specifically, sources[0] is
// the value to be shifted and sources[1] is the shift amount, with the
// result evaluated within the bit-width of the target register.  The
// constant operand is unused.
type BitShl[W word.Word[W]] struct{ OpArith[W] }

// NewBitShl constructs a new bitwise left-shift instruction.
func NewBitShl[W word.Word[W]](target, value, amount register.Id) *BitShl[W] {
	var zero W
	return &BitShl[W]{OpArith[W]{BIT_SHL, target, []register.Id{value, amount}, zero}}
}

// ============================================================================
// Bitwise Shift Right
// ============================================================================

// BitShr computes the bitwise (logical) right-shift of one source register
// by another, assigning the result to the target register.  Specifically,
// sources[0] is the value to be shifted and sources[1] is the shift amount.
// The constant operand is unused.
type BitShr[W word.Word[W]] struct{ OpArith[W] }

// NewBitShr constructs a new bitwise right-shift instruction.
func NewBitShr[W word.Word[W]](target, value, amount register.Id) *BitShr[W] {
	var zero W
	return &BitShr[W]{OpArith[W]{BIT_SHR, target, []register.Id{value, amount}, zero}}
}

// ============================================================================
// Bitwise Concatenarion
// ============================================================================

// BitConcat concatenates the source registers and writes the joined value
// into the target register.  The source ordering is little-endian: the value
// in sources[0] occupies the least-significant bits of the result, sources[1]
// the next-least-significant bits, and so on.  The constant operand is
// unused.
type BitConcat[W word.Word[W]] struct{ OpArith[W] }

// NewBitConcat constructs a new concatenation instruction which concatenates
// the source registers and writes them into the target register.  Observe
// that we have a little endian ordering here for the source registers.  That
// is, the value of the register sources[0] will occupy the least significant
// bits of the result.
func NewBitConcat[W word.Word[W]](target register.Id, sources []register.Id) *BitConcat[W] {
	var zero W
	return &BitConcat[W]{OpArith[W]{BIT_CONCAT, target, sources, zero}}
}

// ============================================================================
// Opcode-Register-Registers-Constant instruction type
// ============================================================================

// OpArith represents an instruction of the following form:
//
// t0 := r0 # ... # rn + c
//
// Here, t0 is the *target register*, whilst r0 .. rn are the source registers
// and c is a constant (which can be 0).  Finally, "#" represents whatever
// operation the given opcode indicates.
type OpArith[W word.Word[W]] struct {
	Op OpCode
	// Target register for assignment
	Target register.Id
	// Source registers for assignment
	Sources []register.Id
	// Constant for assignment
	Constant W
}

// OpCode implementation for Instruction interface
func (p *OpArith[W]) OpCode() OpCode {
	return p.Op
}

// Uses implementation for Instruction interface
func (p *OpArith[W]) Uses() []register.Id {
	return p.Sources
}

// Definitions implementation for Instruction interface
func (p *OpArith[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

// MicroValidate implementation for MicroInstruction interface.
func (p *OpArith[W]) MicroValidate(_ uint, field field.Config, _ SystemMap[W]) []error {
	return nil
}

func (p *OpArith[W]) String(mapping SystemMap[W]) string {
	var (
		builder strings.Builder
		op      = aType2Operation(p.Op)
	)
	//
	builder.WriteString(registersToString(mapping, p.Target))
	builder.WriteString(" = ")
	builder.WriteString(expressionToString(op, p.Sources, p.Constant, mapping))
	//
	return builder.String()
}

func aType2Operation(op OpCode) string {
	switch op {
	case INT_ADD:
		return "+"
	case INT_SUB:
		return "-"
	case INT_MUL:
		return "*"
	case INT_DIV:
		return "/"
	case INT_REM:
		return "%"
	case BIT_AND:
		return "&"
	case BIT_NOT:
		return "~"
	case BIT_OR:
		return "|"
	case BIT_XOR:
		return "^"
	case BIT_SHL:
		return "<<"
	case BIT_SHR:
		return ">>"
	case BIT_CONCAT:
		return "::"
	default:
		panic("unknown type A instruction")
	}
}
