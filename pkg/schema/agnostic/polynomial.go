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

// SplitPolynomial splits the registers in a given polynomial into their limbs,
// producing an equivalent (but not necessarily identical) polynomial.  For
// example, suppose that X and Y split into limbs X'1, X'0 and Y'1, Y'0.  Then
// the polynomial 2*X + Y splits into 512*X'1 + 2*X'0 + 256*Y'1 + Y'0.
func SplitPolynomial(p Polynomial, env schema.RegisterLimbsMap) Polynomial {
	var npoly Polynomial
	//
	for i := range p.Len() {
		ith := SplitMonomial(p.Term(i), env)
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

// SplitMonomial splits a given monomial (e.g. 2*x*y) according to a given
// register-to-limb mapping.  For example, suppose x is u16 and maps to x'0 and
// x'1 (both u8), whilst y maps to itself.  Then, the resulting polynomial is:
//
// 2*(x'0 + 256*x'1)*y --> (2*x'0*y) + (512*x'1*y)
//
// Of course, things get more involved when more than one register is being
// split, but the basic idea above applies.
func SplitMonomial(p Monomial, env schema.RegisterLimbsMap) Polynomial {
	var res Polynomial
	// FIXME: what to do with the coefficient?  This is a problem because its
	// not clear how we should split this.  Presumably it should be split
	// according to the maximum register width.
	res = res.Set(poly.NewMonomial[schema.RegisterId](p.Coefficient()))
	//
	for i := range p.Len() {
		// Determine limbs corresponding to the given constraint.
		limbs := env.LimbIds(p.Nth(i))
		// Construct polynomial representing limbs
		ith := LimbPolynomial(limbs, env)
		//
		res = res.Mul(ith)
	}
	//
	return res
}

// LimbPolynomial constructs a polynomial from the given limbs which represents
// the value of the original register.  For example, suppose x is a u16 register
// which splits into two u8 limbs x'0 and x'1.  Then, the constructed "limb
// polynomial" is simply x'0 + 256*x'1 (recall that x'0 is the last significant
// limb).
func LimbPolynomial(limbs []schema.RegisterId, env schema.RegisterLimbsMap) Polynomial {
	var (
		res Polynomial
		// Offset is used to determine the coefficient for the next limb.
		offset big.Int = *big.NewInt(1)
		//
		terms = make([]Monomial, len(limbs))
	)
	//
	for i, rid := range limbs {
		var (
			coeff big.Int
			reg   = env.Limb(rid)
		)
		// Clone coefficient
		coeff.Set(&offset)
		// Construct term
		terms[i] = poly.NewMonomial(coeff, rid)
		// Shift offset up
		offset.Lsh(&offset, reg.Width)
	}
	// Done
	return res.Set(terms...)
}

// WidthOfPolynomial determines the minimum number of bits required to store
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
func WidthOfPolynomial(source Polynomial, regs []schema.Register) (bitwidth uint, signed bool) {
	var (
		intRange  = IntegerRangeOfPolynomial(source, regs)
		lower     = intRange.MinValue()
		upper     = intRange.MaxValue()
		upperBits = uint(upper.BitLen())
	)
	// Check whether negative range in play.
	if lower.Sign() < 0 {
		// NOTE: this accounts for the fact that, on the negative side, we get
		// an extra value.  For example, with signed 8bit values the range is
		// -128 upto 127.
		lowerBits := uint(lower.Add(&lower, &one).BitLen())
		// Yes, we have negative values.  This mandates the need for an
		// additional signbit.
		return max(lowerBits+1, upperBits+1), true
	}
	// No sign bit required.
	return upperBits, false
}

// SplitWidthOfPolynomial resturns the number of bits required for all positive
// values, along with the number of bits required for all negative values.
// Observe that, unlike WidthOfPolynomial, this does not account for an
// additional sign bit.
func SplitWidthOfPolynomial(source Polynomial, regs []schema.Register) (poswidth uint, negwidth uint) {
	var (
		intRange  = IntegerRangeOfPolynomial(source, regs)
		lower     = intRange.MinValue()
		upper     = intRange.MaxValue()
		upperBits = uint(upper.BitLen())
	)
	// Check whether negative range in play.
	if lower.Sign() < 0 {
		// NOTE: this accounts for the fact that, on the negative side, we get
		// an extra value.  For example, with signed 8bit values the range is
		// -128 upto 127.
		lowerBits := uint(lower.Add(&lower, &one).BitLen())
		// Yes, we have negative values.  This mandates the need for an
		// additional signbit.
		return upperBits, lowerBits
	}
	//
	return upperBits, 0
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
