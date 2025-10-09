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
func (p *Equation) Split(bandwidth uint, env sc.RegisterAllocator) []Equation {
	var (
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
			worklist.PushReversed(next.innerSplit(bandwidth, env))
		} else {
			// no
			completed = append(completed, next)
		}
	}
	// Done
	return completed
}

func (p *Equation) innerSplit(bandwidth uint, env sc.RegisterAllocator) []Equation {
	var (
		// Sort both sides in order of their coefficients.
		lhs = sortByCoefficient(p.LeftHandSide)
		rhs = sortByCoefficient(p.RightHandSide)
		fn  = func(rid sc.RegisterId) string {
			return env.Limb(rid).Name
		}
	)
	// NOTES: at this point, do you just want to gobble each side upto the
	// bandwidth limit?  Then, you add a carry for the lower side.
	//
	for _, l := range lhs {
		fmt.Printf("[%s]", l.String(fn))
	}

	fmt.Println()

	for _, r := range rhs {
		fmt.Printf("[%s]", r.String(fn))
	}

	fmt.Println()
	panic(fmt.Sprintf("TODO: %s", p.String(env)))
}

// Sort the monomials in a given polynomial by their coefficient.
func sortByCoefficient(poly Polynomial) []Monomial {
	var monomials = make([]Monomial, poly.Len())
	// Extract them
	for i := range poly.Len() {
		monomials[i] = poly.Term(i)
	}
	// Sort them
	slices.SortFunc(monomials, func(l, r Monomial) int {
		var (
			lCoeff = l.Coefficient()
			rCoeff = r.Coefficient()
		)
		// Compare coefficients first
		if c := lCoeff.Cmp(&rCoeff); c != 0 {
			return c
		}
		// Compare variables second
		return slices.CompareFunc(l.Vars(), r.Vars(), func(l, r sc.RegisterId) int {
			return cmp.Compare(l.Unwrap(), r.Unwrap())
		})
	})
	// Done
	return monomials
}
