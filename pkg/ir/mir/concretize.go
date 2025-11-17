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
	"fmt"
	"reflect"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Element provides a convenient shorthand.
type Element[F any] = field.Element[F]

// LookupVector provides a convenient shorthand
type LookupVector[F any] = lookup.Vector[F, Term[F]]

// Concretize converts an MIR schema for a given field F1 into an MIR schema for
// another field F2.  This is awkward as we have to rebuild the entire
// Intermediate Representation in order to match the type appropriately. In
// doing this, we take some opportunities to simplify, such as removing labelled
// constants (which no longer make sense).  Furthermore, this stage can
// technically fail if the relevant constraints cannot be correctly concretized.
// For example, they contain a constant which does not fit within the field.
func Concretize[F1 Element[F1], F2 Element[F2]](mapping schema.LimbsMap, rawModules []Module[F1]) []Module[F2] {
	var (
		modules = make([]Module[F2], len(rawModules))
	)
	//
	for i, m := range rawModules {
		modules[i] = concretizeModule[F1, F2](m.Subdivide(mapping))
	}
	//
	return modules
}

func concretizeModule[F1 Element[F1], F2 Element[F2]](m Module[F1]) Module[F2] {
	var (
		r Module[F2]
		// Concreteize Assignments
		assignments = concretizeAssignments[F1, F2](m.RawAssignments())
		// Concreteize Constraints
		constraints = concretizeConstraints[F1, F2](m.RawConstraints())
	)
	// Initialise new module
	r = r.Init(m.Name(), m.LengthMultiplier(), m.AllowPadding(), m.IsPublic(), m.IsSynthetic())
	// Add concretized components
	r.AddRegisters(m.Registers()...)
	r.AddAssignments(assignments...)
	r.AddConstraints(constraints...)
	// Done
	return r
}

// ============================================================================
// Assignments
// ============================================================================

func concretizeAssignments[F1 Element[F1], F2 Element[F2]](assigns []schema.Assignment[F1]) []schema.Assignment[F2] {
	var rs = make([]schema.Assignment[F2], len(assigns))
	//
	for i, a := range assigns {
		rs[i] = concretizeAssignment[F1, F2](a)
	}
	//
	return rs
}

func concretizeAssignment[F1 Element[F1], F2 Element[F2]](assign schema.Assignment[F1]) schema.Assignment[F2] {
	switch a := assign.(type) {
	case *ComputedRegister[F1]:
		expr := concretizeTerm[F1, F2](a.Expr)
		return assignment.NewComputedRegister(a.Target, expr, a.Direction)
	case *assignment.Computation[F1]:
		return assignment.NewComputation[F2](a.Function, a.Targets, a.Sources)
	case *assignment.SortedPermutation[F1]:
		return assignment.NewSortedPermutation[F2](a.Targets, a.Signs, a.Sources)
	default:
		panic(fmt.Sprintf("unknown assignment: %s\n", reflect.TypeOf(a).String()))
	}
}

// ============================================================================
// Constraints
// ============================================================================

func concretizeConstraints[F1 Element[F1], F2 Element[F2]](constraints []Constraint[F1]) []Constraint[F2] {
	var rs = make([]Constraint[F2], len(constraints))
	//
	for i, c := range constraints {
		rs[i] = concretizeConstraint[F1, F2](c)
	}
	//
	return rs
}

