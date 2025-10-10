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
	"cmp"
	"fmt"
	"slices"
	"strings"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/stack"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// Equation provides a generic notion of an equation between two polynomials.
// An equation is in *balanced form* if neither side contains a negative
// coefficient.
type Equation struct {
	// Left hand side.
	LeftHandSide Polynomial
	// Right hand side.
	RightHandSide Polynomial
}

// NewEquation simply constructs a new equation.
func NewEquation(lhs Polynomial, rhs Polynomial) Equation {
	return Equation{lhs, rhs}
}

// Width determines the minimal field width required to safely evaluate this
// assignment.  Hence, this should not exceed the field bandwidth.  The
// calculation is fairly straightforward: it is simply the maximum width of the
// left-hand and right-hand sides.
func (p *Equation) Width(env sc.RegisterLimbsMap) uint {
	var (
		// Determine lhs width
		lhs, lSign = WidthOfPolynomial(p.LeftHandSide, env.Limbs())
		// Determine rhs width
		rhs, rSign = WidthOfPolynomial(p.RightHandSide, env.Limbs())
	)
	// Sanity check
	if lSign || rSign {
		panic("equation not balanced!")
	}
	//
	return max(lhs, rhs)
}

func (p *Equation) String(env sc.RegisterLimbsMap) string {
	var builder strings.Builder
	//
	builder.WriteString("[")
	// Write left-hand side
	builder.WriteString(poly.String(p.LeftHandSide, func(id sc.RegisterId) string {
		return env.Limb(id).Name
	}))
	//
	builder.WriteString(" == ")
	// write right-hand side
	builder.WriteString(poly.String(p.RightHandSide, func(id sc.RegisterId) string {
		return env.Limb(id).Name
	}))
	//
	builder.WriteString(fmt.Sprintf("]^%d", p.Width(env)))
	//
	return builder.String()
}

// Split an equation according to a given field bandwidth.  This creates one
// or more equations implementing the original which operate safely within the
// given bandwidth.
func (p *Equation) Split(env sc.RegisterAllocator) []Equation {
	var (
		bandwidth = env.Field().FieldBandWidth
		// worklist of remaining equations
		worklist stack.Stack[Equation]
		// set of completed equations
		completed []Equation
	)
	// Initialise worklist
	worklist.Push(*p)
	// Continue splitting until no assignments outstanding.
	for !worklist.IsEmpty() {
		next := worklist.Pop()
		// further splitting required?
		if next.Width(env) > bandwidth {
			// yes
			worklist.PushReversed(next.innerSplit(env))
		} else {
			// no
			completed = append(completed, next)
		}
	}
	// Done
	return completed
}

func (p *Equation) innerSplit(env sc.RegisterAllocator) []Equation {
	var (
		// Determine the bitwidth of each chunk
		chunkWidths = p.determineChunkBitwidths()
		// Sort both sides in order of their coefficients.
		lhs       = p.LeftHandSide
		rhs       = p.RightHandSide
		equations []Equation
	)
	//
	for _, chunkWidth := range chunkWidths {
		var l, r Polynomial
		// Chunk the polynomials
		lhs, l = p.dividePolynomial(lhs, chunkWidth)
		rhs, r = p.dividePolynomial(rhs, chunkWidth)
		// Append new equation
		equations = append(equations, NewEquation(l, r))
	}
	// Done
	return equations
}

// Determine the width of individual chunks used to split the equation.  In
// theory, arbitrary chunk widths can be used provided the total bitwidth
// encloses both sides (i.e. contains all possible value for each side).  In
// practice, the chunking used can affect the overall efficiency of the
// splitting.  As an example consider the following simple equation, where x and
// y are u16:
//
//	x == y + 1
//
// Assuming a desired register width of u8, the derived equation is:
//
//	256*x1 + x0 == 256*y1 + y0 + 1
//
// At this point, we can compare the two sides as follows:
//
//	 15             8 7               0
//	+----------------+-----------------+
//	|     2^8*x1     |        x0       |
//	+----------------+-----------------+
//	+----------------+
//	|     2^8*y1     |
//	+----------------+
//	               +----------------+
//	               |     y0 + 1     |
//	               +----------------+
//
// In this example, a good chunking would be to divide into two u8 chunks.  This
// works well since 3/4 of our boxes are byte aligned already.
//
// In general, chunks do not have to have the same size (even though it did make
// sense above).  In particular, the most significant chunk is often a different
// size.
func (p *Equation) determineChunkBitwidths() []uint {
	panic("todo")
}

