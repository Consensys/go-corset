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
	"fmt"
	"strings"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// Equation provides a generic notion of an equation between two polynomials.
// An equation is in *balanced form* if neither side contains a negative
// coefficient.
type Equation struct {
	// Left hand side.
	LeftHandSide RelativePolynomial
	// Right hand side.
	RightHandSide RelativePolynomial
}

// NewEquation simply constructs a new equation.
func NewEquation(lhs RelativePolynomial, rhs RelativePolynomial) Equation {
	return Equation{lhs, rhs}
}

// Balance an equation means to convert it such that no negative coefficients
// remain. For example, balancing the equation "0 == x - 1" gives "1 == x".  The
// benefit of balancing is simply that it eliminates any requirement for an
// interpretation of signed values.
func (p *Equation) Balance() Equation {
	// Check whether any work to do
	if !p.LeftHandSide.Signed() && !p.RightHandSide.Signed() {
		return *p
	}
	// Yes, work to be done
	var (
		lhsPos, lhsNeg = balancePolynomial(p.LeftHandSide)
		rhsPos, rhsNeg = balancePolynomial(p.RightHandSide)
	)
	// Done
	return NewEquation(lhsPos.Add(rhsNeg), rhsPos.Add(lhsNeg))
}

// Width determines the minimal field width required to safely evaluate this
// assignment.  Hence, this should not exceed the field bandwidth.  The
// calculation is fairly straightforward: it is simply the maximum width of the
// left-hand and right-hand sides.
func (p *Equation) Width(mapping sc.RegisterMap) uint {
	var (
		env = EnvironmentFromMap(mapping)
		// Determine lhs width
		lhs, lSign = WidthOfPolynomial(p.LeftHandSide, env)
		// Determine rhs width
		rhs, rSign = WidthOfPolynomial(p.RightHandSide, env)
	)
	// Sanity check
	if lSign || rSign {
		panic("equation not balanced!")
	}
	//
	return max(lhs, rhs)
}

func (p *Equation) String(mapping sc.RegisterMap) string {
	var builder strings.Builder
	//
	builder.WriteString("[")
	// Write left-hand side
	builder.WriteString(poly2string(p.LeftHandSide, mapping))
	//
	builder.WriteString(" == ")
	// write right-hand side
	builder.WriteString(poly2string(p.RightHandSide, mapping))
	//
	builder.WriteString(fmt.Sprintf("]^%d", p.Width(mapping)))
	//
	return builder.String()
}

// Split an equation according to a given field bandwidth.  This creates one
// or more equations implementing the original which operate safely within the
// given bandwidth.
func (p *Equation) Split(bandwidth uint, env sc.RegisterAllocator) (eqs []Equation) {
	var (
		bp = p.Balance()
	)
	//
	// fmt.Printf("BANDWIDTH: %d\n", bandwidth)
	// fmt.Printf("EQUATION: %s\n", p.String(env))
	// Check whether any splitting required
	if bp.Width(env) > bandwidth {
		// Yes!
		eqs = bp.chunkUp(bandwidth, env)
	} else {
		// Nope
		eqs = []Equation{*p}
	}
	//
	return eqs
}

