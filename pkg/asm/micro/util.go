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

	"github.com/consensys/go-corset/pkg/asm/insn"
)

func assignmentToString(dsts []uint, srcs []uint, constant big.Int, regs []Register, c big.Int, op string) string {
	var (
		builder strings.Builder
		n       = len(dsts) - 1
	)
	//
	for i := 0; i != len(dsts); i++ {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(regs[dsts[n-i]].Name)
	}
	//
	builder.WriteString(" = ")
	//
	for i, r := range srcs {
		if i != 0 {
			builder.WriteString(op)
		}
		//
		builder.WriteString(regs[r].Name)
	}
	//
	if len(srcs) == 0 || constant.Cmp(&c) != 0 {
		if len(srcs) > 0 {
			builder.WriteString(op)
		}
		//
		builder.WriteString(constant.String())
	}
	//
	return builder.String()
}

// MaxNumberOfLimbs returns the maximum number of limbs required for any
// register in the given target registers.  For example, given registers r0 and
// r1 of bitwidths 16bits and 8bits (respectively), then 2 is maximum number of
// limbs for an 8bit maximum register width.
func MaxNumberOfLimbs(maxWidth uint, regs []Register, targets []uint) uint {
	var n = uint(0)
	//
	for _, target := range targets {
		regWidth := regs[target].Width
		n = max(n, NumberOfLimbs(maxWidth, regWidth))
	}
	//
	return n
}

// NumberOfLimbs determines the number of limbs required for a given bitwidth.
// For example, a 64bit register splits into two limbs for a maximum register
// width of 32bits. Observe that an e.g. 60bit register also splits into two
// limbs here as well, where the most significant limb is 28bits wide and the
// least significant is 32bits width.
func NumberOfLimbs(maxWidth uint, regWidth uint) uint {
	n := regWidth / maxWidth
	m := regWidth % maxWidth
	// Check for uneven split
	if m != 0 {
		return n + 1
	}
	//
	return n
}

// RegisterSplittingEnvironment represents an environment to assist with register splitting.
// Specifically, it maintains the list of registers as they were before
// splitting, along with the list as they are after splitting and, finally, a
// mapping between them.
type RegisterSplittingEnvironment struct {
	// Maximum permitted width for a register.
	maxWidth uint
	// Set of unsplit registers (i.e. as they were before).
	regsBefore []Register
	// Set of split registers (i.e. as they are after splitting).  Observe that
	// this can include extra registers are allocated to implement the split
	// (e.g. for holding carry flags).
	regsAfter []Register
	// Mapping from indices in regsBefore to indices in regsAfter
	regMap []uint
}

// NewRegisterSplittingEnvironment constructs a new register splitting environment for a given set
// of registers and a desired maximum register width.
func NewRegisterSplittingEnvironment(maxWidth uint, registers []Register) *RegisterSplittingEnvironment {
	var (
		// Mapping from old register ids to new register ids.
		mapping []uint = make([]uint, len(registers))
		//
		splitRegisters []Register
	)
	//
	for i, reg := range registers {
		// Map old id to new id
		mapping[i] = uint(len(splitRegisters))
		// Check whether splitting required.
		if reg.Width > maxWidth {
			// Yes!
			splitRegisters = append(splitRegisters, splitRegister(maxWidth, reg)...)
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
func (p *RegisterSplittingEnvironment) RegistersAfter() []Register {
	return p.regsAfter
}

// SplitConstant splits a given constant into a number of "limbs" of a given
// maximum width. For example, consider splitting the constant 0x7b2d into 8bit
// limbs.  Then, this function returns the array [0x2d,0x7b].
func (p *RegisterSplittingEnvironment) SplitConstant(constant big.Int, nLimbs uint) []big.Int {
	var (
		bound = big.NewInt(2)
		limb  big.Int
		limbs []big.Int = make([]big.Int, nLimbs)
	)
	// Determine upper bound
	bound.Exp(bound, big.NewInt(int64(p.maxWidth)), nil)
	//
	for i := 0; constant.Cmp(&zero) != 0; i++ {
		limb.Mod(&constant, bound)
		limbs[i] = limb

		constant.Rsh(&constant, p.maxWidth)
	}
	//
	return limbs
}

// SplitConstantVariable splits a given constant into a given set of (variable)
// bitwidths.  For example, splitting 0x107 into 8bit and 1bit limbs gives
// [0x07,0x1].
func (p *RegisterSplittingEnvironment) SplitConstantVariable(constant big.Int, limbWidths ...uint) []big.Int {
	var (
		limb  big.Int
		limbs []big.Int = make([]big.Int, len(limbWidths))
	)
	//
	for i := 0; constant.Cmp(&zero) != 0; i++ {
		var (
			bound     = big.NewInt(2)
			limbWidth = limbWidths[i]
		)
		// Determine upper bound
		bound.Exp(bound, big.NewInt(int64(limbWidth)), nil)
		//
		limb.Mod(&constant, bound)
		limbs[i] = limb

		constant.Rsh(&constant, limbWidth)
	}
	//
	return limbs
}

// SplitSourceRegisters splits a given set of source registers into "packets" of
// limbs.  For example, suppose r0 and r1 are source registers of bitwidth
// (respectively) 16bits and 8bits.  Then, splitting for a maximum width of 8
// yields 2 packets: {{r0'0,r1'0}, {r0'1}}
func (p *RegisterSplittingEnvironment) SplitSourceRegisters(sources ...uint) [][]uint {
	ntargets := make([][]uint, MaxNumberOfLimbs(p.maxWidth, p.regsBefore, sources))
	//
	for _, target := range sources {
		ntarget := p.regMap[target]
		reg := p.regsBefore[target]
		// Determine split parameters
		n := NumberOfLimbs(p.maxWidth, reg.Width)
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
		n := NumberOfLimbs(p.maxWidth, reg.Width)
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
	overflowRegister := Register{Name: fmt.Sprintf("c$%d", overflowRegId), Kind: insn.TEMP_REGISTER, Width: overflowWidth}
	// Allocate overflow register
	p.regsAfter = append(p.regsAfter, overflowRegister)
	//
	return overflowRegId
}

// Split a register into a number of limbs with the given maximum bitwidth.  For
// the resulting array, the least significant register is first.  Since
// registers are always split to the maximum width as much as possible, it is
// only the most significant register which may (in some cases) have fewer bits
// than the maximum allowed.
func splitRegister(maxWidth uint, r Register) []Register {
	var (
		nlimbs = NumberOfLimbs(maxWidth, r.Width)
		limbs  = make([]Register, nlimbs)
		width  = r.Width
	)
	//
	for i := uint(0); i < nlimbs; i++ {
		ith_name := fmt.Sprintf("%s'%d", r.Name, i)
		ith_width := min(maxWidth, width)
		limbs[i] = Register{Name: ith_name, Kind: r.Kind, Width: ith_width}
		width -= maxWidth
	}
	//
	return limbs
}
