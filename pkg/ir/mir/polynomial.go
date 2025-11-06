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
package mir

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/ir/term"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
)

var (
	biONE big.Int = *big.NewInt(1)
)

// Polynomial provides a useful alias
type Polynomial = agnostic.DynamicPolynomial

// ============================================================================
// Term => Polynomial
// ============================================================================

// Translate a term into a polynomial.
func termToPolynomial[F field.Element[F]](term Term[F], mapping register.Map) Polynomial {
	switch t := term.(type) {
	case *Add[F]:
		return termAddToPolynomial(*t, mapping)
	case *Constant[F]:
		return termConstantToPolynomial(t.Value, mapping)
	case *RegisterAccess[F]:
		return termRegAccessToPolynomial(*t)
	case *Mul[F]:
		return termMulToPolynomial(*t, mapping)
	case *Sub[F]:
		return termSubToPolynomial(*t, mapping)
	case *VectorAccess[F]:
		return termVecAccessToPolynomial(*t, mapping)
	default:
		panic("unreachable")
	}
}

func termAddToPolynomial[F field.Element[F]](term Add[F], mapping register.Map) Polynomial {
	var result Polynomial
	//
	for i, e := range term.Args {
		ith := termToPolynomial(e, mapping)
		//
		if i == 0 {
			result = ith
		} else {
			result = result.Add(ith)
		}
	}
	//
	return result
}

func termConstantToPolynomial[F field.Element[F]](constant F, mapping register.Map) Polynomial {
	var (
		result Polynomial
		value  big.Int
	)
	value.SetBytes(constant.Bytes())
	monomial := poly.NewMonomial[register.AccessId](value)
	//
	return result.Set(monomial)
}

func termMulToPolynomial[F field.Element[F]](term Mul[F], mapping register.Map) Polynomial {
	var result Polynomial
	//
	for i, e := range term.Args {
		ith := termToPolynomial(e, mapping)
		//
		if i == 0 {
			result = ith
		} else {
			result = result.Mul(ith)
		}
	}
	//
	return result
}

func termRegAccessToPolynomial[F field.Element[F]](term RegisterAccess[F]) Polynomial {
	var (
		identifier = term.Register().AccessOf(term.Bitwidth(), term.Shift())
		monomial   = poly.NewMonomial(biONE, identifier)
		result     Polynomial
	)
	//
	return result.Set(monomial)
}

func termSubToPolynomial[F field.Element[F]](term Sub[F], mapping register.Map) Polynomial {
	var result Polynomial
	//
	for i, e := range term.Args {
		ith := termToPolynomial(e, mapping)
		//
		if i == 0 {
			result = ith
		} else {
			result = result.Sub(ith)
		}
	}
	//
	return result
}

func termVecAccessToPolynomial[F field.Element[F]](term VectorAccess[F], mapping register.Map) Polynomial {
	var (
		result Polynomial
		shift  uint = 0
	)
	//
	for i, v := range term.Vars {
		var (
			reg      = mapping.Register(v.Register())
			regWidth = min(reg.Width, v.Bitwidth())
			ith      = termRegAccessToPolynomial(*v)
		)
		// Add to poly
		if i == 0 {
			result = ith
		} else {
			// Shift ith term
			ith = ith.MulScalar(util_math.Pow2(shift))
			// Add ith term
			result = result.Add(ith)
		}
		// Increase shift
		shift += regWidth
	}
	// Done
	return result
}

// ============================================================================
// Polynomial => Term
// ============================================================================

// Translate a term into a polynomial.
func polynomialToTerm[F field.Element[F]](poly Polynomial) Term[F] {
	var (
		pos []Term[F]
		neg []Term[F]
	)
	//
	for i := range poly.Len() {
		ith := poly.Term(i)
		//
		if ith.IsNegative() {
			neg = append(neg, monomialToTerm[F](poly.Term(i)))
		} else {
			pos = append(pos, monomialToTerm[F](poly.Term(i)))
		}
	}
	// Handle negative monomials (if applicable)
	if len(neg) != 0 {
		return term.Subtract(term.Sum(pos...), term.Sum(neg...))
	}
	//
	return term.Sum(pos...)
}

func monomialToTerm[F field.Element[F]](monomial agnostic.DynamicMonomial) Term[F] {
	var (
		terms = make([]Term[F], monomial.Len()+1)
		tmp   = monomial.Coefficient()
		coeff F
	)
	// Add coefficient
	terms[0] = term.Const[F, Term[F]](coeff.SetBytes(tmp.Bytes()))
	//
	for i := range monomial.Len() {
		ith := monomial.Nth(i)
		terms[i+1] = term.NarrowRegisterAccess[F, Term[F]](ith.Id(), ith.Bitwidth(), ith.Shift())
	}
	//
	return term.Product(terms...)
}
