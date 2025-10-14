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
package hir

import (
	"fmt"
	"reflect"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

// LowerToMir lowers (or refines) an HIR schema into an MIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func LowerToMir[F field.Element[F]](modules []Module[F]) []mir.Module[F] {
	lowering := NewMirLowering(modules)
	//
	return lowering.Lower()
}

// MirLowering captures all auxiliary state required in the process of lowering
// modules from HIR to MIR.  This state is because, as part of the lowering
// process, we may introduce some number of additional modules (e.g. for
// managing type proofs).
type MirLowering[F field.Element[F]] struct {
	// Modules we are lowering from
	hirModules []Module[F]
	// Modules we are lowering to
	mirSchema mir.SchemaBuilder[F]
}

// NewMirLowering constructs an initial state for lowering a given MIR schema.
func NewMirLowering[F field.Element[F]](modules []Module[F]) MirLowering[F] {
	var (
		mirSchema = ir.NewSchemaBuilder[F, mir.Constraint[F], mir.Term[F], mir.Module[F]]()
	)
	// Initialise MIR modules
	for _, m := range modules {
		mirSchema.NewModule(m.Name(), m.LengthMultiplier(), m.AllowPadding(), m.IsPublic(), m.IsSynthetic())
	}
	//
	return MirLowering[F]{
		modules,
		mirSchema,
	}
}

// Lower the MIR schema provide when this lowering instance was created into an
// equivalent MIR schema.
func (p *MirLowering[F]) Lower() []mir.Module[F] {
	// Initialise modules
	for i := range len(p.hirModules) {
		p.InitialiseModule(uint(i))
	}
	// Lower modules
	for i := range len(p.hirModules) {
		p.LowerModule(uint(i))
	}
	// Build concrete modules from schema
	return ir.BuildSchema[mir.Module[F]](p.mirSchema)
}

// InitialiseModule simply initialises all registers within the module, but does
// not lower any constraint or assignments.
func (p *MirLowering[F]) InitialiseModule(index uint) {
	var (
		hirModule = p.hirModules[index]
		mirModule = p.mirSchema.Module(index)
	)
	// Initialise registers in MIR module
	mirModule.NewRegisters(hirModule.Registers()...)
}

// LowerModule lowers the given MIR module into the corresponding MIR module.
// This includes all constraints and assignments.
func (p *MirLowering[F]) LowerModule(index uint) {
	var (
		hirModule = p.hirModules[index]
		mirModule = p.mirSchema.Module(index)
	)
	// Add assignments.  At this time, there is nothing to do in terms of
	// lowering.  Observe that this must be done *before* lowering assignments
	// to ensure a correct ordering.  For example, if a constraint refers to one
	// of these assigned columns and generates a corresponding computed column
	// (e.g. for the inverse).
	for iter := hirModule.Assignments(); iter.HasNext(); {
		mirModule.AddAssignment(iter.Next())
	}
	// Lower constraints
	for iter := hirModule.Constraints(); iter.HasNext(); {
		// Following should always hold
		constraint := iter.Next().(Constraint[F])
		//
		p.lowerConstraint(constraint, mirModule)
	}
}

