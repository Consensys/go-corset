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
package micro

import (
	"fmt"
	"math/big"
	"slices"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
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

// Clone this micro code.
func (p *Add) Clone() Code {
	var constant big.Int
	//
	constant.Set(&p.Constant)
	//
	return &Add{
		slices.Clone(p.Targets),
		slices.Clone(p.Sources),
		constant,
	}
}

// MicroExecute a given micro-code, using a given state.  This may update the
// register values, and returns either the number of micro-codes to "skip over"
// when executing the enclosing instruction or, if skip==0, a destination
// program counter (which can signal return of enclosing function).
func (p *Add) MicroExecute(state io.State) (uint, uint) {
	var value big.Int
	// Add constant
	value.Set(&p.Constant)
	// Add register values
	for _, src := range p.Sources {
		value.Add(&value, state.Load(src))
	}
	// Write value
	state.StoreAcross(value, p.Targets...)
	//
	return 1, 0
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

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.  Here, regsBefore
// represents the registers are they are for this code, whilst regsAfter
// represent those for the resulting split codes.  The regMap provides a
// mapping from registers in regsBefore to those in regsAfter.
//
// Split up target registers according to the given maximum width.  There is a
// problem if the resulting targets are not aligned with respect to the maximum
// width.  For example, consider (where x,y,z are 16bit registers and b a 1 bit
// register):
//
// > b, x := y + z + 1
//
// Then, splitting to a maximum register width of 8bits yields the following:
//
// > b,x1,x0 := (256*y1+y0) + (256*z1+z0) + 1
//
// This is then factored as such:
//
// > b,x1,x0 := 256*(y1+z1) + (y0+z0+1)
//
// Thus, y0+z0+1 define all of the bits for x0 and some of the bits for x1.
func (p *Add) Split(env io.SplittingEnvironment) []Code {
	//
	if len(p.Sources) == 0 {
		// Actually just an assignment, so easy.
		return p.splitAssignment(env)
	} else {
		// var (
		// 	ncodes        []Code
		// 	targetLimbs                 = env.SplitTargetRegisters(p.Targets...)
		// 	sourcePackets               = env.SplitSourceRegisters(p.Sources...)
		// 	constantLimbs               = agnosticity.SplitConstant(uint(len(sourcePackets)), env.MaxWidth(), p.Constant)
		// 	carry         io.RegisterId = schema.NewUnusedRegisterId()
		// )
		// // Allocate all source packets
		// for i, pkt := range sourcePackets {
		// 	var (
		// 		targets     []io.RegisterId
		// 		targetWidth uint
		// 	)
		// 	//
		// 	targetWidth, targets, targetLimbs = env.AllocateTargetLimbs(targetLimbs)
		// 	//
		// 	if i != 0 && carry.IsUsed() {
		// 		// Include carry from previous round
		// 		pkt = append(pkt, carry)
		// 	}
		// 	// Allocate carry flag (if applicable).
		// 	if i+1 != len(sourcePackets) {
		// 		sourceWidth := sumSourceBits(p.Sources, constantLimbs[i], env.RegistersAfter())
		// 		carry = env.AllocateCarryRegister(targetWidth, sourceWidth)
		// 		//
		// 		if carry.IsUsed() {
		// 			targets = append(targets, carry)
		// 		}
		// 	} else {
		// 		// Allocate all outstanding limbs for final packet.
		// 		targets = append(targets, targetLimbs...)
		// 	}
		// 	// Construct split micro code
		// 	code := &Add{targets, pkt, constantLimbs[i]}
		// 	// Done
		// 	ncodes = append(ncodes, code)
		// }
		// //
		// return ncodes
		return []Code{p}
	}
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

func (p *Add) splitAssignment(env io.SplittingEnvironment) []Code {
	var (
		ncodes []Code
		// map target registers into corresponding limbs
		targetLimbs = agnostic.ApplyMapping(env, p.Targets)
		// extract width of each limb
		targetLimbWidths = agnostic.LimbWidths(env, targetLimbs)
		// split constant according to given limb widths
		constantLimbs = agnostic.SplitConstant(p.Constant, targetLimbWidths...)
	)
	//
	for i, target := range targetLimbs {
		code := &Add{Targets: []io.RegisterId{target}, Sources: nil, Constant: constantLimbs[i]}
		ncodes = append(ncodes, code)
	}
	//
	return ncodes
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
