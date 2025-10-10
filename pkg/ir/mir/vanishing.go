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
	"github.com/consensys/go-corset/pkg/ir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideVanishing[F field.Element[F]](p VanishingConstraint[F], mapping sc.RegisterAllocator,
) VanishingConstraint[F] {
	// Split all registers occurring in the logical term.
	c := splitLogicalTerm(p.Constraint, mapping)
	// FIXME: this is an insufficient solution because it does not address the
	// potential issues around bandwidth.  Specifically, where additional carry
	// lines are needed, etc.
	return vanishing.NewConstraint(p.Handle, p.Context, p.Domain, c)
}

func splitLogicalTerm[F field.Element[F]](term LogicalTerm[F], mapping sc.RegisterAllocator) LogicalTerm[F] {
	switch t := term.(type) {
	case *Conjunct[F]:
		return ir.Conjunction(splitLogicalTerms(t.Args, mapping)...)
	case *Disjunct[F]:
		return ir.Disjunction(splitLogicalTerms(t.Args, mapping)...)
	case *Equal[F]:
		return splitEquality(true, t.Lhs, t.Rhs, mapping)
	case *Ite[F]:
		condition := splitLogicalTerm(t.Condition, mapping)
		trueBranch := splitOptionalLogicalTerm(t.TrueBranch, mapping)
		falseBranch := splitOptionalLogicalTerm(t.FalseBranch, mapping)
		//
		return ir.IfThenElse(condition, trueBranch, falseBranch)
	case *Negate[F]:
		return ir.Negation(splitLogicalTerm(t.Arg, mapping))
	case *NotEqual[F]:
		return splitEquality(false, t.Lhs, t.Rhs, mapping)
	default:
		panic("unreachable")
	}
}

func splitOptionalLogicalTerm[F field.Element[F]](term LogicalTerm[F],
	mapping sc.RegisterAllocator) LogicalTerm[F] {
	//
	if term == nil {
		return nil
	}
	//
	return splitLogicalTerm(term, mapping)
}

func splitLogicalTerms[F field.Element[F]](terms []LogicalTerm[F],
	mapping sc.RegisterAllocator) []LogicalTerm[F] {
	//
	var nterms = make([]LogicalTerm[F], len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitLogicalTerm(terms[i], mapping)
	}
	//
	return nterms
}

func splitEquality[F field.Element[F]](sign bool, lhs, rhs Term[F], mapping sc.RegisterAllocator) LogicalTerm[F] {
	var (
		// Split terms accordingl to mapping, and translate into polynomials
		left  = termToPolynomial(splitTerm(lhs, mapping), mapping.LimbsMap())
		right = termToPolynomial(splitTerm(rhs, mapping), mapping.LimbsMap())
		// Construct equality for spltting
		equation = agnostic.NewEquation(left, right)
		// Split the equation
		splitEquations = equation.Split(mapping)
		// Prepare resulting conjunct / disjunct
		terms = make([]LogicalTerm[F], len(splitEquations))
	)
	//
	for i, eq := range splitEquations {
		// reconstruct original term
		l := polynomialToTerm[F](eq.LeftHandSide)
		r := polynomialToTerm[F](eq.RightHandSide)
		//
		if sign {
			terms[i] = ir.Equals[F, LogicalTerm[F]](l, r)
		} else {
			terms[i] = ir.NotEquals[F, LogicalTerm[F]](l, r)
		}
	}
	// Done (for now)
	if sign {
		return ir.Conjunction(terms...)
	}
	//
	return ir.Disjunction(terms...)
}
