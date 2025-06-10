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
package macro

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
)

// Mul represents a generic operation of the following form:
//
// tn, .., t0 := s0 * ... * sm * c
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, s0 ... sm are the source registers, and c is a
// given (non-negative) constant. Observe the n == m is not required, meaning
// one can assign multiple registers.  For example, consider this case:
//
// c, r0 := r1 * 2
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 * 2 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to c.  Thus, in this
// particular example, c represents a carry flag.
type Mul struct {
	// Target registers for addition
	Targets []io.RegisterId
	// Source register for addition
	Sources []io.RegisterId
	// Constant value (if applicable)
	Constant big.Int
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *Mul) Execute(state io.State) uint {
	var value big.Int
	// Assign first value
	value.Set(state.Load(p.Sources[0]))
	// Multiply register values
	for _, src := range p.Sources[1:] {
		value.Mul(&value, state.Load(src))
	}
	// Multiply constant
	value.Mul(&value, &p.Constant)
	// Write value
	state.Store(value, p.Targets...)
	//
	return state.Next()
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Mul) Lower(pc uint) micro.Instruction {
	code := &micro.Mul{
		Targets:  p.Targets,
		Sources:  p.Sources,
		Constant: p.Constant,
	}
	// Lowering here produces an instruction containing a single microcode.
	return micro.NewInstruction(code, &micro.Jmp{Target: pc + 1})
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Mul) RegistersRead() []io.RegisterId {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Mul) RegistersWritten() []io.RegisterId {
	return p.Targets
}

func (p *Mul) String(fn io.Function[Instruction]) string {
	return assignmentToString(p.Targets, p.Sources, p.Constant, fn, one, " * ")
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Mul) Validate(fieldWidth uint, fn io.Function[Instruction]) error {
	var (
		regs     = fn.Registers()
		lhs_bits = sumTargetBits(p.Targets, regs)
		rhs_bits = mulSourceBits(p.Sources, p.Constant, regs)
	)
	// check
	if lhs_bits < rhs_bits {
		return fmt.Errorf("bit overflow (%d bits into %d bits)", rhs_bits, lhs_bits)
	} else if rhs_bits > fieldWidth {
		return fmt.Errorf("field overflow (%d bits into %d bit field)", rhs_bits, fieldWidth)
	}
	//
	return io.CheckTargetRegisters(p.Targets, regs)
}

func mulSourceBits(sources []io.RegisterId, constant big.Int, regs []io.Register) uint {
	var rhs big.Int
	//
	rhs.Set(&one)
	//
	for _, target := range sources {
		rhs.Mul(&rhs, regs[target.Unwrap()].MaxValue())
	}
	// Include constant (if relevant)
	rhs.Mul(&rhs, &constant)
	//
	return uint(rhs.BitLen())
}
