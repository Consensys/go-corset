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
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
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
		// Record initial number of registers
		n = uint(len(mapping.Registers()))
		//
		divisions = initialiseVariableDivisions(n)
		// Determine the bitwidth of each chunk
		rhsChunks []RhsChunk
		// Equations being constructed
		assignments []Assignment2
		//
		lhsChunks = determineLhsChunks(p.LeftHandSide, field.RegisterWidth, mapping)
	)
	// Attempt to divide polynomials into chunks.  If this fails, iterative
	// decrease chunk width until something fits.
	for {
		var (
			overflows bit.Set
			// Right-hand side
			right = splitDividedVariables(divisions, p.RightHandSide, mapping)
		)
		// Attempt to chunk right-hand side
		rhsChunks, overflows = determineRhsChunks(right, lhsChunks, field, mapping)
		//
		if overflows.Count() == 0 {
			// Successful chunking, therefore include any constraints necessary
			// for splitting of non-linear terms and construct final equations.
			break
		}
		// Update divisions based on identified overflows
		updateVariableDivisions(divisions, overflows)
		// Reset any allocated carry registers as we are starting over
		mapping.Reset(n)
	}
	// Reconstruct equations
	for i := range len(lhsChunks) {
		l := lhsChunks[i]
		r := rhsChunks[i]
		//
		assignments = append(assignments, NewAssignment2(l.contents, r.contents))
	}
	// Done
	return assignments
}

func initialiseVariableDivisions(n uint) []uint {
	var divisions = make([]uint, n)
	//
	for i := range divisions {
		divisions[i] = 1
	}
	//
	return divisions
}

func updateVariableDivisions(divisions []uint, vars bit.Set) {
	//
	for i := range divisions {
		if vars.Contains(uint(i)) {
			divisions[i] *= 2
		}
	}
}

func splitDividedVariables(divisions []uint, p StaticPolynomial, mapping RegisterAllocator) StaticPolynomial {
	panic("todo")
}

func determineLhsChunks(regs []register.Id, chunkWidth uint, mapping register.Map) []LhsChunk {
	var chunks []Chunk[[]register.Id]
	//
	for len(regs) != 0 {
		var chunk Chunk[[]register.Id]
		// Determine next chunkd
		chunk, regs = getNextLhsChunk(regs, chunkWidth, mapping)
		chunks = append(chunks, chunk)
	}
	//
	return chunks
}

func getNextLhsChunk(regs []register.Id, chunkWidth uint, mapping register.Map) (LhsChunk, []register.Id) {
	var bitwidth uint
	//
	for i, r := range regs {
		reg := mapping.Register(r)
		//
		if bitwidth+reg.Width > chunkWidth {
			return LhsChunk{bitwidth, regs[:i]}, regs[i:]
		}
		//
		bitwidth += reg.Width
	}
	//
	return LhsChunk{bitwidth, regs}, nil
}

// Divide a polynomial into "chunks", each of which has a maximum bitwidth as
// determined by the chunk widths.  This inserts carry lines as needed to ensure
// correctness.
func determineRhsChunks(p StaticPolynomial, lhsChunks []LhsChunk, field field.Config,
	mapping RegisterAllocator) ([]RhsChunk, bit.Set) {
	//
	var (
		env    = StaticEnvironment(mapping)
		chunks []RhsChunk
		vars   bit.Set
	)
	// Subdivide polynomial into chunks
	for _, ith := range lhsChunks {
		// TODO: carry lines
		var remainder StaticPolynomial
		// Chunk the polynomial
		p, remainder = p.Shr(ith.bitwidth)
		// Determine chunk width
		chunkWidth, _ := WidthOfPolynomial(remainder, env)
		// Check whether chunk fits
		if chunkWidth > field.BandWidth {
			// No, it does not.
			panic("got here")
		}
		//
		chunks = append(chunks, RhsChunk{chunkWidth, remainder})
	}
	//
	return chunks, vars
}

// Chunk represents a "chunk information bits".
type Chunk[T any] struct {
	bitwidth uint
	contents T
}

// LhsChunk captures the chunk type used for the Left-Hand Side (LHS) of an
// assignment.
type LhsChunk = Chunk[[]register.Id]

// RhsChunk captures the chunk type used for the Right-Hand Side (RHS) of an
// assignment.
type RhsChunk = Chunk[StaticPolynomial]

// StaticPoly2String provides a convenient helper function for debugging polynomials.
func StaticPoly2String(p StaticPolynomial, env register.Map) string {
	return poly.String(p, func(r register.Id) string {
		return env.Register(r.Id()).Name
	})
}