func concretizeConstraint[F1 Element[F1], F2 Element[F2]](constraint Constraint[F1]) Constraint[F2] {
	//
	switch c := constraint.Unwrap().(type) {
	case Assertion[F1]:
		term := concretizeLogicalTerm[F1, F2](c.Property)
		//
		return NewAssertion(c.Handle, c.Context, c.Domain, term)
	case FunctionCall[F1]:
		return concretizeFunctionCall[F1, F2](c)
	case InterleavingConstraint[F1]:
		target := concretizeTerm[F1, F2](c.Target)
		sources := concretizeTerms[F1, F2](c.Sources)
		//
		return NewInterleavingConstraint(c.Handle, c.TargetContext, c.SourceContext, target, sources)
	case LookupConstraint[F1]:
		targets := concretizeLookupVectors[F1, F2](c.Targets)
		sources := concretizeLookupVectors[F1, F2](c.Sources)
		//
		return NewLookupConstraint(c.Handle, targets, sources)
	case PermutationConstraint[F1]:
		return NewPermutationConstraint[F2](c.Handle, c.Context, c.Targets, c.Sources)
	case RangeConstraint[F1]:
		term := concretizeTerm[F1, F2](c.Expr)
		//
		return NewRangeConstraint(c.Handle, c.Context, term, c.Bitwidth)
	case SortedConstraint[F1]:
		sources := concretizeTerms[F1, F2](c.Sources)
		selector := concretizeOptionalTerm[F1, F2](c.Selector)
		//
		return NewSortedConstraint(c.Handle, c.Context, c.BitWidth, selector, sources, c.Signs, c.Strict)
	case VanishingConstraint[F1]:
		term := concretizeLogicalTerm[F1, F2](c.Constraint)
		//
		return NewVanishingConstraint(c.Handle, c.Context, c.Domain, term)
	default:
		panic("unreachable")
	}
}

func concretizeFunctionCall[F1 Element[F1], F2 Element[F2]](fc FunctionCall[F1]) Constraint[F2] {
	var (
		nargs    = len(fc.Arguments)
		nrets    = len(fc.Returns)
		selector util.Option[Term[F2]]
		rets     = concretizeTerms[F1, F2](fc.Returns)
		args     = concretizeTerms[F1, F2](fc.Arguments)
		sources  = make([]lookup.Vector[F2, Term[F2]], 1)
		targets  = make([]lookup.Vector[F2, Term[F2]], 1)
	)
	// Concretize optional selector
	if fc.Selector.HasValue() {
		var (
			cond      = concretizeLogicalTerm[F1, F2](fc.Selector.Unwrap())
			truth     = ir.Const64[F2, Term[F2]](1)
			falsehood = ir.Const64[F2, Term[F2]](0)
		)
		// Construct conversion
		selector = util.Some(ir.IfElse(cond, truth, falsehood))
	}
	// Construct source vector
	sources[0] = lookup.NewVector(fc.Caller, selector, append(args, rets...)...)
	// Construct target vector
	targetTerms := make([]Term[F2], nargs+nrets)
	//
	for i := range nargs + nrets {
		var rid = sc.NewRegisterId(uint(i))
		//
		targetTerms[i] = ir.NewRegisterAccess[F2, Term[F2]](rid, 0)
	}
	// Done
	targets[0] = lookup.NewVector(fc.Callee, util.None[Term[F2]](), targetTerms...)
	//
	return NewLookupConstraint(fc.Handle, targets, sources)
}

func concretizeLookupVectors[F1 Element[F1], F2 Element[F2]](vecs []LookupVector[F1]) []LookupVector[F2] {
	var nvecs = make([]LookupVector[F2], len(vecs))
	//
	for i, vec := range vecs {
		nvecs[i] = concretizeLookupVector[F1, F2](vec)
	}
	//
	return nvecs
}

func concretizeLookupVector[F1 Element[F1], F2 Element[F2]](vec LookupVector[F1]) LookupVector[F2] {
	var (
		selector = concretizeOptionalTerm[F1, F2](vec.Selector)
		terms    = concretizeTerms[F1, F2](vec.Terms)
	)
	//
	return lookup.NewVector(vec.Module, selector, terms...)
}

// ============================================================================
// LogicalTerms
// ============================================================================

