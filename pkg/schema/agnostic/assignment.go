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
	"math/big"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/stack"
	"github.com/consensys/go-corset/pkg/util/poly"
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
			worklist.PushAll(next.innerSplit(bandwidth, env))
		} else {
			// no
			completed = append(completed, next)
		}
	}
	// Done
	return completed
}

// innerSplit performs one split of the given assignment according to the given
// bandwidth, but does not guarantee that the resulting assignments fit within
// the given bandwidth.  Rather, the resulting assignments may themselves need
// to be split further.
func (p *Assignment) innerSplit(bandwidth uint, env sc.RegisterMapping) []Assignment {
	var assignments []Assignment = p.initialiseSplit(env)
	// Merge to exploit available bandwidth.
	assignments = coalesceAssignments(assignments, bandwidth, env)
	// Add carry registers as needed
	return assignments
}

// InitialiseSplit constructs an initial set of assignments with one for each
// target register.  Here, each assignment includes all monomials from the
// right-hand side whose coefficient begins within the range of the
// corresponding target register.  For example, consider the following
// assignment:
//
// b, y'1, y'0 = 2^8*x'1 + x'0 + 1
//
// would be initialised as follows:
//
// y'0 := x'0 + 1 ; y'1 := x'1 ; b := 0
//
// Observe that b has no initial assignment.  This is because, in the end, it
// will form part of the previous assignment (i.e. after grouping).
//
// One mildly complicating factor is that of "dividing monomials".  Consider our
// example, above where the term 2^8*x'1 becomes just x'1.  This makes sense as,
// in the final assignment "y'1 := x'1", we know that y'1 start at bit offset 8.
// Thus, we can see that "y'1 := 2^8*x'1" doesn't make sense.  Thus, to
// determine the right coefficients, a division process is employed.  Since
// division may not be exact, remainders are left to be processed again and a
// worklist is used to manage those bits that still need processing.
func (p *Assignment) initialiseSplit(env sc.RegisterMapping) []Assignment {
	var (
		// Final list of assignments to be constructed
		monomials = make([][]Monomial, len(p.LeftHandSide))
		// Worklist contains list of monomials being processed.
		worklist stack.Stack[Monomial]
		// Assignments to be constructed
		assignments = make([]Assignment, len(p.LeftHandSide))
	)
	// Initialise worklist from source polynomial
	for j := range p.RightHandSide.Len() {
		worklist.Push(p.RightHandSide.Term(j))
	}
	// Continue processing monomials until no more remain.
	for !worklist.IsEmpty() {
		// Extract next item to process
		next := worklist.Pop()
		// Identify target register
		i, offset := identifyEnclosingRegister(p.LeftHandSide, next.Coefficient(), env)
		// Divide monomial by bit offset
		next, rest := divideMonomial(next, offset)
		// Check wether division was exact
		if !rest.IsZero() {
			// No, therefore some remainder must still be processed
			worklist.Push(rest)
		}
		//
		monomials[i] = append(monomials[i], next)
	}
	// Finally construct assignments
	for i, lid := range p.LeftHandSide {
		var tmp Polynomial
		// Construct ith assignment
		assignments[i] = Assignment{
			// left-hand side
			[]sc.RegisterId{lid},
			// right-hand side
			tmp.Set(monomials[i]...),
		}
	}
	// Done
	return assignments
}

// Identify enclosing register determines the index into the given regs array of
// the register whose bitrange encloses the given value.  This also returns the
// starting (bit) offset of this register.
func identifyEnclosingRegister(regs []sc.RegisterId, value big.Int, env sc.RegisterMapping) (uint, uint) {
	var bitOffset uint
	//
	for i, lid := range regs {
		var limb = env.Limb(lid)
		// Value contained by this register?
		if withinBitRange(value, bitOffset, bitOffset+limb.Width) {
			// Yes!
			return uint(i), bitOffset
		}
		// Shift offset along
		bitOffset += limb.Width
	}
	// It should not be possible to get here if the original assignments were
	// well-formed.
	panic("unreachable")
}

// CoalesceAssignments attempts to merge consecutive assignments to exploit the
// available bandwidth as much as possible.  For example, consider the example
// initial splitting obtained above:
//
// [y'0 := x'0 + 1]^9 ; [y'1 := x'1]^8 ; [b := 0]^1
//
// Here, the width of each assignment is given as a superscript.  Roughly
// speaking, any two assignments can be safely merged when the combined width
// remains within the target bandwidth.  For our example, we merge the last two
// assignments as follows:
//
// [y'0 := x'0 + 1]^9 ; [b, y'1 := x'1]^9
//
// This is permitted because the combined assignment still meets the necessary
// bandwidth requirements.
//
// The process of combining assignments is not as simple as outlined above. This
// is due to the need for carry registers which, at this point, have not yet
// been added.  When a carry register is required, the effective width of an
// assignment may increase.  To manage this, the given algorithm allocates carry
// registers as it proceeds, thereby allowing for accurate width calculations.
//
// NOTE: this implementation does not attempt to find an optimal allocation of
// assignments (as this may indeed be a hard computational problem).  Instead,
// assignments are merged greedily starting from the least significant position.
func coalesceAssignments(assignments []Assignment, bandwidth uint, env sc.RegisterMapping) []Assignment {
	// TODO: implement algorithm
	return assignments
}

// DividingMonomial divides a given monomial m by some value n.  The division
// maybe exact, in which case the remainder will be zero.  For example, dividing
// 7x by 3 gives 2x (val) + x (rem).
func divideMonomial(m Monomial, n uint) (val Monomial, rem Monomial) {
	var (
		coeff     = m.Coefficient()
		nb        = big.NewInt(2)
		quotient  big.Int
		remainder big.Int
	)
	// sanity check division by zero!
	if n == 0 {
		return m, rem
	}
	// Determine 2^n
	nb.Exp(nb, big.NewInt(int64(n)), nil)
	// Determine quotient and remainder
	quotient.Div(&coeff, nb)
	remainder.Mod(&coeff, nb)
	// Done
	return poly.NewMonomial(quotient, m.Vars()...),
		poly.NewMonomial(remainder, m.Vars()...)
}

// withinBitRange checks whether a given integer value it contained within a
// given bit range [s,e).  For example, 123 is contained within the range 0..8,
// but 256 is not.
func withinBitRange(val big.Int, start, end uint) bool {
	var (
		s = big.NewInt(2)
		e = big.NewInt(2)
	)
	//
	s.Exp(s, big.NewInt(int64(start)), nil)
	e.Exp(e, big.NewInt(int64(end)), nil)
	// Check interval
	return val.Cmp(s) >= 0 && val.Cmp(e) < 0
}
