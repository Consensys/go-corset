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
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// Sub represents a generic operation of the following form:
//
// tn, .., t0 := s0 - ... - sm - c
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, s0 ... sm are the source registers, and c is a
// given (non-negative) constant. Observe the n == m is not required, meaning
// one can assign multiple registers.  For example, consider this case:
//
// b, r0 := r1 - 1
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 - 2 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to b.  Thus, in this
// particular example, b represents a borrow flag.
type Sub struct {
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
func (p *Sub) Execute(state io.State) uint {
	var value big.Int
	// Clone initial value
	value.Set(state.Load(p.Sources[0]))
	// Subtract register values
	for _, src := range p.Sources[1:] {
		value.Sub(&value, state.Load(src))
	}
	// Subtract constant
	value.Sub(&value, &p.Constant)
	// Write value
	state.StoreAcross(value, p.Targets...)
	//
	return state.Pc() + 1
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Sub) Lower(pc uint) micro.Instruction {
	var (
		source    agnostic.Polynomial
		monomials []agnostic.Monomial = make([]agnostic.Monomial, len(p.Sources)+1)
		constant  big.Int
	)
	// Include source terms
	for i, src := range p.Sources {
		//
		if i == 0 {
			monomials[i] = poly.NewMonomial(one, src)
		} else {
			monomials[i] = poly.NewMonomial(minusOne, src)
		}
	}
	// Construct negative constant
	constant.Set(&p.Constant)
	constant.Mul(&constant, &minusOne)
	// Include source constant
	monomials[len(p.Sources)] = poly.NewMonomial[io.RegisterId](constant)
	// Construct source polynomial
	source = source.Set(monomials...)
	//
	code := &micro.Assign{
		Targets: p.Targets,
		Source:  source,
	}
	// Lowering here produces an instruction containing a single microcode.
	return micro.NewInstruction(code, &micro.Jmp{Target: pc + 1})
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Sub) RegistersRead() []io.RegisterId {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Sub) RegistersWritten() []io.RegisterId {
	return p.Targets
}

func (p *Sub) String(fn schema.Module) string {
	return assignmentToString(p.Targets, p.Sources, p.Constant, fn, zero, " - ")
}

// Validate checks whether or not this instruction is correctly balanced.  The
// algorithm here may seem a little odd at first.  It counts the number of
// *unique values* required to hold both the positive and negative components of
// the right-hand side.  This gives the minimum bitwidth required.
func (p *Sub) Validate(fieldWidth uint, fn schema.Module) error {
	var (
		regs     = fn.Registers()
		lhs_bits = sumTargetBits(p.Targets, regs)
		rhs_bits = subSourceBits(p.Sources, p.Constant, regs)
	)
	// check
	if lhs_bits < rhs_bits {
		return fmt.Errorf("bit overflow (%d bits into %d bits)", rhs_bits, lhs_bits)
	} else if rhs_bits > fieldWidth {
		return fmt.Errorf("field overflow (%d bits into %d bit field)", rhs_bits, fieldWidth)
	} else if err := checkSignBit(p.Targets, regs); err != nil {
		return err
	}
	// Finally, ensure unique targets
	return io.CheckTargetRegisters(p.Targets, regs)
}

// the sign bit check is necessary to ensure there is always exactly one sign bit.
func checkSignBit(targets []io.RegisterId, regs []io.Register) error {
	var n = len(targets) - 1
	// Sanity check targets
	if n < 0 {
		return errors.New("malformed assignment")
	}
	// Determine width of sign bit
	signBitWidth := regs[targets[n].Unwrap()].Width
	// Check it is a single bit
	if signBitWidth == 1 {
		return nil
	}
	// Problem, no alignment.
	return fmt.Errorf("missing sign bit (found u%d most significant bits)", signBitWidth)
}

func subSourceBits(sources []io.RegisterId, constant big.Int, regs []io.Register) uint {
	var rhs big.Int = *regs[sources[0].Unwrap()].MaxValue()
	// Now, add negative components
	for _, target := range sources[1:] {
		rhs.Add(&rhs, regs[target.Unwrap()].MaxValue())
	}
	// Include constant (if relevant)
	rhs.Add(&rhs, &constant)
	// lhs must be able to hold both.
	return uint(rhs.BitLen())
}
