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
	"math/big"
	"reflect"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

type mirTerm = mir.Term[word.BigEndian]
type mirLogicalTerm = mir.LogicalTerm[word.BigEndian]

// LowerToMir lowers (or refines) an HIR schema into an MIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func LowerToMir(modules []Module) []mir.Module[word.BigEndian] {
	lowering := NewMirLowering(modules)
	//
	return lowering.Lower()
}

// MirLowering captures all auxiliary state required in the process of lowering
// modules from HIR to MIR.  This state is because, as part of the lowering
// process, we may introduce some number of additional modules (e.g. for
// managing type proofs).
type MirLowering struct {
	// Modules we are lowering from
	hirModules []Module
	// Modules we are lowering to
	mirSchema mir.SchemaBuilder[word.BigEndian]
}

// NewMirLowering constructs an initial state for lowering a given MIR schema.
func NewMirLowering(modules []Module) MirLowering {
	var (
		mirSchema = ir.NewSchemaBuilder[word.BigEndian, mir.Constraint[word.BigEndian], mirTerm, mir.Module[word.BigEndian]]()
	)
	// Initialise MIR modules
	for _, m := range modules {
		mirSchema.NewModule(m.Name(), m.LengthMultiplier(), m.AllowPadding(), m.IsPublic(), m.IsSynthetic())
	}
	//
	return MirLowering{
		modules,
		mirSchema,
	}
}

// Lower the MIR schema provide when this lowering instance was created into an
// equivalent MIR schema.
func (p *MirLowering) Lower() []mir.Module[word.BigEndian] {
	// Initialise modules
	for i := range len(p.hirModules) {
		p.InitialiseModule(uint(i))
	}
	// Lower modules
	for i := range len(p.hirModules) {
		p.LowerModule(uint(i))
	}
	// Build concrete modules from schema
	return ir.BuildSchema[mir.Module[word.BigEndian]](p.mirSchema)
}

// InitialiseModule simply initialises all registers within the module, but does
// not lower any constraint or assignments.
func (p *MirLowering) InitialiseModule(index uint) {
	var (
		hirModule = p.hirModules[index]
		mirModule = p.mirSchema.Module(index)
	)
	// Initialise registers in MIR module
	mirModule.NewRegisters(hirModule.Registers()...)
}

// LowerModule lowers the given MIR module into the corresponding MIR module.
// This includes all constraints and assignments.
func (p *MirLowering) LowerModule(index uint) {
	var (
		hirModule = p.hirModules[index]
		mirModule = p.mirSchema.Module(index)
	)
	// Lower assignments.
	for iter := hirModule.Assignments(); iter.HasNext(); {
		mirModule.AddAssignment(iter.Next())
	}
	// Lower constraints
	for iter := hirModule.Constraints(); iter.HasNext(); {
		// Following should always hold
		constraint := iter.Next().(Constraint)
		//
		p.lowerConstraint(constraint, mirModule)
	}
}

