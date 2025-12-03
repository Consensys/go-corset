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
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
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
		// Record initial number of registers
		n = uint(len(mapping.Registers()))
		//
		splitter = NewRegisterSplitter(n)
		// Determine the bitwidth of each chunk
		rhsChunks []RhsChunk
		lhsChunks []LhsChunk
		//
		initLhsChunks = initialiseLhsChunks(p.LeftHandSide, field.RegisterWidth, mapping)
	)
	// Attempt to divide polynomials into chunks.  If this fails, iterative
	// decrease chunk width until something fits.
	for {
		var (
			overflows bit.Set
			// Right-hand side
			right = splitter.Apply(p.RightHandSide, mapping)
		)
		// Attempt to chunk right-hand side
		lhsChunks, rhsChunks, overflows = determineRhsChunks(right, initLhsChunks, field, mapping)
		//
		if overflows.Count() == 0 {
			// Successful chunking, therefore include any constraints necessary
			// for splitting of non-linear terms and construct final equations.
			break
		}
		// Update divisions based on identified overflows
		splitter.Subdivide(overflows)
		// Reset any allocated carry registers as we are starting over
		mapping.Reset(n)
		splitter.Reset()
	}
	// Initialise with splits
	assignments := splitter.assignments
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

func initialiseLhsChunks(regs []register.Id, chunkWidth uint, mapping register.Map) []LhsChunk {
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
func determineRhsChunks(p StaticPolynomial, chunks []LhsChunk, field field.Config,
	mapping RegisterAllocator) ([]LhsChunk, []RhsChunk, bit.Set) {
	//
	var (
		env       = StaticEnvironment(mapping)
		rhsChunks []RhsChunk
		lhsChunks []LhsChunk
		vars      bit.Set
	)
	// Subdivide polynomial into chunks
	for i, ith := range chunks {
		var (
			remainder StaticPolynomial
			lhsChunk  LhsChunk
		)
		// Chunk the polynomial
		p, remainder = p.Shr(ith.bitwidth)
		// Determine chunk width
		chunkWidth, _ := WidthOfPolynomial(remainder, env)
		// Check whether chunk fits
		if chunkWidth > field.BandWidth {
			// No, it does not.
			vars.Union(RegisterReadSet(remainder))
		} else if i+1 != len(lhsChunks) && chunkWidth > ith.bitwidth {
			// Overflow case.
			var carry StaticPolynomial
			// Determine width of carry register
			overflow := chunkWidth - ith.bitwidth
			// Allocate new register to get an Id
			carryRegId := mapping.Allocate("c", overflow)
			// Propage carry forward
			p = p.Add(carry.Set(poly.NewMonomial(one, carryRegId)))
			// include carry in lhs
			lhsChunk = LhsChunk{ith.bitwidth, array.Append(ith.contents, carryRegId)}
		} else {
			// lhs chunk unchanged
			lhsChunk = ith
		}
		//
		rhsChunks = append(rhsChunks, RhsChunk{chunkWidth, remainder})
		lhsChunks = append(lhsChunks, lhsChunk)
	}
	//
	return lhsChunks, rhsChunks, vars
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

// RegisterSplitter is used to manage the mechanism of splitting variables into
// limbs in order to improve the precision of a chunk.
type RegisterSplitter struct {
	divisions   []uint
	assignments []Assignment2
}

// NewRegisterSplitter constructs a new splitter for a given number of variables.
func NewRegisterSplitter(n uint) RegisterSplitter {
	var divisions = make([]uint, n)
	//
	for i := range divisions {
		divisions[i] = 1
	}
	//
	return RegisterSplitter{divisions, nil}
}

// Subdivide takes all registers in the given set and further subdivides them.
// For example, if they were previously being divided in 2, then they will now
// be divided in 4, etc.
func (p *RegisterSplitter) Subdivide(vars bit.Set) {
	//
	for i := range p.divisions {
		if vars.Contains(uint(i)) {
			p.divisions[i] *= 2
		}
	}
}

// Reset assignments create for the current splitting.
func (p *RegisterSplitter) Reset() {
	p.assignments = nil
}

// Apply the current subdivisions to a given polynomial.  Specifically, this
// splits all registers into their limbs and subsitutes them into the
// polynomial.  As a by-product this also records the assignments needed for the
// mapping from registers to their limbs.
func (p *RegisterSplitter) Apply(poly StaticPolynomial, mapping RegisterAllocator) StaticPolynomial {
	var cache = make(map[register.Id]StaticPolynomial)
	//
	for i, div := range p.divisions {
		var (
			rid = register.NewId(uint(i))
			reg = mapping.Register(rid)
		)
		//
		if div != 1 && reg.Width > 1 {
			var limbPoly = p.splitVariable(rid, div, mapping)
			//
			cache[rid] = limbPoly
		}
	}
	//
	return SubstitutePolynomial(poly, func(reg register.Id) StaticPolynomial {
		if rp, ok := cache[reg]; ok {
			return rp
		}
		// no substitution
		return nil
	})
}

func (p *RegisterSplitter) splitVariable(rid register.Id, div uint, mapping RegisterAllocator) (r StaticPolynomial) {
	var (
		one      big.Int = *big.NewInt(1)
		rhs      StaticPolynomial
		terms    []StaticMonomial
		reg      = mapping.Register(rid)
		maxWidth = reg.Width / div
		bitwidth = reg.Width
		width    uint
	)
	// Round up (if necessary)
	if (maxWidth * div) < reg.Width {
		maxWidth++
	}
	// Determine limb widths
	limbWidths := register.LimbWidths(maxWidth, reg.Width)
	// Allocate limbs
	limbs := mapping.AllocateN(reg.Name, limbWidths)
	// Construct limb polynomial
	for i, limb := range limbs {
		var (
			c         = util_math.Pow2(width)
			limbWidth = min(bitwidth, limbWidths[i])
		)
		//
		if limbWidth > 0 {
			terms = append(terms, poly.NewMonomial(*c, limb))
			width += limbWidth
		}
		//
		bitwidth -= limbWidth
	}
	// Update assignments
	assignment := NewAssignment2(limbs, rhs.Set(poly.NewMonomial(one, rid)))
	p.assignments = append(p.assignments, assignment)
	//
	return r.Set(terms...)
}
