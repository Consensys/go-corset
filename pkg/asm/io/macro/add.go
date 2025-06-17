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
	"github.com/consensys/go-corset/pkg/schema"
)

// Add represents a generic operation of the following form:
//
// tn, .., t0 := s0 + ... + sm + c
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, s0 ... sm are the source registers, and c is a
// given (non-negative) constant. Observe the n == m is not required, meaning
// one can assign multiple registers.  For example, consider this case:
//
// c, r0 := r1 + 1
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 + 1 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to c.  Thus, in this
// particular example, c represents a carry flag.
type Add struct {
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
func (p *Add) Execute(state io.State) uint {
	var value big.Int
	// Add constant
	value.Set(&p.Constant)
	// Add register values
	for _, src := range p.Sources {
		value.Add(&value, state.Load(src))
	}
	// Write value across targets
	state.Store(value, p.Targets...)
	//
	return state.Next()
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Add) Lower(pc uint) micro.Instruction {
	code := &micro.Add{
		Targets:  p.Targets,
		Sources:  p.Sources,
		Constant: p.Constant,
	}
	// Lowering here produces an instruction containing a single microcode.
	return micro.NewInstruction(code, &micro.Jmp{Target: pc + 1})
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Add) RegistersRead() []io.RegisterId {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Add) RegistersWritten() []io.RegisterId {
	return p.Targets
}

func (p *Add) String(fn schema.Module) string {
	return assignmentToString(p.Targets, p.Sources, p.Constant, fn, zero, " + ")
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Add) Validate(fieldWidth uint, fn schema.Module) error {
	var (
		regs     = fn.Registers()
		lhs_bits = sumTargetBits(p.Targets, regs)
		rhs_bits = sumSourceBits(p.Sources, p.Constant, regs)
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

func sumSourceBits(sources []io.RegisterId, constant big.Int, regs []io.Register) uint {
	var rhs big.Int
	//
	for _, target := range sources {
		rhs.Add(&rhs, regs[target.Unwrap()].MaxValue())
	}
	// Include constant (if relevant)
	rhs.Add(&rhs, &constant)
	//
	return uint(rhs.BitLen())
}

// Sum the total number of bits used by the given set of target registers.
func sumTargetBits(targets []io.RegisterId, regs []io.Register) uint {
	sum := uint(0)
	//
	for _, target := range targets {
		sum += regs[target.Unwrap()].Width
	}
	//
	return sum
}
