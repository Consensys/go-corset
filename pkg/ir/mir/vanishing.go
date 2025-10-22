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
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideVanishing[F field.Element[F]](p VanishingConstraint[F], mapping module.LimbsMap,
	env agnostic.RegisterAllocator) VanishingConstraint[F] {
	//
	var (
		modmap = mapping.Module(p.Context)
		// Split all registers occurring in the logical term.
		c = splitLogicalTerm(p.Constraint, modmap, env)
	)
	// FIXME: this is an insufficient solution because it does not address the
	// potential issues around bandwidth.  Specifically, where additional carry
	// lines are needed, etc.
	return vanishing.NewConstraint(p.Handle, p.Context, p.Domain, c)
}

func splitLogicalTerm[F field.Element[F]](term LogicalTerm[F], mapping register.LimbsMap,
	env agnostic.RegisterAllocator) LogicalTerm[F] {
	//
	switch t := term.(type) {
	case *Conjunct[F]:
		return ir.Conjunction(splitLogicalTerms(t.Args, mapping, env)...)
	case *Disjunct[F]:
		return ir.Disjunction(splitLogicalTerms(t.Args, mapping, env)...)
	case *Equal[F]:
		return splitEquality(true, t.Lhs, t.Rhs, mapping, env)
	case *Ite[F]:
		condition := splitLogicalTerm(t.Condition, mapping, env)
		trueBranch := splitOptionalLogicalTerm(t.TrueBranch, mapping, env)
		falseBranch := splitOptionalLogicalTerm(t.FalseBranch, mapping, env)
		//
		return ir.IfThenElse(condition, trueBranch, falseBranch)
	case *Negate[F]:
		return ir.Negation(splitLogicalTerm(t.Arg, mapping, env))
	case *NotEqual[F]:
		return splitEquality(false, t.Lhs, t.Rhs, mapping, env)
	default:
		panic("unreachable")
	}
}

func splitOptionalLogicalTerm[F field.Element[F]](term LogicalTerm[F],
	mapping register.LimbsMap, env agnostic.RegisterAllocator) LogicalTerm[F] {
	//
	if term == nil {
		return nil
	}
	//
	return splitLogicalTerm(term, mapping, env)
}

func splitLogicalTerms[F field.Element[F]](terms []LogicalTerm[F],
	mapping register.LimbsMap, env agnostic.RegisterAllocator) []LogicalTerm[F] {
	//
	var nterms = make([]LogicalTerm[F], len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitLogicalTerm(terms[i], mapping, env)
	}
	//
	return nterms
}

func splitEquality[F field.Element[F]](sign bool, lhs, rhs Term[F], mapping register.LimbsMap,
	env agnostic.RegisterAllocator) LogicalTerm[F] {
	//
	var (
		// Split terms accordingl to mapping, and translate into polynomials
		left  = termToPolynomial(splitTerm(lhs, mapping), mapping.LimbsMap())
		right = termToPolynomial(splitTerm(rhs, mapping), mapping.LimbsMap())
		// Construct equality for spltting
		equation = agnostic.NewEquation(left, right)
		// Split the equation
		splitEquations = equation.Split(mapping.Field(), env)
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