// Lower a constraint to the MIR level.
func (p *MirLowering[F]) lowerConstraint(c Constraint[F], mirModule *mir.ModuleBuilder[F]) {
	// Check what kind of constraint we have
	switch v := c.constraint.(type) {
	case Assertion[F]:
		p.lowerAssertion(v, mirModule)
	case InterleavingConstraint[F]:
		p.lowerInterleavingConstraint(v, mirModule)
	case LookupConstraint[F]:
		p.lowerLookupConstraint(v, mirModule)
	case PermutationConstraint[F]:
		p.lowerPermutationConstraint(v, mirModule)
	case RangeConstraint[F]:
		p.lowerRangeConstraint(v, mirModule)
	case SortedConstraint[F]:
		p.lowerSortedConstraint(v, mirModule)
	case VanishingConstraint[F]:
		p.lowerVanishingConstraint(v, mirModule)
	default:
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

// Lowering an assertion is straightforward since its not a true constraint.
func (p *MirLowering[F]) lowerAssertion(v Assertion[F], mirModule *mir.ModuleBuilder[F]) {
	var term = p.lowerLogical(v.Property, mirModule)
	//
	mirModule.AddConstraint(mir.NewAssertion(v.Handle, v.Context, v.Domain, term))
}

// Lower a vanishing constraint to the MIR level.  This is relatively
// straightforward and simply relies on lowering the expression being
// constrained.
func (p *MirLowering[F]) lowerVanishingConstraint(v VanishingConstraint[F], mirModule *mir.ModuleBuilder[F]) {
	var term = p.lowerLogical(v.Constraint, mirModule)
	//
	mirModule.AddConstraint(
		mir.NewVanishingConstraint(v.Handle, v.Context, v.Domain, term))
}

// Lower a permutation constraint to the MIR level.  This is trivial because
// permutation constraints do not currently support complex forms.
func (p *MirLowering[F]) lowerPermutationConstraint(v PermutationConstraint[F], mirModule *mir.ModuleBuilder[F]) {
	mirModule.AddConstraint(
		mir.NewPermutationConstraint[F](v.Handle, v.Context, v.Targets, v.Sources),
	)
}

// Lower a range constraint to the MIR level.  Since range constraints at the
// MIR level can only access columns directly, we must expand the source
// expressions into computed columns with corresponding constraints.
func (p *MirLowering[F]) lowerRangeConstraint(v RangeConstraint[F], mirModule *mir.ModuleBuilder[F]) {
	var term = p.lowerTerm(v.Expr, mirModule)
	//
	mirModule.AddConstraint(
		mir.NewRangeConstraint(v.Handle, v.Context, term, v.Bitwidth))
}

// Lower an interleaving constraint to the MIR level.  Since interleaving
// constraints at the MIR level can only access columns directly, we must expand
// the source expressions into computed columns with corresponding constraints.
func (p *MirLowering[F]) lowerInterleavingConstraint(c InterleavingConstraint[F], mirModule *mir.ModuleBuilder[F]) {
	//
	// Lower sources
	sources := p.lowerTerms(c.Sources, mirModule)
	// Lower target
	target := p.lowerTerm(c.Target, mirModule)
	// Add constraint
	mirModule.AddConstraint(
		mir.NewInterleavingConstraint(c.Handle, c.TargetContext, c.SourceContext, target, sources))
}

// Lower a lookup constraint to the MIR level.  Since lookup constraints at the
// MIR level can only access columns directly, we must expand the source
// expressions into computed columns with corresponding constraints.
func (p *MirLowering[F]) lowerLookupConstraint(c LookupConstraint[F], mirModule *mir.ModuleBuilder[F]) {
	var (
		sources = make([]lookup.Vector[F, mir.Term[F]], len(c.Sources))
		targets = make([]lookup.Vector[F, mir.Term[F]], len(c.Targets))
	)
	// Lower sources
	for i, ith := range c.Sources {
		sources[i] = p.lowerLookupVectorToAir(ith, mirModule)
	}
	// Lower targets
	for i, ith := range c.Targets {
		targets[i] = p.lowerLookupVectorToAir(ith, mirModule)
	}
	// Add constraint
	mirModule.AddConstraint(mir.NewLookupConstraint(c.Handle, targets, sources))
}

func (p *MirLowering[F]) lowerLookupVectorToAir(vector lookup.Vector[F, Term[F]], mirModule *mir.ModuleBuilder[F],
) lookup.Vector[F, mir.Term[F]] {
	var (
		terms    = p.lowerTerms(vector.Terms, mirModule)
		selector util.Option[mir.Term[F]]
	)
	//
	if vector.HasSelector() {
		sel := p.lowerTerm(vector.Selector.Unwrap(), mirModule)
		selector = util.Some(sel)
	}
	//
	return lookup.NewVector(vector.Module, selector, terms...)
}

// Lower a sorted constraint to the MIR level.  Since sorting constraints at the
// MIR level can only access columns directly, we must expand the source
// expressions into computed columns with corresponding constraints.
func (p *MirLowering[F]) lowerSortedConstraint(c SortedConstraint[F], mirModule *mir.ModuleBuilder[F]) {
	var (
		terms    = p.lowerTerms(c.Sources, mirModule)
		selector util.Option[mir.Term[F]]
	)
	//
	if c.Selector.HasValue() {
		sel := p.lowerTerm(c.Selector.Unwrap(), mirModule)
		selector = util.Some(sel)
	}
	// Add constraint
	mirModule.AddConstraint(
		mir.NewSortedConstraint(c.Handle, c.Context, c.BitWidth, selector, terms, c.Signs, c.Strict))
}

func (p *MirLowering[F]) lowerLogical(e LogicalTerm[F], mirModule *mir.ModuleBuilder[F]) mir.LogicalTerm[F] {
	//
	switch e := e.(type) {
	case *Conjunct[F]:
		return ir.Conjunction(p.lowerLogicals(e.Args, mirModule)...)
	case *Disjunct[F]:
		return ir.Disjunction(p.lowerLogicals(e.Args, mirModule)...)
	case *Equal[F]:
		return p.lowerEquality(true, e.Lhs, e.Rhs, mirModule)
	case *Ite[F]:
		return p.lowerIte(e, mirModule)
	case *Negate[F]:
		return ir.Negation(p.lowerLogical(e.Arg, mirModule))
	case *NotEqual[F]:
		return p.lowerEquality(false, e.Lhs, e.Rhs, mirModule)
	case *Inequality[F]:
		return p.lowerInequality(*e, mirModule)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func (p *MirLowering[F]) lowerOptionalLogical(e LogicalTerm[F], mirModule *mir.ModuleBuilder[F]) mir.LogicalTerm[F] {
	if e == nil {
		return nil
	}
	//
	return p.lowerLogical(e, mirModule)
}

func (p *MirLowering[F]) lowerLogicals(terms []LogicalTerm[F], mirModule *mir.ModuleBuilder[F],
) []mir.LogicalTerm[F] {
	//
	nexprs := make([]mir.LogicalTerm[F], len(terms))

	for i := range len(terms) {
		nexprs[i] = p.lowerLogical(terms[i], mirModule)
	}

	return nexprs
}

func (p *MirLowering[F]) lowerEquality(sign bool, left Term[F], right Term[F], mirModule *mir.ModuleBuilder[F],
) mir.LogicalTerm[F] {
	//
	var (
		lhs mir.Term[F] = p.lowerTerm(left, mirModule)
		rhs mir.Term[F] = p.lowerTerm(right, mirModule)
	)
	//
	if sign {
		return ir.Equals[F, mir.LogicalTerm[F]](lhs, rhs)
	}
	//
	return ir.NotEquals[F, mir.LogicalTerm[F]](lhs, rhs)
}

func (p *MirLowering[F]) lowerInequality(term Inequality[F], mirModule *mir.ModuleBuilder[F],
) mir.LogicalTerm[F] {
	//
	var (
		lhs mir.Term[F] = p.lowerTerm(term.Lhs, mirModule)
		rhs mir.Term[F] = p.lowerTerm(term.Rhs, mirModule)
	)
	//
	if term.Strict {
		return ir.LessThan[F, mir.LogicalTerm[F]](lhs, rhs)
	}
	//
	return ir.LessThanOrEquals[F, mir.LogicalTerm[F]](lhs, rhs)
}

func (p *MirLowering[F]) lowerIte(term *Ite[F], mirModule *mir.ModuleBuilder[F]) mir.LogicalTerm[F] {
	var (
		condition   = p.lowerLogical(term.Condition, mirModule)
		trueBranch  = p.lowerOptionalLogical(term.TrueBranch, mirModule)
		falseBranch = p.lowerOptionalLogical(term.FalseBranch, mirModule)
	)
	//
	return ir.IfThenElse(condition, trueBranch, falseBranch)
}

// Inner form is used for recursive calls and does not repeat the constant
// propagation phase.
func (p *MirLowering[F]) lowerTerm(e Term[F], mirModule *mir.ModuleBuilder[F]) mir.Term[F] {
	//
	switch e := e.(type) {
	case *Add[F]:
		args := p.lowerTerms(e.Args, mirModule)
		return ir.Sum(args...)
	case *Cast[F]:
		return p.lowerTerm(e.Arg, mirModule)
	case *Constant[F]:
		return ir.Const[F, mir.Term[F]](e.Value)
	case *RegisterAccess[F]:
		return ir.NewRegisterAccess[F, mir.Term[F]](e.Register, e.Shift)
	case *Exp[F]:
		arg := p.lowerTerm(e.Arg, mirModule)
		return ir.Exponent(arg, e.Pow)
	case *IfZero[F]:
		condition := p.lowerLogical(e.Condition, mirModule)
		trueBranch := p.lowerTerm(e.TrueBranch, mirModule)
		falseBranch := p.lowerTerm(e.FalseBranch, mirModule)
		//
		return ir.IfElse(condition, trueBranch, falseBranch)
	case *LabelledConst[F]:
		return ir.Const[F, mir.Term[F]](e.Value)
	case *Mul[F]:
		args := p.lowerTerms(e.Args, mirModule)
		return ir.Product(args...)
	case *Norm[F]:
		arg := p.lowerTerm(e.Arg, mirModule)
		return ir.Normalise(arg)
	case *Sub[F]:
		args := p.lowerTerms(e.Args, mirModule)
		return ir.Subtract(args...)
	case *VectorAccess[F]:
		return p.lowerVectorAccess(e)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

// Lower a set of zero or more HIR expressions.
func (p *MirLowering[F]) lowerTerms(exprs []Term[F], mirModule *mir.ModuleBuilder[F]) []mir.Term[F] {
	nexprs := make([]mir.Term[F], len(exprs))

	for i := range len(exprs) {
		nexprs[i] = p.lowerTerm(exprs[i], mirModule)
	}

	return nexprs
}

func (p *MirLowering[F]) lowerVectorAccess(e *VectorAccess[F]) mir.Term[F] {
	var (
		vars = make([]*ir.RegisterAccess[F, mir.Term[F]], len(e.Vars))
	)
	//
	for i, v := range e.Vars {
		vars[i] = ir.RawRegisterAccess[F, mir.Term[F]](v.Register, v.Shift)
	}
	//
	return ir.NewVectorAccess(vars)
}
