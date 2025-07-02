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

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// Polynomial defines the type of polynomials over which packets (and register
// splitting in general) operate.
type Polynomial = *poly.ArrayPoly[schema.RegisterId]

// Monomial defines the type of monomials contained within a given polynomial.
type Monomial = poly.Monomial[schema.RegisterId]

// SplitPolynomial splits
func SplitPolynomial(p Polynomial, env schema.RegisterMapping) Polynomial {
	var npoly Polynomial
	//
	for i := range p.Len() {
		ith := splitMonomial(p.Term(i), env)
		//
		if i == 0 {
			npoly = ith
		} else {
			npoly = npoly.Add(ith)
		}
	}
	//
	return npoly
}

func splitMonomial(p Monomial, env schema.RegisterMapping) Polynomial {
	panic("todo")
}

// BitwidthOfPolynomial determines the minimum number of bits required to store
// all possible evaluations of this polynomial.  Observe that, in the case of
// negative values, this must include the sign bit as well.  For example, a
// polynomial contained within the range 0..255 has a width of 8 bits. Likewise,
// a polynomial contained within the range -17 .. 15 has a width of 6bits.  To
// understand this, consider that the positive component (0..15) has a width of
// 4 and the negative component (-17..-1) a width of 5.  Since a sign bit is
// needed to distinguish the two cases, we have an overall width of 6 bits
// required for the polynomial.
//
// To determine the bitwidth of a polynomial, this function first determines its
// smallest enclosing integer range.  From this is then determines the required
// widths of the negative and positive components, before combining them to give
// the result.
func BitwidthOfPolynomial(source Polynomial, regs []schema.Register) uint {
	var (
		intRange  = IntegerRangeOfPolynomial(source, regs)
		lower     = intRange.MinValue()
		upper     = intRange.MaxValue()
		lowerBits = uint(lower.BitLen())
		upperBits = uint(upper.BitLen())
	)
	// Check whether negative range in play.
	if lower.Cmp(&zero) < 0 {
		// Yes, we have negative values.  This mandates the need for an
		// additional signbit.
		return max(lowerBits+1, upperBits)
	}
	// No sign bit required.
	return upperBits
}

// IntegerRangeOfPolynomial determines the smallest integer range in which all
// evaluations of this polynomial lie.  For example, consider "2*X + 1" where X
// is an 8bit register.  Then, the smallest integer range which includes this
// polynomial is "0..511".
func IntegerRangeOfPolynomial(poly Polynomial, regs []schema.Register) math.Interval {
	var intRange math.Interval
	//
	for i := range poly.Len() {
		intRange.Add(IntegerRangeOfMonomial(poly.Term(i), regs))
	}
	//
	return intRange
}

// IntegerRangeOfMonomial determines the smallest integer range in which all
// evaluations of the monomial lie.  For example, consider the monomial "3*X*Y"
// where X and are 8bit and 16bit registers respectively.  Then, the smallest
// enclosing integer range is 0 .. 3*255*65535.
func IntegerRangeOfMonomial(mono Monomial, regs []schema.Register) *math.Interval {
	var (
		coeff    = mono.Coefficient()
		intRange = math.NewInterval(&coeff, &coeff)
	)
	//
	for i := range mono.Len() {
		intRange.Mul(IntegerRangeOfRegister(mono.Nth(i), regs))
	}
	//
	return intRange
}

// IntegerRangeOfRegister determines the smallest integer range enclosing all possible
// values for a given register.  For example, a register of width 16 has an
// integer range of 0..65535 (inclusive).
func IntegerRangeOfRegister(rid schema.RegisterId, regs []schema.Register) *math.Interval {
	var (
		val   = big.NewInt(2)
		width = regs[rid.Unwrap()].Width
	)
	// NOTE: following is safe since the width of any registers must sure be
	// less than 65536 bits :)
	val.Exp(val, big.NewInt(int64(width)), nil)
	// Subtract one since the interval is inclusive.
	return math.NewInterval(&zero, val.Sub(val, &one))
}
