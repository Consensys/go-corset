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
package poly

import "math/big"

// Monomial represents a monomial within an array polynomial.
type Monomial[S comparable] struct {
	coefficient big.Int
	vars        []S
}

// NewMonomial constructs a new array term with a given coefficient and zero or
// more variables.
func NewMonomial[S comparable](coefficient big.Int, vars ...S) Monomial[S] {
	return Monomial[S]{coefficient, vars}
}

// Coefficient returns the coefficient of this term.
func (p Monomial[S]) Coefficient() big.Int {
	return p.coefficient
}

// Len returns the number of variables in this polynomial term.
func (p Monomial[S]) Len() uint {
	return uint(len(p.vars))
}

// Nth returns the nth variable in this polynomial term.
func (p Monomial[S]) Nth(index uint) S {
	return p.vars[index]
}

// Neg returns a negated copy of this monomial
func (p Monomial[S]) Neg() Monomial[S] {
	var res = p.Clone()
	// Negate Coefficient
	res.coefficient.Neg(&res.coefficient)
	// Done
	return res
}

// Mul returns a fresh monomial representing the multiplication of this monomial
// and another.
func (p Monomial[S]) Mul(other Monomial[S]) Monomial[S] {
	var res = p.Clone()
	// Multiply coefficients
	res.coefficient.Mul(&res.coefficient, &other.coefficient)
	// Append variables
	res.vars = append(res.vars, other.vars...)
	// Done
	return res
}

// Matches determines whether or not the variables of this term match those
// of the other.
func (p Monomial[S]) Matches(other Monomial[S]) bool {
	if p.Len() != other.Len() {
		return false
	}
	//
	for i := uint(0); i < p.Len(); i++ {
		if p.vars[i] != other.Nth(i) {
			return false
		}
	}
	//
	return true
}

// Clone an array term
func (p *Monomial[S]) Clone() Monomial[S] {
	var (
		val   big.Int
		nvars = make([]S, len(p.vars))
	)
	// Copy variables
	copy(nvars, p.vars)
	// Copy coefficient
	val.Set(&p.coefficient)
	//
	return Monomial[S]{val, nvars}
}
