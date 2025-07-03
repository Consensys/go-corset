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
package agnostic

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/stack"
)

// Assignment provides a generic notion of an assignment from an arbitrary
// polynomial to a given set of target registers.
type Assignment struct {
	// Target registers with least significant first
	LeftHandSide []sc.RegisterId
	// Right hand side.
	RightHandSide Polynomial
}

// NewAssignment constructs a new assignment with a given Left-Hand Side (LHS)
// and Right-Hand Side (RHS).
func NewAssignment(lhs []sc.RegisterId, rhs Polynomial) Assignment {
	return Assignment{lhs, rhs}
}

// Width determines the minimal field width required to safely evaluate this
// assignment.  Hence, this should not exceed the field bandwidth.  The
// calculation is fairly straightforward: it is simply the maximum width of the
// left-hand and right-hand sides.
func (p *Assignment) Width(env sc.RegisterMapping) uint {
	var (
		// Determine lhs width
		lhs = CombinedWidthOfLimbs(env, p.LeftHandSide...)
		// Determine rhs width
		rhs = WidthOfPolynomial(p.RightHandSide, env.Limbs())
	)
	//
	return max(lhs, rhs)
}

// Split an assignment according to a given field bandwidth.  This creates one
// or more assignments implementing the original which operate safely within the
// given bandwidth.  For example, consider the following assignment where all
// limbs are u8 (i.e. X'0, X'1, Y'0, and Y'1) and b is u1:
//
// b, X'1, X'0 = (2^8*Y'1 + Y'0) + 1
//
// This assignment cannot be safely evaluated within a field bandwidth of 16
// bits (i.e. because the right-hand side could overflow).  This is determined
// by checking the bandwidth against the computed width of the assignment
// (which, in this case, is 17).  Since the computed width exceeds the available
// bandwidth, the assignment needs to split as follows:
//
// b, X'1 = Y'1 + c
//
// c, X'0 = Y'0 + 1
//
// Here, c is an introduced u1 register for holding the "carry" (this is
// analoguous to carry flags as commonly found in CPU architectures).  In
// general, the algorithm can result in temporary registers of arbitrary size
// being introduced.  For example, consider a more complex case (again, u8
// limbs):
//
// X'3,X'2,X'1,X'0 = (2^8*Y'1 + Y'0) * (2^8*Z'1 + Z'0)
//
// Here, the right-hand side expands as follows into the appropriate polynomial
// representation:
//
// X'3,X'2,X'1,X'0 = (2^16*Y'1*Z'1) + (2^8*Y'1*Z'0) + (2^8*Y'0*Z'1) + (Y'0*Z'0)
//
// The difficulty here is that the left- and right-hand sides are somewhat
// "misaligned".  We can attempt to resolve this through large carry registers
// as follows:
//
//	         X'3 = (Y'1*Z'1) + c1
//
//	c1, X'2, X'1 = (Y'1*Z'0) + (Y'0*Z'1) + c0
//
//	     c0, X'0 = (Y'0*Z'0)
//
// Here, c0 and c1 are u8 and u1 carry registers respectively.  Unfortunately,
// this means the middle assignment has a bandwidth requirement of 17bits (which
// still exceeds our original target of 16bits).  Of course, if our bandwidth
// requirement was just slightly larger, then it would fit and we would be done.
//
// For (sub-)assignments which still exceed the bandwidth requirement (such as
// above), we must further split them by introducing additional temporary
// registers.
func (p *Assignment) Split(bandwidth uint, env sc.RegisterMapping) []Assignment {
	var (
		// worklist of remaining assignments
		worklist stack.Stack[Assignment]
		// set of completed assignments
		completed []Assignment
	)
	// Initialise worklist
	worklist.Push(*p)
	// Continue splitting until no assignments outstanding.
	for !worklist.IsEmpty() {
		next := worklist.Pop()
		// further splitting required?
		if next.Width(env) > bandwidth {
			// yes
			worklist.PushAll(next.InnerSplit(bandwidth, env))
		} else {
			// no
			completed = append(completed, next)
		}
	}
	// Done
	return completed
}

// InnerSplit performs one split of the given assignment according to the given
// bandwidth, but does not guarantee that the resulting assignments fit within
// the given bandwidth.  Rather, the resulting assignments may themselves need
// to be split further.
func (p *Assignment) InnerSplit(bandwidth uint, env sc.RegisterMapping) []Assignment {
	var (
		worklist = NewWidthList(p.LeftHandSide, p.RightHandSide, env)
		//
		assignments []Assignment
	)
	//
	for !worklist.IsEmpty() {
		// FIXME: this is broken because it cannot allocate carry flags.
		next := worklist.Next(bandwidth)
		//
		assignments = append(assignments, next)
	}
	//
	return assignments
}

// WidthList is a form of worklist designed for managing the width-oriented
// algorithm needed here.
type WidthList struct {
	// Left-hand side working set
	lhs []sc.RegisterId
	// Right-hand side working set
	rhs []Packet
	// Current widths of the lhs / rhs
	lhsWidth, rhsWidth uint
	// Current bit offset position
	offset uint
	// Next assignment being constructed
	next Assignment
	//
	env sc.RegisterMapping
}

// NewWidthList constructs a new width list.
func NewWidthList(lhs []sc.RegisterId, rhs Polynomial, env sc.RegisterMapping) WidthList {
	return WidthList{
		lhs,
		Packetize(rhs),
		0, 0, 0,
		Assignment{},
		env,
	}
}

// IsEmpty checks whether or not the width list is empty.
func (p *WidthList) IsEmpty() bool {
	// Sanity check.  In theory, this should be unreachable.  In practice, ...
	if len(p.lhs) != len(p.rhs) {
		if len(p.lhs) == 0 || len(p.rhs) == 0 {
			panic("inconsistent widthlist")
		}
	}
	//
	return len(p.lhs) == 0
}

// AdvanceLeft advances the left-hand side bit offset.
func (p *WidthList) AdvanceLeft() {
	//return offset + p.env.Limb(p.lhs[i]).Width
	panic("todo")
}

// AdvanceRight advances the right-hand side bit offset.
func (p *WidthList) AdvanceRight() {
	// FIXME: add width somehow
	// FIXME: normalise contents somehow?
	p.next.RightHandSide.Add(p.rhs[0].Contents)
	panic("todo")
}

// Next returns the next assignment matching the given bitwidth requirement.
func (p *WidthList) Next(bandwidth uint) Assignment {
	// // Reset for next assignment
	// p.next = Assignment{}
	// p.lhsWidth = 0
	// p.rhsWidth = 0
	// //
	// for p.lhsWidth < bandwidth && p.rhsWidth < bandwidth {
	// 	if lBitOffset < rBitOffset {
	// 		// advance left
	// 		p.AdvanceLeft()
	// 	} else if lBitOffset > rBitOffset {
	// 		// advance right
	// 		p.AdvanceRight()
	// 	} else {
	// 		// advance both
	// 		p.AdvanceLeft()
	// 		p.AdvanceRight()
	// 	}
	// }
	// // Update offset position
	// p.offset += p.rhsWidth
	// // FIXME: carry flags!
	// return p.next
	// NOTE: we need to allocate packets upto the bandwidth.  Then, allocate
	// registers accordingly.  Every register allocated has to be completely
	// defined by the given packets, and cannot overlap a subsequent packet. Any
	// discrepancy between registers and packets can be handled with carry
	// registers.
	//
	// SO: any register within the current packet frontier is automatically
	// allocated.  Hence, we advance the frontier whilst there is at least one
	// register which can be allocated as a result.
	panic("todo")
}
