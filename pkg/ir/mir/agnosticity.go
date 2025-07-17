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
	"github.com/consensys/go-corset/pkg/schema/constraint"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideAssertion(c Assertion, _ schema.RegisterMap) Assertion {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideInterleaving(c InterleavingConstraint, _ schema.RegisterMap) InterleavingConstraint {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideLookup(c LookupConstraint, mapping schema.RegisterMap) LookupConstraint {
	var (
		sources = splitEnclosedTerms(c.Sources, mapping)
		targets = splitEnclosedTerms(c.Targets, mapping)
	)
	// FIXME: this is not really safe in the general case.  For example, this
	// could result in a mismatched number of columns.  Furthermore, its
	// possible these columns are incorrectly aligned, etc.
	targets = flattenEnclosedVectors(targets)
	sources = flattenEnclosedVectors(sources)
	// Sanity check for now
	if len(sources) != len(targets) {
		panic("misaligned lookup")
	}
	//
	return constraint.NewLookupConstraint(c.Handle, targets, sources)
}

// Subdivide implementation for the FieldAgnostic interface.
func subdividePermutation(c PermutationConstraint, _ schema.RegisterMap) PermutationConstraint {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideRange(c RangeConstraint, _ schema.RegisterMap) RangeConstraint {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideSorted(c SortedConstraint, _ schema.RegisterMap) SortedConstraint {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideVanishing(p VanishingConstraint, mapping schema.ModuleRegisterMap) VanishingConstraint {
	// Split all registers occurring in the logical term.
	c := splitLogicalTerm(p.Constraint, mapping)
	// FIXME: this is an insufficient solution because it does not address the
	// potential issues around bandwidth.  Specifically, where additional carry
	// lines are needed, etc.
	return constraint.NewVanishingConstraint(p.Handle, p.Context, p.Domain, c)
}

func splitLogicalTerm(term LogicalTerm, mapping schema.ModuleRegisterMap) LogicalTerm {
	switch t := term.(type) {
	case *Conjunct:
		return ir.Conjunction(splitLogicalTerms(t.Args, mapping)...)
	case *Disjunct:
		return ir.Disjunction(splitLogicalTerms(t.Args, mapping)...)
	case *Equal:
		return ir.Equals[LogicalTerm](splitTerm(t.Lhs, mapping), splitTerm(t.Rhs, mapping))
	case *Ite:
		condition := splitLogicalTerm(t.Condition, mapping)
		trueBranch := splitOptionalLogicalTerm(t.TrueBranch, mapping)
		falseBranch := splitOptionalLogicalTerm(t.FalseBranch, mapping)
		//
		return ir.IfThenElse(condition, trueBranch, falseBranch)
	case *Negate:
		return ir.Negation(splitLogicalTerm(t.Arg, mapping))
	case *NotEqual:
		return ir.NotEquals[LogicalTerm](splitTerm(t.Lhs, mapping), splitTerm(t.Rhs, mapping))
	case *Inequality:
		if t.Strict {
			return ir.LessThan[LogicalTerm](splitTerm(t.Lhs, mapping), splitTerm(t.Rhs, mapping))
		}
		//
		return ir.LessThanOrEquals[LogicalTerm](splitTerm(t.Lhs, mapping), splitTerm(t.Rhs, mapping))
	default:
		panic("unreachable")
	}
}

func splitEnclosedTerms(vectors []ir.Enclosed[[]Term], mapping schema.RegisterMap) []ir.Enclosed[[]Term] {
	var nterms = make([]ir.Enclosed[[]Term], len(vectors))
	//
	for i, vector := range vectors {
		var (
			modmap = mapping.Module(vector.Module)
			terms  = splitTerms(vector.Item, modmap)
		)
		//
		nterms[i] = ir.Enclose(vector.Module, terms)
	}
	//
	return nterms
}

func splitOptionalLogicalTerm(term LogicalTerm, mapping schema.ModuleRegisterMap) LogicalTerm {
	if term == nil {
		return nil
	}
	//
	return splitLogicalTerm(term, mapping)
}

func splitLogicalTerms(terms []LogicalTerm, mapping schema.ModuleRegisterMap) []LogicalTerm {
	var nterms = make([]LogicalTerm, len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitLogicalTerm(terms[i], mapping)
	}
	//
	return nterms
}

func splitTerm(term Term, mapping schema.ModuleRegisterMap) Term {
	switch t := term.(type) {
	case *Add:
		return ir.Sum(splitTerms(t.Args, mapping)...)
	case *Cast:
		return ir.CastOf(splitTerm(t.Arg, mapping), t.BitWidth)
	case *Constant:
		return t
	case *IfZero:
		return ir.IfElse(
			splitLogicalTerm(t.Condition, mapping),
			splitTerm(t.TrueBranch, mapping),
			splitTerm(t.FalseBranch, mapping),
		)
	case *LabelledConst:
		return t
	case *RegisterAccess:
		if t.Register.IsUsed() {
			return splitRegisterAccess(t, mapping)
		}
		// NOTE: this indicates an unused register access.  Currently, this can
		// only occur for the selector column of a lookup.  This behaviour maybe
		// deprecated in the future, and that would make this check
		// unnecessary.
		return t
	case *Exp:
		return ir.Exponent(splitTerm(t.Arg, mapping), t.Pow)
	case *Mul:
		return ir.Product(splitTerms(t.Args, mapping)...)
	case *Norm:
		return ir.Normalise(splitTerm(t.Arg, mapping))
	case *Sub:
		return ir.Subtract(splitTerms(t.Args, mapping)...)
	case *VectorAccess:
		return splitVectorAccess(t, mapping)
	default:
		panic("unreachable")
	}
}

func splitTerms(terms []Term, mapping schema.ModuleRegisterMap) []Term {
	var nterms []Term = make([]Term, len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitTerm(terms[i], mapping)
	}
	//
	return nterms
}

func splitRegisterAccess(term *RegisterAccess, mapping schema.ModuleRegisterMap) Term {
	var (
		// Determine limbs for this register
		limbs = mapping.LimbIds(term.Register)
		// Construct appropriate terms
		terms = make([]*RegisterAccess, len(limbs))
	)
	// Check whether anything to do?
	if len(limbs) == 1 {
		// Nope
		return term
	}
	//
	for i, limb := range limbs {
		terms[i] = &ir.RegisterAccess[Term]{Register: limb, Shift: term.Shift}
	}
	//
	return ir.NewVectorAccess(terms)
}

func splitVectorAccess(term *VectorAccess, mapping schema.ModuleRegisterMap) Term {
	var terms []*RegisterAccess
	//
	for _, v := range term.Vars {
		for _, limb := range mapping.LimbIds(v.Register) {
			term := &ir.RegisterAccess[Term]{Register: limb, Shift: v.Shift}
			terms = append(terms, term)
		}
	}
	//
	return ir.NewVectorAccess(terms)
}

func flattenEnclosedVectors(vectors []ir.Enclosed[[]Term]) []ir.Enclosed[[]Term] {
	var nterms = make([]ir.Enclosed[[]Term], len(vectors))
	//
	for i, vector := range vectors {
		var (
			terms = flattenTerms(vector.Item)
		)
		//
		nterms[i] = ir.Enclose(vector.Module, terms)
	}
	//
	return nterms
}

func flattenTerms(terms []Term) []Term {
	var nterms []Term
	//
	for _, t := range terms {
		if va, ok := t.(*VectorAccess); ok {
			for _, v := range va.Vars {
				nterms = append(nterms, v)
			}
		} else {
			nterms = append(nterms, t)
		}
	}
	//
	return nterms
}
