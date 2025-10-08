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
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideAssertion[F field.Element[F]](c Assertion[F], _ schema.LimbsMap) Assertion[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideInterleaving[F field.Element[F]](c InterleavingConstraint[F], _ schema.LimbsMap,
) InterleavingConstraint[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdividePermutation[F field.Element[F]](c PermutationConstraint[F], _ schema.LimbsMap) PermutationConstraint[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideRange[F field.Element[F]](c RangeConstraint[F], _ schema.LimbsMap) RangeConstraint[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideSorted[F field.Element[F]](c SortedConstraint[F], _ schema.LimbsMap) SortedConstraint[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideVanishing[F field.Element[F]](p VanishingConstraint[F], mapping schema.RegisterLimbsMap,
) VanishingConstraint[F] {
	// Split all registers occurring in the logical term.
	c := splitLogicalTerm(p.Constraint, mapping)
	// FIXME: this is an insufficient solution because it does not address the
	// potential issues around bandwidth.  Specifically, where additional carry
	// lines are needed, etc.
	return vanishing.NewConstraint(p.Handle, p.Context, p.Domain, c)
}

func splitLogicalTerm[F field.Element[F]](term LogicalTerm[F], mapping schema.RegisterLimbsMap) LogicalTerm[F] {
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

func splitOptionalLogicalTerm[F field.Element[F]](term LogicalTerm[F], mapping schema.RegisterLimbsMap) LogicalTerm[F] {
	if term == nil {
		return nil
	}
	//
	return splitLogicalTerm(term, mapping)
}

func splitLogicalTerms[F field.Element[F]](terms []LogicalTerm[F], mapping schema.RegisterLimbsMap) []LogicalTerm[F] {
	var nterms = make([]LogicalTerm[F], len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitLogicalTerm(terms[i], mapping)
	}
	//
	return nterms
}

func splitEquality[F field.Element[F]](sign bool, lhs, rhs Term[F], mapping schema.RegisterLimbsMap) LogicalTerm[F] {
	//
	lhs = splitTerm(lhs, mapping)
	rhs = splitTerm(rhs, mapping)
	//
	if sign {
		return ir.Equals[F, LogicalTerm[F]](lhs, rhs)
	}
	//
	return ir.NotEquals[F, LogicalTerm[F]](lhs, rhs)
}

func splitTerm[F field.Element[F]](term Term[F], mapping schema.RegisterLimbsMap) Term[F] {
	switch t := term.(type) {
	case *Add[F]:
		return ir.Sum(splitTerms(t.Args, mapping)...)
	case *Constant[F]:
		return t
	case *RegisterAccess[F]:
		return splitRegisterAccess(t, mapping)
	case *Mul[F]:
		return ir.Product(splitTerms(t.Args, mapping)...)
	case *Sub[F]:
		return ir.Subtract(splitTerms(t.Args, mapping)...)
	case *VectorAccess[F]:
		return splitVectorAccess(t, mapping)
	default:
		panic("unreachable")
	}
}

func splitTerms[F field.Element[F]](terms []Term[F], mapping schema.RegisterLimbsMap) []Term[F] {
	var nterms []Term[F] = make([]Term[F], len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitTerm(terms[i], mapping)
	}
	//
	return nterms
}

func splitRegisterAccess[F field.Element[F]](term *RegisterAccess[F], mapping schema.RegisterLimbsMap) Term[F] {
	var (
		// Determine limbs for this register
		limbs = mapping.LimbIds(term.Register)
		// Construct appropriate terms
		terms = make([]*RegisterAccess[F], len(limbs))
	)
	//
	for i, limb := range limbs {
		terms[i] = &ir.RegisterAccess[F, Term[F]]{Register: limb, Shift: term.Shift}
	}
	// Check whether vector required, or not
	if len(limbs) == 1 {
		// NOTE: we cannot return the original term directly, as its index may
		// differ under the limb mapping.
		return terms[0]
	}
	//
	return ir.NewVectorAccess(terms)
}

func splitVectorAccess[F field.Element[F]](term *VectorAccess[F], mapping schema.RegisterLimbsMap) Term[F] {
	var terms []*RegisterAccess[F]
	//
	for _, v := range term.Vars {
		for _, limb := range mapping.LimbIds(v.Register) {
			term := &ir.RegisterAccess[F, Term[F]]{Register: limb, Shift: v.Shift}
			terms = append(terms, term)
		}
	}
	//
	return ir.NewVectorAccess(terms)
}

func splitRawRegisterAccesses[F field.Element[F]](terms []*RegisterAccess[F], mapping schema.RegisterLimbsMap,
) []*VectorAccess[F] {
	//
	var (
		vecs = make([]*VectorAccess[F], len(terms))
	)
	//
	for i, term := range terms {
		vecs[i] = splitRawRegisterAccess(term, mapping)
	}
	//
	return vecs
}

func splitRawRegisterAccess[F field.Element[F]](term *RegisterAccess[F], mapping schema.RegisterLimbsMap,
) *VectorAccess[F] {
	//
	var (
		// Determine limbs for this register
		limbs = mapping.LimbIds(term.Register)
		// Construct appropriate terms
		terms = make([]*RegisterAccess[F], len(limbs))
	)
	//
	for i, limb := range limbs {
		terms[i] = &ir.RegisterAccess[F, Term[F]]{Register: limb, Shift: term.Shift}
	}
	//
	return ir.RawVectorAccess(terms)
}