// Lower a constraint to the MIR level.
func (p *MirLowering) lowerConstraint(c Constraint, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	// Check what kind of constraint we have
	switch v := c.constraint.(type) {
	case Assertion:
		p.lowerAssertion(v, mirModule)
	case InterleavingConstraint:
		p.lowerInterleavingConstraint(v, mirModule)
	case LookupConstraint:
		p.lowerLookupConstraint(v, mirModule)
	case PermutationConstraint:
		p.lowerPermutationConstraint(v, mirModule)
	case RangeConstraint:
		p.lowerRangeConstraint(v, mirModule)
	case SortedConstraint:
		p.lowerSortedConstraint(v, mirModule)
	case VanishingConstraint:
		p.lowerVanishingConstraint(v, mirModule)
	default:
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

// Lowering an assertion is straightforward since its not a true constraint.
func (p *MirLowering) lowerAssertion(v Assertion, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	var term = p.lowerLogical(v.Property, mirModule)
	//
	mirModule.AddConstraint(mir.NewAssertion(v.Handle, v.Context, v.Domain, term))
}

// Lower a vanishing constraint to the MIR level.  This is relatively
// straightforward and simply relies on lowering the expression being
// constrained.
func (p *MirLowering) lowerVanishingConstraint(v VanishingConstraint, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	var term = p.lowerLogical(v.Constraint, mirModule)
	//
	mirModule.AddConstraint(
		mir.NewVanishingConstraint(v.Handle, v.Context, v.Domain, term))
}

// Lower a permutation constraint to the MIR level.  This is trivial because
// permutation constraints do not currently support complex forms.
func (p *MirLowering) lowerPermutationConstraint(v PermutationConstraint, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	mirModule.AddConstraint(
		mir.NewPermutationConstraint[word.BigEndian](v.Handle, v.Context, v.Targets, v.Sources),
	)
}

// Lower a range constraint to the MIR level.  Since range constraints at the
// MIR level can only access columns directly, we must expand the source
// expressions into computed columns with corresponding constraints.
func (p *MirLowering) lowerRangeConstraint(v RangeConstraint, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	var term = p.expandTerm(v.Expr, mirModule)
	//
	mirModule.AddConstraint(
		mir.NewRangeConstraint(v.Handle, v.Context, term, v.Bitwidth))
}

// Lower an interleaving constraint to the MIR level.  Since interleaving
// constraints at the MIR level can only access columns directly, we must expand
// the source expressions into computed columns with corresponding constraints.
func (p *MirLowering) lowerInterleavingConstraint(c InterleavingConstraint, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	//
	// Lower sources
	sources := p.expandTerms(c.Sources, mirModule)
	// Lower target
	target := p.expandTerm(c.Target, mirModule)
	// Add constraint
	mirModule.AddConstraint(
		mir.NewInterleavingConstraint(c.Handle, c.TargetContext, c.SourceContext, target, sources))
}

// Lower a lookup constraint to the MIR level.  Since lookup constraints at the
// MIR level can only access columns directly, we must expand the source
// expressions into computed columns with corresponding constraints.
func (p *MirLowering) lowerLookupConstraint(c LookupConstraint, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	var (
		sources = make([]lookup.Vector[word.BigEndian, mirTerm], len(c.Sources))
		targets = make([]lookup.Vector[word.BigEndian, mirTerm], len(c.Targets))
	)
	// Lower sources
	for i, ith := range c.Sources {
		sources[i] = p.lowerLookupVector(ith, mirModule)
	}
	// Lower targets
	for i, ith := range c.Targets {
		targets[i] = p.lowerLookupVector(ith, mirModule)
	}
	// Add constraint
	mirModule.AddConstraint(mir.NewLookupConstraint(c.Handle, targets, sources))
}

func (p *MirLowering) lowerLookupVector(vector lookup.Vector[word.BigEndian, Term], mirModule *mir.ModuleBuilder[word.BigEndian],
) lookup.Vector[word.BigEndian, mirTerm] {
	var (
		terms    = p.expandTerms(vector.Terms, mirModule)
		selector util.Option[mirTerm]
	)
	//
	if vector.HasSelector() {
		sel := p.expandTerm(vector.Selector.Unwrap(), mirModule)
		selector = util.Some[mirTerm](sel)
	}
	//
	return lookup.NewVector(vector.Module, selector, terms...)
}

// Lower a sorted constraint to the MIR level.  Since sorting constraints at the
// MIR level can only access columns directly, we must expand the source
// expressions into computed columns with corresponding constraints.
func (p *MirLowering) lowerSortedConstraint(c SortedConstraint, mirModule *mir.ModuleBuilder[word.BigEndian]) {
	var (
		terms    = p.expandTerms(c.Sources, mirModule)
		selector util.Option[mirTerm]
	)
	//
	if c.Selector.HasValue() {
		sel := p.expandTerm(c.Selector.Unwrap(), mirModule)
		selector = util.Some[mirTerm](sel)
	}
	// Add constraint
	mirModule.AddConstraint(
		mir.NewSortedConstraint(c.Handle, c.Context, c.BitWidth, selector, terms, c.Signs, c.Strict))
}

func (p *MirLowering) lowerLogical(e LogicalTerm, mirModule *mir.ModuleBuilder[word.BigEndian]) mirLogicalTerm {
	//
	switch e := e.(type) {
	case *Conjunct:
		return ir.Conjunction[word.BigEndian](p.lowerLogicals(e.Args, mirModule)...)
	case *Disjunct:
		return ir.Disjunction[word.BigEndian](p.lowerLogicals(e.Args, mirModule)...)
	case *Equal:
		return p.lowerEquality(true, e.Lhs, e.Rhs, mirModule)
	case *Ite:
		return p.lowerIte(e, mirModule)
	case *Negate:
		return ir.Negation[word.BigEndian](p.lowerLogical(e.Arg, mirModule))
	case *NotEqual:
		return p.lowerEquality(false, e.Lhs, e.Rhs, mirModule)
	case *Inequality:
		return p.lowerInequality(*e, mirModule)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func (p *MirLowering) lowerOptionalLogical(e LogicalTerm, mirModule *mir.ModuleBuilder[word.BigEndian]) mirLogicalTerm {
	if e == nil {
		return nil
	}
	//
	return p.lowerLogical(e, mirModule)
}

func (p *MirLowering) lowerLogicals(terms []LogicalTerm, mirModule *mir.ModuleBuilder[word.BigEndian],
) []mirLogicalTerm {
	//
	nexprs := make([]mirLogicalTerm, len(terms))

	for i := range len(terms) {
		nexprs[i] = p.lowerLogical(terms[i], mirModule)
	}

	return nexprs
}

func (p *MirLowering) lowerEquality(sign bool, left Term, right Term, mirModule *mir.ModuleBuilder[word.BigEndian],
) mirLogicalTerm {
	//
	var fn = func(lhs, rhs mirTerm) mirLogicalTerm {
		//
		if sign {
			return ir.Equals[word.BigEndian, mirLogicalTerm](lhs, rhs)
		}
		//
		return ir.NotEquals[word.BigEndian, mirLogicalTerm](lhs, rhs)
	}
	//
	return p.lowerBinaryLogical(left, right, fn, mirModule)
}

func (p *MirLowering) lowerInequality(term Inequality, mirModule *mir.ModuleBuilder[word.BigEndian],
) mirLogicalTerm {
	//
	var fn = func(lhs, rhs mirTerm) mirLogicalTerm {
		if term.Strict {
			return ir.LessThan[word.BigEndian, mirLogicalTerm](lhs, rhs)
		}
		//
		return ir.LessThanOrEquals[word.BigEndian, mirLogicalTerm](lhs, rhs)
	}
	//
	return p.lowerBinaryLogical(term.Lhs, term.Rhs, fn, mirModule)
}

func (p *MirLowering) lowerIte(term *Ite, mirModule *mir.ModuleBuilder[word.BigEndian]) mirLogicalTerm {
	var (
		condition   = p.lowerLogical(term.Condition, mirModule)
		trueBranch  = p.lowerOptionalLogical(term.TrueBranch, mirModule)
		falseBranch = p.lowerOptionalLogical(term.FalseBranch, mirModule)
	)
	//
	return ir.IfThenElse(condition, trueBranch, falseBranch)
}

func (p *MirLowering) lowerBinaryLogical(lhs, rhs Term, fn BinaryLogicalFn, mirModule *mir.ModuleBuilder[word.BigEndian],
) mirLogicalTerm {
	//
	var (
		lTerm = p.lowerTerm(lhs, mirModule)
		rTerm = p.lowerTerm(rhs, mirModule)
	)
	//
	return DisjunctIfTerms(fn, lTerm, rTerm)
}

func (p *MirLowering) expandTerms(es []Term, mirModule *mir.ModuleBuilder[word.BigEndian]) (terms []mirTerm) {
	//
	terms = make([]mirTerm, len(es))
	//
	for i, e := range es {
		terms[i] = p.expandTerm(e, mirModule)
	}
	//
	return terms
}

// Expand an arbitrary term into a column as necessary.  This is used to lower
// constraints by compiling out expressions, such that the lowered constraint
// only operates over column accesses (i.e. because this is the form required
// for the AIR layer used by the prover).  To do this, requires two pieces:
// first, the expression is evaluated using an assignment which stores the
// result into what is essentially a temporary column; second, a constraint is
// used to enforce the relationship between that column and the original
// expression.
func (p *MirLowering) expandTerm(e Term, module *mir.ModuleBuilder[word.BigEndian]) *mir.RegisterAccess[word.BigEndian] {
	// Check whether this really requires expansion (or not).
	if ca, ok := e.(*RegisterAccess); ok && ca.Shift == 0 {
		// No, expansion is not required
		return ir.RawRegisterAccess[word.BigEndian, mirTerm](ca.Register, ca.Shift)
	}
	// Yes, expansion is really necessary
	var (
		term = p.lowerTerm(e, module)
		// Determine bitwidth required for target register
		bitwidth = term.BitWidth(module)
		// Determine computed column name
		name = e.Lisp(true, module).String(false)
		// Look up column
		index, ok = module.HasRegister(name)
		// Default padding (for now)
		padding big.Int = ir.PaddingFor(e, module)
	)
	// Add new column (if it does not already exist)
	if !ok {
		// Convert expression into a generic computation
		computation := ir.NewComputation[word.BigEndian, LogicalTerm](e)
		// Declared a new computed column
		index = module.NewRegister(schema.NewComputedRegister(name, bitwidth, padding))
		// Add assignment for filling said computed column
		module.AddAssignment(
			assignment.NewComputedRegister(computation, true, module.Id(), index))
		// Construct v == [e]
		eq_e_v := term.Equate(index)
		// Ensure v == e, where v is value of computed column.
		module.AddConstraint(
			mir.NewVanishingConstraint(name, module.Id(), util.None[int](), eq_e_v))
	}
	// FIXME: eventually we just want to return the index
	return ir.RawRegisterAccess[word.BigEndian, mirTerm](index, 0)
}

// Lower a given HIR expression into one or more "conditional" MIR expressions.
// In the majority of cases, an HIR expression is lowered into a single
// unconditional MIR expression.  However, the presence of nested "if" terms
// introduces the need to separate out terms with conditions.  Consider the
// following minimal example:
//
// (== X
//
//	(if (== 0 Y) 0 7))
//
// In this case, we lower into (effectively the following two MIR conditional
// expressions:
//
// (1) (== X 0) when (== 0 Y)
// (2) (== X 7) when (!= 0 Y)
//
// The reason for doing this is to enable lowering to a polynomial expression
// (i.e. since these cannot contain nested normalisation expressions, etc).  An
// alternative approach would be to expand nested normalisations arising into
// inverse columns as they arise.  However, this was deemed to be less than
// desirable because it introduces products of the form (x*x⁻¹) which are
// expensive in the context of small fields.
func (p *MirLowering) lowerTerm(e Term, mirModule *mir.ModuleBuilder[word.BigEndian]) IfTerm {
	//
	switch e := e.(type) {
	case *Add:
		fn := func(args []mirTerm) mirTerm {
			return ir.Sum[word.BigEndian](args...)
		}
		//
		return p.lowerTerms(fn, mirModule, e.Args...)
	case *Cast:
		return p.lowerTerm(e.Arg, mirModule)
	case *Constant:
		return UnconditionalTerm(ir.Const[word.BigEndian, mirTerm](e.Value))
	case *RegisterAccess:
		return UnconditionalTerm(ir.NewRegisterAccess[word.BigEndian, mirTerm](e.Register, e.Shift))
	case *Exp:
		return p.lowerExpTo(e, mirModule)
	case *IfZero:
		condition := p.lowerLogical(e.Condition, mirModule)
		trueBranch := p.lowerTerm(e.TrueBranch, mirModule)
		falseBranch := p.lowerTerm(e.FalseBranch, mirModule)
		//
		return IfThenElse(condition, trueBranch, falseBranch)
	case *LabelledConst:
		return UnconditionalTerm(ir.Const[word.BigEndian, mirTerm](e.Value))
	case *Mul:
		fn := func(args []mirTerm) mirTerm {
			return ir.Product(args...)
		}
		//
		return p.lowerTerms(fn, mirModule, e.Args...)
	case *Norm:
		var (
			zero = ir.Const[word.BigEndian, mirTerm](field.Zero[word.BigEndian]())
			one  = ir.Const[word.BigEndian, mirTerm](field.One[word.BigEndian]())
			arg  = p.lowerTerm(e.Arg, mirModule)
		)
		//
		return IfEqElse(arg, zero, zero, one)
	case *Sub:
		fn := func(args []mirTerm) mirTerm {
			return ir.Subtract[word.BigEndian](args...)
		}
		//
		return p.lowerTerms(fn, mirModule, e.Args...)
	case *VectorAccess:
		return UnconditionalTerm(p.lowerVectorAccess(e))
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

// Lower a set of zero or more HIR expressions.
func (p *MirLowering) lowerTerms(fn NaryFn, mirModule *mir.ModuleBuilder[word.BigEndian], exprs ...Term) IfTerm {
	var nexprs = make([]IfTerm, len(exprs))
	//
	for i := range len(exprs) {
		nexprs[i] = p.lowerTerm(exprs[i], mirModule)
	}
	//
	return MapIfTerms(fn, nexprs...)
}

// LowerTo lowers an exponent expression to the MIR level by lowering the
// argument, and then constructing a multiplication.  This is because the AIR
// level does not support an explicit exponent operator.
func (p *MirLowering) lowerExpTo(e *Exp, mirModule *mir.ModuleBuilder[word.BigEndian]) IfTerm {
	var (
		// Lower expression being raised
		term = p.lowerTerm(e.Arg, mirModule)
	)
	//
	return term.Map(func(arg mirTerm) mirTerm {
		// Multiply it out k times
		es := make([]mirTerm, e.Pow)
		//
		for i := uint64(0); i < e.Pow; i++ {
			es[i] = arg
		}
		// Done
		return ir.Product[word.BigEndian](es...)
	})
}

func (p *MirLowering) lowerVectorAccess(e *VectorAccess) mirTerm {
	var (
		vars = make([]*ir.RegisterAccess[word.BigEndian, mirTerm], len(e.Vars))
	)
	//
	for i, v := range e.Vars {
		vars[i] = ir.RawRegisterAccess[word.BigEndian, mirTerm](v.Register, v.Shift)
	}
	//
	return ir.NewVectorAccess(vars)
}