// For a given bitwidth n, divide a polynomial by 2^n produces a quotient and
// remainder.  For example, dividing 256*x1+x0 by 2^8 gives x1 remainder x0.
// This algorithm is somehow akin to "shifting" a polynomial downwards.  For
// example, consider our example again:
//
//	 15             8 7               0
//	+----------------+-----------------+
//	|     2^8*x1     |        x0       |
//	+----------------+-----------------+
//
// Then, shifting this down by 8bits gives:
//
//	                  7               0
//	                 +-----------------+
//	>>>>>>>>>>>>>>>> |        x1       |
//	                 +-----------------+
//
// And we are left with a remainder as well.
func (p *Equation) dividePolynomial(poly Polynomial, n uint) (Polynomial, Polynomial) {
	var (
		quotient, remainder Polynomial
		quotients           []Monomial
		remainders          []Monomial
	)
	//
	for i := range poly.Len() {
		quot, rem := divideMonomial(poly.Term(i), n)
		//
		quotients = append(quotients, quot)
		remainders = append(remainders, rem)
	}
	//
	return quotient.Set(quotients...), remainder.Set(remainders...)
}

// Allocate a chunk of continguous bits from the given set of monomials upto the
// specified bandwidth (i.e. allocated chunk should not exceed this amount).
// For example, consider the monomials 2^16*x + 2^8*x + y (where x and y are u8):
//
// +----------------+--------+--------+
// |     2^16*x     |  2^8*x |    y   |
// +----------------+--------+--------+
//
//	23            16 15     8 7      0
//
// Suppose we want to grab upto 20bits.  Then, this should return the mononmials
// [y,2^8*x] (in that order), and a bitwidth of 16 (i.e. since that is the how
// many bits the returned monomials occupy).
func (p *Equation) allocBandwidthBits(terms []BitSlice, offset uint, env sc.RegisterAllocator) (index uint, bits uint) {
	var bandwidth = env.Field().FieldBandWidth
	//
	for i, t := range terms {
		delta := (t.start - offset)
		fmt.Printf("(i=%d) %d+%d >= %d ... ", i, delta, t.bitwidth, bandwidth)
		//
		if delta+t.bitwidth >= bandwidth {
			fmt.Printf("[YES]\n")
			return uint(i), offset
		}
		//
		fmt.Printf("[NO]\n")
	}
	//
	fmt.Printf("[DONE]\n")
	// If we get here, then we've allocated everything that remains.
	return uint(len(terms)), offset
}

// Sort the monomials in a given polynomial by their coefficient.
func sortByCoefficient(poly Polynomial, env sc.RegisterAllocator) []BitSlice {
	var monomials = make([]BitSlice, poly.Len())
	// Extract them
	for i := range poly.Len() {
		monomials[i] = newBitSlice(poly.Term(i), env.LimbsMap())
	}
	// Sort them
	slices.SortFunc(monomials, func(l, r BitSlice) int {
		if c := cmp.Compare(l.start, r.start); c != 0 {
			return c
		}
		// Compare variables second
		return cmp.Compare(l.bitwidth, r.bitwidth)
	})
	// Done
	return monomials
}

// BitSlice represents a region of values determined by a given monomial.
type BitSlice struct {
	// least significant bit of this slice
	start uint
	// bitwidth of this slice.
	bitwidth uint
	// contents of the slice
	body Monomial
}

func newBitSlice(m Monomial, env sc.RegisterMap) BitSlice {
	var (
		coeff         = m.Coefficient()
		start         = uint(coeff.BitLen()) - 1
		bitwidth uint = uint(1)
	)
	//
	for i := range m.Len() {
		// Determine bitwidth of the given register
		ithBitWidth := env.Register(m.Nth(i)).Width
		//
		bitwidth = bitwidth * ithBitWidth
	}
	// Done
	return BitSlice{start, bitwidth, m}
}

func slicesToPolynomial(offset uint, slices []BitSlice) Polynomial {
	var (
		terms = make([]Monomial, len(slices))
		poly  Polynomial
	)
	//
	for i, bits := range slices {
		div, rem := divideMonomial(bits.body, bits.start-offset)
		//
		if !rem.IsZero() {
			panic("handle remainder?")
		}
		// Add to poly
		terms[i] = div
	}
	// Construct polynomial
	return poly.Set(terms...)
}