// Cap all terms within a polynomial to ensure they can be safely evaluated
// within the given bandwidth.  For example, consider the following constraint
// (where both registers are u8):
//
// 0 == X * Y
//
// Suppose a bandwidth of 15bits.  Then, X*Y cannot be safely evaluated since it
// requires 16bits of information.  Instead, we have to break up either X or Y
// into smaller chunks.  Suppose we break X into two 4bit chunks, X'0 and X'1.
// Then we have:
//
// 0 == (256*X'1 + X'0) * Y
//
// --> 0 == 16*X'1*Y + X'0*Y
//
// At this point, each term can be safely evaluated within the given bandwidth
// and this equation can be chunked.  Observe that we assume supplementary
// constraints are included to enforce that X == 16*X'1 + X'0.
//
// The real challenge with this algorithm is, for a polynomial which cannot be
// chunked, to determine which variable(s) to subdivide and by how much.
func (p *Equation) chunkUp(bandwidth uint, mapping sc.RegisterAllocator) []Equation {
	var (
		// Determine the bitwidth of each chunk
		chunkWidths = p.determineChunkBitwidths(mapping)
		lhs, rhs    []RelativePolynomial
		equations   []Equation
	)
	// Divide polynomials into chunks
	for {
		var lhsChunked, rhsChunked bool
		//
		lhs, lhsChunked = chunkPolynomial(p.LeftHandSide, chunkWidths, bandwidth, mapping)
		rhs, rhsChunked = chunkPolynomial(p.RightHandSide, chunkWidths, bandwidth, mapping)
		//
		if lhsChunked && rhsChunked {
			// Successful chunking, therefore finalise any required allocations
			// and continue on to construct final equations.
			break
		}
		//
		if !lhsChunked {
			fmt.Printf("RHS=%s\n", poly2string(lhs[0], mapping))
		}
		//
		if !rhsChunked {
			fmt.Printf("LHS=%s\n", poly2string(rhs[0], mapping))
		}
		// Chunking unsuccessful, therefore start over and attempt another
		// split.
		// Reset any speculative register allocations.
		// Reset any constraints previously considered.
		equations = nil

		panic("got here")
	}
	// Reconstruct equations
	for i := range chunkWidths {
		if lhs[i].Len() > 0 || rhs[i].Len() > 0 {
			equations = append(equations, NewEquation(lhs[i], rhs[i]))
		}
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
func (p *Equation) determineChunkBitwidths(env sc.RegisterAllocator) []uint {
	return p.chunkOnMaximumRegisterWidth(env)
}

// A naive chunking algorithm which chunks according to register boundaries.
// Specifically, the boundaries determined by the largest register in the
// equation.  As such, it may often produce suboptimal chunkings.  However, it
// is very simple to implement and make progress with.
func (p *Equation) chunkOnMaximumRegisterWidth(env sc.RegisterAllocator) []uint {
	var (
		leftLargestWidth  = largestWidth(RegistersRead(p.LeftHandSide), env)
		rightLargestWidth = largestWidth(RegistersRead(p.RightHandSide), env)
		largestWidth      = max(leftLargestWidth, rightLargestWidth)
		bitwidth          = p.Width(env)
		chunks            []uint
	)
	//
	for bitwidth > 0 {
		// Determine how much to take off
		width := min(largestWidth, bitwidth)
		// Update the chunk
		chunks = append(chunks, width)
		bitwidth -= width
	}
	//
	return chunks
}

// Determine largest bitwidth of any register in a given set of registers.
func largestWidth(registers []sc.LimbId, env sc.RegisterAllocator) uint {
	var width = uint(0)

	for _, reg := range registers {
		width = max(width, env.Register(reg).Width)
	}
	//
	return width
}

// Divide a polynomial into "chunks", each of which has a maximum bitwidth as
// determined by the chunk widths.  This inserts carry lines as needed to ensure
// correctness.
func chunkPolynomial(p RelativePolynomial, chunkWidths []uint, bandwidth uint,
	mapping sc.RegisterAllocator) ([]RelativePolynomial, bool) {
	//
	var (
		env    = EnvironmentFromMap(mapping)
		chunks []RelativePolynomial
	)
	// Subdivide polynomial into chunks
	for _, chunkWidth := range chunkWidths {
		var remainder RelativePolynomial
		// Chunk the polynomials
		p, remainder = dividePolynomial(p, chunkWidth)
		// Include remainder as chunk
		chunks = append(chunks, remainder)
	}
	// Add carry lines as necessary
	for i := 0; i < len(chunks); i++ {
		var (
			carry, borrow RelativePolynomial
			ithWidth, _   = WidthOfPolynomial(chunks[i], env)
			chunkWidth    = chunkWidths[i]
		)
		// Calculate overflow from ith chunk (if any)
		if ithWidth > bandwidth {
			// This arises when a given term of the polynomial being chunked
			// cannot be safely evaluated within the given bandwidth (i.e.
			// cannot be evaluated without overflow).  To resolve this
			// situation, we need to further subdivide one or more registers to
			// reduce the maximum bandwidth required for any particular term.
			return []RelativePolynomial{chunks[i]}, false
		} else if (i+1) != len(chunks) && ithWidth > chunkWidth {
			var (
				carryReg   = mapping.Allocate("c", ithWidth-chunkWidth)
				chunkShift = math.Pow2(chunkWidth)
			)
			// Set assignment for filling carry register
			mapping.Assign(carryReg.Id(), chunkWidth, chunks[i])
			// Subtract carry from this chunk
			chunks[i] = chunks[i].Sub(borrow.Set(poly.NewMonomial(*chunkShift, carryReg.Shift(0))))
			// Add carry to next chunk
			chunks[i+1] = chunks[i+1].Add(carry.Set(poly.NewMonomial(one, carryReg.Shift(0))))
		}
	}
	//
	return chunks, true
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
func dividePolynomial(poly RelativePolynomial, n uint) (RelativePolynomial, RelativePolynomial) {
	var (
		quotient, remainder RelativePolynomial
		quotients           []RelativeMonomial
		remainders          []RelativeMonomial
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

// Split a polynomial into its positive and negative components.
func balancePolynomial(poly RelativePolynomial) (pos, neg RelativePolynomial) {
	// Set both sides to zero
	pos = pos.Set()
	neg = neg.Set()
	//
	for i := range poly.Len() {
		var (
			tmp RelativePolynomial
			ith = poly.Term(i)
		)
		//
		tmp = tmp.Set(ith)
		//
		if ith.IsNegative() {
			neg = neg.Sub(tmp)
		} else {
			pos = pos.Add(tmp)
		}
	}
	//
	return pos, neg
}

// Convenient helper function
func poly2string(p RelativePolynomial, env sc.RegisterMap) string {
	return poly.String(p, func(r sc.RelativeRegisterId) string {
		return env.Register(r.Id()).Name
	})
}
