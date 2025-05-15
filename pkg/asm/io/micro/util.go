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
	"math"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
)

func assignmentToString(dsts []uint, srcs []uint, constant big.Int, fn io.Function[Instruction],
	c big.Int, op string) string {
	//
	var (
		builder strings.Builder
		regs    = fn.Registers()
	)
	//
	builder.WriteString(io.RegistersReversedToString(dsts, regs))
	builder.WriteString(" = ")
	//
	for i, r := range srcs {
		if i != 0 {
			builder.WriteString(op)
		}
		//
		if r < uint(len(regs)) {
			builder.WriteString(regs[r].Name)
		} else {
			builder.WriteString(fmt.Sprintf("?%d", r))
		}
	}
	//
	if len(srcs) == 0 || constant.Cmp(&c) != 0 {
		if len(srcs) > 0 {
			builder.WriteString(op)
		}
		//
		builder.WriteString("0x")
		builder.WriteString(constant.Text(16))
	}
	//
	return builder.String()
}

// RegisterSplittingEnvironment represents an environment to assist with register splitting.
// Specifically, it maintains the list of registers as they were before
// splitting, along with the list as they are after splitting and, finally, a
// mapping between them.
type RegisterSplittingEnvironment struct {
	// Maximum permitted width for a register.
	maxWidth uint
	// Set of unsplit registers (i.e. as they were before).
	regsBefore []io.Register
	// Set of split registers (i.e. as they are after splitting).  Observe that
	// this can include extra registers are allocated to implement the split
	// (e.g. for holding carry flags).
	regsAfter []io.Register
	// Mapping from indices in regsBefore to indices in regsAfter
	regMap []uint
}

// NewRegisterSplittingEnvironment constructs a new register splitting environment for a given set
// of registers and a desired maximum register width.
func NewRegisterSplittingEnvironment(maxWidth uint, registers []io.Register) *RegisterSplittingEnvironment {
	var (
		// Mapping from old register ids to new register ids.
		mapping []uint = make([]uint, len(registers))
		//
		splitRegisters []io.Register
	)
	//
	for i, reg := range registers {
		// Map old id to new id
		mapping[i] = uint(len(splitRegisters))
		// Check whether splitting required.
		if reg.Width > maxWidth {
			// Yes!
			splitRegisters = append(splitRegisters, io.SplitRegister(maxWidth, reg)...)
		} else {
			splitRegisters = append(splitRegisters, reg)
		}
	}
	//
	return &RegisterSplittingEnvironment{
		maxWidth,
		registers,
		splitRegisters,
		mapping,
	}
}

// RegistersAfter returns the set of registers as they appear after splitting.
func (p *RegisterSplittingEnvironment) RegistersAfter() []io.Register {
	return p.regsAfter
}

// SplitSourceRegisters splits a given set of source registers into "packets" of
// limbs.  For example, suppose r0 and r1 are source registers of bitwidth
// (respectively) 16bits and 8bits.  Then, splitting for a maximum width of 8
// yields 2 packets: {{r0'0,r1'0}, {r0'1}}
func (p *RegisterSplittingEnvironment) SplitSourceRegisters(sources ...uint) [][]uint {
	ntargets := make([][]uint, io.MaxNumberOfLimbs(p.maxWidth, p.regsBefore, sources))
	//
	for _, target := range sources {
		ntarget := p.regMap[target]
		reg := p.regsBefore[target]
		// Determine split parameters
		n := io.NumberOfLimbs(p.maxWidth, reg.Width)
		// Split up n limbs
		for j := uint(0); j != n; j++ {
			ntargets[j] = append(ntargets[j], ntarget+j)
		}
	}
	//
	return ntargets
}

// SplitTargetRegisters splits a set of registers, e.g. for an assignment.  For
// example, suppose we have:
//
// > b,x,y = ...
//
// Where x,y are 16bit registers and b is a 1bit overflow.  For a maximum
// register width of 8bits, the above is transformed into:
//
// > b,x'1,x'0',y'1,y'0 = ...
//
// And this set of expanded target registers is returned.
func (p *RegisterSplittingEnvironment) SplitTargetRegisters(targets ...uint) []uint {
	var ntargets []uint
	//
	for _, target := range targets {
		ntarget := p.regMap[target]
		reg := p.regsBefore[target]
		n := io.NumberOfLimbs(p.maxWidth, reg.Width)
		// Split into n limbs
		for j := uint(0); j != n; j++ {
			ntargets = append(ntargets, ntarget+j)
		}
	}
	//
	return ntargets
}

// AllocateTargetLimbs allocates upto maxWidth bits from a given set of target
// limbs.
func (p *RegisterSplittingEnvironment) AllocateTargetLimbs(targetLimbs []uint) (uint, []uint, []uint) {
	var (
		width   = uint(0)
		targets []uint
	)
	// Allocate targets from first packet
	for width < p.maxWidth && len(targetLimbs) > 0 {
		target := targetLimbs[0]
		targets = append(targets, target)
		targetLimbs = targetLimbs[1:]
		width = width + p.regsAfter[target].Width
	}
	// Sanity  check
	if width > p.maxWidth {
		panic("mis-aligned target registers")
	}
	//
	return width, targets, targetLimbs
}

// AllocateCarryRegister allocates a carry flag to hold bits which "overflow" the
// left-hand side of an assignment (i.e. where sourceWidth is greater than
// targetWidth).
func (p *RegisterSplittingEnvironment) AllocateCarryRegister(targetWidth uint, sourceWidth uint) uint {
	var (
		overflowRegId = uint(len(p.regsAfter))
	)
	// Sanity check
	if targetWidth > sourceWidth {
		// should be
		panic(fmt.Sprintf("unreachable (target width %d vs source width %d)", targetWidth, sourceWidth))
	} else if targetWidth == sourceWidth {
		// Indicates carry flag not required (e.g. because no carry in lower
		// portion of addition).
		return math.MaxUint
	}
	// Determine number of bits of overflow
	overflowWidth := sourceWidth - targetWidth
	// Construct register for holding overflow
	overflowRegister := io.Register{Name: fmt.Sprintf("c$%d", overflowRegId), Kind: io.TEMP_REGISTER, Width: overflowWidth}
	// Allocate overflow register
	p.regsAfter = append(p.regsAfter, overflowRegister)
	//
	return overflowRegId
}