func concretizeLogicalTerm[F1 Element[F1], F2 Element[F2]](t LogicalTerm[F1]) LogicalTerm[F2] {
	switch t := t.(type) {
	case *Conjunct[F1]:
		return ir.Conjunction(concretizeLogicalTerms[F1, F2](t.Args)...)
	case *Disjunct[F1]:
		return ir.Disjunction(concretizeLogicalTerms[F1, F2](t.Args)...)
	case *Equal[F1]:
		lhs := concretizeTerm[F1, F2](t.Lhs)
		rhs := concretizeTerm[F1, F2](t.Rhs)
		//
		return ir.Equals[F2, LogicalTerm[F2]](lhs, rhs)
	case *Inequality[F1]:
		lhs := concretizeTerm[F1, F2](t.Lhs)
		rhs := concretizeTerm[F1, F2](t.Rhs)
		//
		if t.Strict {
			return ir.LessThan[F2, LogicalTerm[F2]](lhs, rhs)
		}
		//
		return ir.LessThanOrEquals[F2, LogicalTerm[F2]](lhs, rhs)
	case *Ite[F1]:
		var tb, fb LogicalTerm[F2]
		//
		cond := concretizeLogicalTerm[F1, F2](t.Condition)
		//
		if t.TrueBranch != nil {
			tb = concretizeLogicalTerm[F1, F2](t.TrueBranch)
		}
		//
		if t.FalseBranch != nil {
			fb = concretizeLogicalTerm[F1, F2](t.FalseBranch)
		}
		//
		return ir.IfThenElse(cond, tb, fb)
	case *Negate[F1]:
		return ir.Negation(concretizeLogicalTerm[F1, F2](t.Arg))
	case *NotEqual[F1]:
		lhs := concretizeTerm[F1, F2](t.Lhs)
		rhs := concretizeTerm[F1, F2](t.Rhs)
		//
		return ir.NotEquals[F2, LogicalTerm[F2]](lhs, rhs)
	default:
		panic("unreachable")
	}
}

func concretizeLogicalTerms[F1 Element[F1], F2 Element[F2]](terms []LogicalTerm[F1]) []LogicalTerm[F2] {
	var nterms = make([]LogicalTerm[F2], len(terms))
	//
	for i, t := range terms {
		nterms[i] = concretizeLogicalTerm[F1, F2](t)
	}
	//
	return nterms
}

// ============================================================================
// Terms
// ============================================================================

func concretizeTerm[F1 Element[F1], F2 Element[F2]](t Term[F1]) Term[F2] {
	var tmp F2
	//
	switch t := t.(type) {
	case *Add[F1]:
		return ir.Sum(concretizeTerms[F1, F2](t.Args)...)
	case *Cast[F1]:
		return ir.CastOf(concretizeTerm[F1, F2](t.Arg), t.BitWidth)
	case *Constant[F1]:
		// NOTE: could fail if  F1 value does not fit into F2 value.
		return ir.Const[F2, Term[F2]](tmp.SetBytes(t.Value.Bytes()))
	case *IfZero[F1]:
		cond := concretizeLogicalTerm[F1, F2](t.Condition)
		tb := concretizeTerm[F1, F2](t.TrueBranch)
		fb := concretizeTerm[F1, F2](t.FalseBranch)
		//
		return ir.IfElse(cond, tb, fb)
	case *LabelledConst[F1]:
		// NOTE: no need really to support labelled constants here.
		return ir.Const[F2, Term[F2]](tmp.SetBytes(t.Value.Bytes()))
	case *RegisterAccess[F1]:
		return ir.NewRegisterAccess[F2, Term[F2]](t.Register, t.Shift)
	case *Exp[F1]:
		return ir.Exponent(concretizeTerm[F1, F2](t.Arg), t.Pow)
	case *Mul[F1]:
		return ir.Product(concretizeTerms[F1, F2](t.Args)...)
	case *Norm[F1]:
		return ir.Normalise(concretizeTerm[F1, F2](t.Arg))
	case *Sub[F1]:
		return ir.Subtract(concretizeTerms[F1, F2](t.Args)...)
	case *VectorAccess[F1]:
		var nterms = make([]*RegisterAccess[F2], len(t.Vars))
		//
		for i, t := range t.Vars {
			nterms[i] = ir.RawRegisterAccess[F2, Term[F2]](t.Register, t.Shift)
		}
		//
		return ir.NewVectorAccess(nterms)
	default:
		panic("unreachable")
	}
}

func concretizeOptionalTerm[F1 Element[F1], F2 Element[F2]](t util.Option[Term[F1]]) util.Option[Term[F2]] {
	if t.IsEmpty() {
		return util.None[Term[F2]]()
	}
	//
	return util.Some(concretizeTerm[F1, F2](t.Unwrap()))
}

func concretizeTerms[F1 Element[F1], F2 Element[F2]](terms []Term[F1]) []Term[F2] {
	var nterms = make([]Term[F2], len(terms))
	//
	for i, t := range terms {
		nterms[i] = concretizeTerm[F1, F2](t)
	}
	//
	return nterms
}
