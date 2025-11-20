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

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// Assignment2 provides a generic notion of an assignment from an arbitrary
// polynomial to a given set of target registers.
type Assignment2 struct {
	// Target registers with least significant first
	LeftHandSide []register.Id
	// Right hand side.
	RightHandSide StaticPolynomial
}

// NewAssignment2 constructs a new assignment with a given Left-Hand Side (LHS)
// and Right-Hand Side (RHS).
func NewAssignment2(lhs []register.Id, rhs StaticPolynomial) Assignment2 {
	// Sanity check
	if rhs == nil {
		panic("malformed assignment")
	}
	//
	return Assignment2{lhs, rhs}
}

func (p *Assignment2) String(env register.Map) string {
	var builder strings.Builder
	//
	builder.WriteString("[")
	//
	for i := len(p.LeftHandSide); i > 0; {
		if i != len(p.LeftHandSide) {
			builder.WriteString(",")
		}

		i = i - 1

		builder.WriteString(env.Register(p.LeftHandSide[i]).Name)
	}
	//
	builder.WriteString(" := ")
	builder.WriteString(StaticPoly2String(p.RightHandSide, env))
	//
	builder.WriteString(fmt.Sprintf("]^%d", p.Width(env)))
	//
	return builder.String()
}

// Width determines the minimal field width required to safely evaluate this
// assignment.  Hence, this should not exceed the field bandwidth.  The
// calculation is fairly straightforward: it is simply the maximum width of the
// left-hand and right-hand sides.
func (p *Assignment2) Width(env register.Map) uint {
	var (
		// Determine lhs width
		lhs = CombinedWidthOfRegisters(env, p.LeftHandSide...)
		// Determine rhs width
		rhs, _ = WidthOfPolynomial(p.RightHandSide, StaticEnvironment(env))
	)
	//
	return max(lhs, rhs)
}

// Split an equation according to a given field bandwidth.  This creates one
// or more equations implementing the original which operate safely within the
// given bandwidth.
func (p *Assignment2) Split(field field.Config, env RegisterAllocator) (eqs []Assignment2) {
	// Check whether any splitting required
	if p.Width(env) > field.BandWidth {
		// Yes!
		eqs = p.chunkUp(field, env)
	} else {
		// Nope
		eqs = []Assignment2{*p}
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
func (p *Assignment2) chunkUp(field field.Config, mapping RegisterAllocator) []Assignment2 {
	var (
		lhs [][]register.Id
		// Determine the bitwidth of each chunk
		rhs []StaticPolynomial
		// Equations being constructed
		assignments []Assignment2
		// Current chunk width
		chunkWidth = field.BandWidth
	)
	// Attempt to divide polynomials into chunks.  If this fails, iterative
	// decrease chunk width until something fits.
	for {
		var (
			rhsChunked  bool
			chunkWidths []uint
			// FIXME: split non-linea terms
			right = p.RightHandSide
		)
		// Calculate actual chunk widths based on current chunk width.
		chunkWidths, lhs = determineChunkBitwidths(p.LeftHandSide, chunkWidth, mapping)
		//
		rhs, rhsChunked = chunkAssignedPolynomial(right, chunkWidths, field, mapping)
		//
		if rhsChunked {
			// Successful chunking, therefore include any constraints necessary
			// for splitting of non-linear terms and construct final equations.
			break
		}
		//
		panic("todo: split non-linea terms")
	}
	// Reconstruct equations
	for i := range len(lhs) {
		if len(lhs[i]) > 0 || rhs[i].Len() > 0 {
			assignments = append(assignments, NewAssignment2(lhs[i], rhs[i]))
		}
	}
	// Done
	return assignments
}

func determineChunkBitwidths(regs []register.Id, chunkWidth uint, mapping register.Map) ([]uint, [][]register.Id) {
	var (
		chunks      [][]register.Id
		chunkWidths []uint
	)
	//
	for len(regs) != 0 {
		var (
			width uint
			chunk []register.Id
		)
		// Determine next chunkd
		width, chunk, regs = getNextChunk(regs, chunkWidth, mapping)
		chunks = append(chunks, chunk)
		chunkWidths = append(chunkWidths, width)
	}
	//
	return chunkWidths, chunks
}

func getNextChunk(regs []register.Id, chunkWidth uint, mapping register.Map) (uint, []register.Id, []register.Id) {
	var bitwidth uint
	//
	for i, r := range regs {
		reg := mapping.Register(r)
		//
		if bitwidth+reg.Width > chunkWidth {
			return bitwidth, regs[:i], regs[i:]
		}
		//
		bitwidth += reg.Width
	}
	//
	return bitwidth, regs, nil
}

// Divide a polynomial into "chunks", each of which has a maximum bitwidth as
// determined by the chunk widths.  This inserts carry lines as needed to ensure
// correctness.
func chunkAssignedPolynomial(p StaticPolynomial, chunkWidths []uint, field field.Config,
	mapping RegisterAllocator) ([]StaticPolynomial, bool) {
	//
	var (
		env    = StaticEnvironment(mapping)
		chunks []StaticPolynomial
	)
	// Subdivide polynomial into chunks
	for _, chunkWidth := range chunkWidths {
		var remainder StaticPolynomial
		//
		fmt.Printf("CHUNK WIDTH %d\n", chunkWidth)
		// Chunk the polynomials
		p, remainder = p.Shr(chunkWidth)
		// Include remainder as chunk
		chunks = append(chunks, remainder)
	}
	// Add carry lines as necessary
	for i := 0; i < len(chunks); i++ {
		var (
			carry, borrow StaticPolynomial
			ithWidth, _   = WidthOfPolynomial(chunks[i], env)
			chunkWidth    = chunkWidths[i]
		)
		//
		fmt.Printf("CHUNK[%d]=%s\n", i, StaticPoly2String(chunks[i], mapping))
		// Calculate overflow from ith chunk (if any)
		if ithWidth > field.BandWidth {
			fmt.Printf("Failed chunking %d > %d\n", ithWidth, field.BandWidth)
			// This arises when a given term of the polynomial being chunked
			// cannot be safely evaluated within the given bandwidth (i.e.
			// cannot be evaluated without overflow).  To resolve this
			// situation, we need to further subdivide one or more registers to
			// reduce the maximum bandwidth required for any particular term.
			return []StaticPolynomial{chunks[i]}, false
		} else if (i+1) != len(chunks) && ithWidth > chunkWidth {
			var (
				// Determine width of carry register
				carryWidth = ithWidth - chunkWidth
				// Allocate carry register
				carryReg = mapping.Allocate("c", carryWidth)
				// Calculate amount to shift carry
				chunkShift = util_math.Pow2(chunkWidth)
			)
			// FIXME: missing carry assignment

			// Subtract carry from this chunk
			chunks[i] = chunks[i].Sub(borrow.Set(poly.NewMonomial(*chunkShift, carryReg)))
			// Add carry to next chunk
			chunks[i+1] = chunks[i+1].Add(carry.Set(poly.NewMonomial(one, carryReg)))
		}
	}
	//
	return chunks, true
}

// StaticPoly2String provides a convenient helper function for debugging polynomials.
func StaticPoly2String(p StaticPolynomial, env register.Map) string {
	return poly.String(p, func(r register.Id) string {
		return env.Register(r.Id()).Name
	})
}
