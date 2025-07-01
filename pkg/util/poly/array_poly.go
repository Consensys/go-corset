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

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/array"
)

// ArrayPoly is the simpliest (and least efficient) polynomial implementation.
// It provides a reference against which other (more efficient) implementations
// can be compared.  Observe that an unitialised ArrayPoly variable corresponds
// with zero.
type ArrayPoly[S comparable] struct {
	terms []Monomial[S]
}

// Len returns the number of terms in this polynomial.
func (p *ArrayPoly[S]) Len() uint {
	return uint(len(p.terms))
}

// Term returns the ith term in this polynomial.
func (p *ArrayPoly[S]) Term(ith uint) Monomial[S] {
	return p.terms[ith]
}

// Set initialises this polynomial from zero or more terms.
func (p *ArrayPoly[S]) Set(terms ...Monomial[S]) *ArrayPoly[S] {
	if p != nil {
		p.terms = terms
		return p
	}
	//
	return &ArrayPoly[S]{terms}
}

// Clone performs a deep copy of this polynomial
func (p *ArrayPoly[S]) Clone() *ArrayPoly[S] {
	nterms := make([]Monomial[S], len(p.terms))
	//
	for i := range nterms {
		nterms[i] = p.terms[i].Clone()
	}
	//
	return &ArrayPoly[S]{nterms}
}

// IsZero returns an indication as to whether this polynomial is equivalent
// to zero (or not).  This is a three valued logic system which can return
// either "yes", "no" or "maybe" where: (i) "yes" means the polynomial
// always evaluates to zero; (ii) "no" means the polynomial never evaluates
// to zero; (iii) "maybe" indicates the polynomial may sometimes evaluate to
// zero.  When the return ok holds then res indicates either yes or not.
// Otherwise, the result is maybe.
func (p *ArrayPoly[S]) IsZero() (res bool, ok bool) {
	panic("todo")
}

// Add another polynomial onto this polynomial.
func (p *ArrayPoly[S]) Add(other *ArrayPoly[S]) *ArrayPoly[S] {
	var res = p.Clone()
	//
	for i := uint(0); i < other.Len(); i++ {
		res.AddTerm(other.Term(i))
	}
	//
	return res
}

// Sub another polynomial from this polynomil
func (p *ArrayPoly[S]) Sub(other *ArrayPoly[S]) *ArrayPoly[S] {
	var res = p.Clone()
	//
	for i := uint(0); i < other.Len(); i++ {
		res.SubTerm(other.Term(i))
	}
	//
	return res
}

// Mul this polynomial by another polynomial.
func (p *ArrayPoly[S]) Mul(other *ArrayPoly[S]) *ArrayPoly[S] {
	var res ArrayPoly[S]
	//
	for _, ith := range p.terms {
		for _, jth := range other.terms {
			res.AddTerm(ith.Mul(jth))
		}
	}
	//
	return &res
}

// AddTerm adds a single term into this polynomial.
func (p *ArrayPoly[S]) AddTerm(other Monomial[S]) {
	var zero = big.NewInt(0)
	//
	for i, term := range p.terms {
		if term.Matches(other) {
			ith := &p.terms[i]
			// Add term at this position
			ith.coefficient.Add(&ith.coefficient, &other.coefficient)
			// Check whether its now zero (or not)
			if ith.coefficient.Cmp(zero) == 0 {
				array.RemoveAt(p.terms, uint(i))
			}
			//
			return
		}
	}
	// Sort?
	p.terms = append(p.terms, other.Clone())
}

// SubTerm subtracts a single term from this polynomial.
func (p *ArrayPoly[S]) SubTerm(other Monomial[S]) {
	var zero = big.NewInt(0)
	//
	for i, term := range p.terms {
		if term.Matches(other) {
			ith := &p.terms[i]
			// Add term at this position
			ith.coefficient.Add(&ith.coefficient, &other.coefficient)
			// Check whether its now zero (or not)
			if ith.coefficient.Cmp(zero) == 0 {
				array.RemoveAt(p.terms, uint(i))
			}
			//
			return
		}
	}
	// Append negation to end
	// Sort?
	p.terms = append(p.terms, other.Neg())
}
