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
	"math"
	"reflect"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	air_gadgets "github.com/consensys/go-corset/pkg/ir/air/gadgets"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	util_math "github.com/consensys/go-corset/pkg/util/math"
)

// LowerToAir lowers (or refines) an MIR schema into an AIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func LowerToAir(schema Schema, config OptimisationConfig) air.Schema {
	lowering := NewAirLowering(schema)
	// Configure optimisations
	lowering.ConfigureOptimisation(config)
	//
	return lowering.Lower()
}

// AirLowering captures all auxiliary state required in the process of lowering
// modules from MIR to AIR.  This state is because, as part of the lowering
// process, we may introduce some number of additional modules (e.g. for
// managing type proofs).
type AirLowering struct {
	config OptimisationConfig
	// Modules we are lowering from
	mirSchema Schema
	// Modules we are lowering to
	airSchema air.SchemaBuilder
}

// NewAirLowering constructs an initial state for lowering a given MIR schema.
func NewAirLowering(mirSchema Schema) AirLowering {
	var (
		airSchema = ir.NewSchemaBuilder[bls12_377.Element, air.Constraint, air.Term, schema.Module[bls12_377.Element]]()
	)
	// Initialise AIR modules
	for _, m := range mirSchema.RawModules() {
		airSchema.NewModule(m.Name(), m.LengthMultiplier(), m.AllowPadding())
	}
	//
	return AirLowering{
		DEFAULT_OPTIMISATION_LEVEL,
		mirSchema,
		airSchema,
	}
}

// ConfigureOptimisation configures the amount of optimisation to apply during
// the lowering process.
func (p *AirLowering) ConfigureOptimisation(config OptimisationConfig) {
	p.config = config
}

// Lower the MIR schema provide when this lowering instance was created into an
// equivalent AIR schema.
func (p *AirLowering) Lower() air.Schema {
	// Initialise modules
	for i := 0; i < int(p.mirSchema.Width()); i++ {
		p.InitialiseModule(uint(i))
	}
	// Lower modules
	for i := 0; i < int(p.mirSchema.Width()); i++ {
		p.LowerModule(uint(i))
	}
	// Done
	return schema.NewUniformSchema(p.airSchema.Build())
}

// InitialiseModule simply initialises all registers within the module, but does
// not lower any constraint or assignments.
func (p *AirLowering) InitialiseModule(index uint) {
	var (
		mirModule = p.mirSchema.Module(index)
		airModule = p.airSchema.Module(index)
	)
	// Initialise registers in AIR module
	airModule.NewRegisters(mirModule.Registers()...)
}

// LowerModule lowers the given MIR module into the correspondind AIR module.
// This includes all constraints and assignments.
func (p *AirLowering) LowerModule(index uint) {
	var (
		mirModule = p.mirSchema.Module(index)
		airModule = p.airSchema.Module(index)
	)
	// Add assignments.  At this time, there is nothing to do in terms of
	// lowering.  Observe that this must be done *before* lowering assignments
	// to ensure a correct ordering.  For example, if a constraint refers to one
	// of these assigned columns and generates a corresponding computed column
	// (e.g. for the inverse).
	for iter := mirModule.Assignments(); iter.HasNext(); {
		airModule.AddAssignment(iter.Next())
	}
	// Lower constraints
	for iter := mirModule.Constraints(); iter.HasNext(); {
		// Following should always hold
		constraint := iter.Next().(Constraint)
		//
		p.lowerConstraintToAir(constraint, airModule)
	}
}

// Lower a constraint to the AIR level.
func (p *AirLowering) lowerConstraintToAir(c Constraint, airModule *air.ModuleBuilder) {
	// Check what kind of constraint we have
	switch v := c.constraint.(type) {
	case Assertion:
		p.lowerAssertionToAir(v, airModule)
	case InterleavingConstraint:
		p.lowerInterleavingConstraintToAir(v, airModule)
	case LookupConstraint:
		p.lowerLookupConstraintToAir(v, airModule)
	case PermutationConstraint:
		p.lowerPermutationConstraintToAir(v, airModule)
	case RangeConstraint:
		p.lowerRangeConstraintToAir(v, airModule)
	case SortedConstraint:
		p.lowerSortedConstraintToAir(v, airModule)
	case VanishingConstraint:
		p.lowerVanishingConstraintToAir(v, airModule)
	default:
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

// Lowering an assertion is straightforward since its not a true constraint.
func (p *AirLowering) lowerAssertionToAir(v Assertion, airModule *air.ModuleBuilder) {
	airModule.AddConstraint(air.NewAssertion(v.Handle, v.Context, v.Property))
}

// Lower a vanishing constraint to the AIR level.  This is relatively
// straightforward and simply relies on lowering the expression being
// constrained.  This may result in the generation of computed columns, e.g. to
// hold inverses, etc.
func (p *AirLowering) lowerVanishingConstraintToAir(v VanishingConstraint, airModule *air.ModuleBuilder) {
	//
	var (
		terms = p.lowerAndSimplifyLogicalTo(v.Constraint, airModule)
	)
	//
	for i, air_expr := range terms {
		// // Check whether this is a constant
		// constant := air_expr.AsConstant()
		// // Check for compile-time constants
		// if constant != nil && !constant.IsZero() {
		// 	panic(fmt.Sprintf("constraint %s cannot vanish!", v.Handle))
		// } else if constant == nil {
		// Construct suitable handle to distinguish this case
		handle := fmt.Sprintf("%s#%d", v.Handle, i)
		// Add constraint
		airModule.AddConstraint(
			air.NewVanishingConstraint(handle, v.Context, v.Domain, air_expr))
	}
}

// Lower a permutation constraint to the AIR level.  This is trivial because
// permutation constraints do not currently support complex forms.
func (p *AirLowering) lowerPermutationConstraintToAir(v PermutationConstraint, airModule *air.ModuleBuilder) {
	airModule.AddConstraint(
		air.NewPermutationConstraint(v.Handle, v.Context, v.Targets, v.Sources),
	)
}

// Lower a range constraint to the AIR level.  The challenge here is that a
// range constraint at the AIR level cannot use arbitrary expressions; rather it
// can only constrain columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func (p *AirLowering) lowerRangeConstraintToAir(v RangeConstraint, airModule *air.ModuleBuilder) {
	var (
		mirModule        = p.mirSchema.Module(v.Context)
		valRange         = v.Expr.ValueRange(mirModule)
		bitwidth, signed = valRange.BitWidth()
	)
	// Sanity check bitwidth result
	if signed {
		// We can't determine a suitable bitwidth, so it should be the maximum
		// value for the underlying field.
		bitwidth = math.MaxUint
	}
	// Lower target expression
	target := p.lowerAndSimplifyTermTo(v.Expr, airModule)
	// Expand target expression (if necessary)
	register := air_gadgets.Expand(bitwidth, target, airModule)
	// Apply bitwidth gadget
	ref := schema.NewRegisterRef(airModule.Id(), register)
	// Construct gadget
	gadget := air_gadgets.NewBitwidthGadget(&p.airSchema).
		WithLimitless(p.config.LimitlessTypeProofs).
		WithMaxRangeConstraint(p.config.MaxRangeConstraint)
	//
	gadget.Constrain(ref, v.Bitwidth)
}

// Lower an interleaving constraint to the AIR level.  The challenge here is
// that interleaving constraints at the AIR level cannot use arbitrary
// expressions; rather, they can only access columns directly.  Therefore,
// whenever a general expression is encountered, we must generate a computed
// column to hold the value of that expression, along with appropriate
// constraints to enforce the expected value.
func (p *AirLowering) lowerInterleavingConstraintToAir(c InterleavingConstraint, airModule *air.ModuleBuilder) {
	// Lower sources
	sources := p.expandTerms(c.SourceContext, c.Sources...)
	// Lower target
	target := p.expandTerms(c.SourceContext, c.Target)[0]
	// Add constraint
	airModule.AddConstraint(
		air.NewInterleavingConstraint(c.Handle, c.TargetContext, c.SourceContext, *target, sources))
}

// Lower a lookup constraint to the AIR level.  The challenge here is that
// lookup constraints at the AIR level cannot use arbitrary expressions; rather,
// they can only access columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func (p *AirLowering) lowerLookupConstraintToAir(c LookupConstraint, airModule *air.ModuleBuilder) {
	var (
		sources = make([]lookup.Vector[bls12_377.Element, *air.ColumnAccess], len(c.Sources))
		targets = make([]lookup.Vector[bls12_377.Element, *air.ColumnAccess], len(c.Targets))
	)
	// Lower sources
	for i, ith := range c.Sources {
		sources[i] = p.expandLookupVectorToAir(ith)
	}
	// Lower targets
	for i, ith := range c.Targets {
		targets[i] = p.expandLookupVectorToAir(ith)
	}
	// Add constraint
	airModule.AddConstraint(air.NewLookupConstraint(c.Handle, targets, sources))
}

func (p *AirLowering) expandLookupVectorToAir(vector lookup.Vector[bls12_377.Element, Term],
) lookup.Vector[bls12_377.Element, *air.ColumnAccess] {
	var (
		terms    = p.expandTerms(vector.Module, vector.Terms...)
		selector util.Option[*air.ColumnAccess]
	)
	//
	if vector.HasSelector() {
		sel := p.expandTerm(vector.Module, vector.Selector.Unwrap())
		selector = util.Some(sel)
	}
	//
	return lookup.NewVector(vector.Module, selector, terms...)
}

// Lower a sorted constraint to the AIR level.  The challenge here is that there
// is not concept of sorting constraints at the AIR level.  Instead, we have to
// generate the necessary machinery to enforce the sorting constraint.
func (p *AirLowering) lowerSortedConstraintToAir(c SortedConstraint, airModule *air.ModuleBuilder) {
	sources := make([]schema.RegisterId, len(c.Sources))
	//
	for i := 0; i < len(sources); i++ {
		var (
			ith                    = c.Sources[i]
			ithRange               = ith.ValueRange(airModule)
			sourceBitwidth, signed = ithRange.BitWidth()
		)
		// Sanity check
		if signed {
			panic(fmt.Sprintf("signed expansion encountered (%s)", ith.Lisp(false, airModule).String(true)))
		}
		// Lower source expression
		source := p.lowerTermTo(c.Sources[i], airModule)
		// Expand them
		sources[i] = air_gadgets.Expand(sourceBitwidth, source, airModule)
	}
	// Determine number of ordered columns
	numSignedCols := len(c.Signs)
	// finally add the sorting constraint
	gadget := air_gadgets.NewLexicographicSortingGadget(c.Handle, sources, c.BitWidth).
		WithSigns(c.Signs...).
		WithStrictness(c.Strict).
		WithLimitless(p.config.LimitlessTypeProofs).
		WithMaxRangeConstraint(p.config.MaxRangeConstraint)
	// Add (optional) selector
	if c.Selector.HasValue() {
		selector := p.lowerTermTo(c.Selector.Unwrap(), airModule)
		gadget.WithSelector(selector)
	}
	// Done
	gadget.Apply(airModule.Id(), &p.airSchema)
	// Sanity check bitwidth
	bitwidth := uint(0)

	for i := 0; i < numSignedCols; i++ {
		// Extract bitwidth of ith column
		ith := airModule.Register(sources[i]).Width
		if ith > bitwidth {
			bitwidth = ith
		}
	}
	//
	if bitwidth != c.BitWidth {
		// Should be unreachable.
		msg := fmt.Sprintf("incompatible bitwidths (%d vs %d)", bitwidth, c.BitWidth)
		panic(msg)
	}
}

func (p *AirLowering) expandTerms(context schema.ModuleId, terms ...Term) []*air.ColumnAccess {
	var nterms = make([]*air.ColumnAccess, len(terms))
	//
	for i, ith := range terms {
		nterms[i] = p.expandTerm(context, ith)
	}
	//
	return nterms
}

func (p *AirLowering) expandTerm(context schema.ModuleId, term Term) *air.ColumnAccess {
	var (
		airModule = p.airSchema.Module(context)
	)
	//
	var source_register schema.RegisterId
	//
	sourceRange := term.ValueRange(airModule)
	sourceBitwidth, signed := sourceRange.BitWidth()
	//
	if signed {
		panic(fmt.Sprintf("signed expansion encountered (%s)", term.Lisp(false, airModule).String(true)))
	}
	// Lower source expressions
	source := p.lowerAndSimplifyTermTo(term, airModule)
	// Expand them
	source_register = air_gadgets.Expand(sourceBitwidth, source, airModule)
	//
	return ir.RawRegisterAccess[bls12_377.Element, air.Term](source_register, 0)
}

func (p *AirLowering) lowerAndSimplifyLogicalTo(term LogicalTerm,
	airModule *air.ModuleBuilder) []air.Term {
	// Apply all reasonable simplifications
	term = term.Simplify(false)
	// Lower properly
	return simplify(p.lowerLogicalTo(true, term, airModule))
}

func (p *AirLowering) lowerLogicalTo(sign bool, e LogicalTerm, airModule *air.ModuleBuilder) []air.Term {
	//
	switch e := e.(type) {
	case *Conjunct:
		return p.lowerConjunctionTo(sign, e, airModule)
	case *Disjunct:
		return p.lowerDisjunctionTo(sign, e, airModule)
	case *Equal:
		return p.lowerEqualityTo(sign, e.Lhs, e.Rhs, airModule)
	case *Ite:
		return p.lowerIteTo(sign, e, airModule)
	case *Negate:
		return p.lowerLogicalTo(!sign, e.Arg, airModule)
	case *NotEqual:
		return p.lowerEqualityTo(!sign, e.Lhs, e.Rhs, airModule)
	case *Inequality:
		panic("inequalities cannot (currently) be lowered to AIR")
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func (p *AirLowering) lowerLogicalsTo(sign bool, airModule *air.ModuleBuilder, terms ...LogicalTerm) [][]air.Term {
	nexprs := make([][]air.Term, len(terms))

	for i := range len(terms) {
		nexprs[i] = p.lowerLogicalTo(sign, terms[i], airModule)
	}

	return nexprs
}

func (p *AirLowering) lowerConjunctionTo(sign bool, e *Conjunct, airModule *air.ModuleBuilder) []air.Term {
	var terms = p.lowerLogicalsTo(sign, airModule, e.Args...)
	//
	if sign {
		return conjunction(terms...)
	}
	//
	return disjunction(terms...)
}

func (p *AirLowering) lowerDisjunctionTo(sign bool, e *Disjunct, airModule *air.ModuleBuilder) []air.Term {
	var terms = p.lowerLogicalsTo(sign, airModule, e.Args...)
	//
	if sign {
		return disjunction(terms...)
	}
	//
	return conjunction(terms...)
}

func (p *AirLowering) lowerEqualityTo(sign bool, left Term, right Term, airModule *air.ModuleBuilder) []air.Term {
	//
	var (
		lhs air.Term = p.lowerTermTo(left, airModule)
		rhs air.Term = p.lowerTermTo(right, airModule)
		eq           = ir.Subtract(lhs, rhs)
	)
	//
	if sign {
		return []air.Term{eq}
	}
	//
	one := ir.Const64[bls12_377.Element, air.Term](1)
	// construct norm(eq)
	norm_eq := p.normalise(eq, airModule)
	// construct 1 - norm(eq)
	return []air.Term{ir.Subtract(one, norm_eq)}
}

func (p *AirLowering) lowerIteTo(sign bool, e *Ite, airModule *air.ModuleBuilder) []air.Term {
	if sign {
		return p.lowerPositiveIteTo(e, airModule)
	}
	//
	return p.lowerNegativeIteTo(e, airModule)
}

func (p *AirLowering) lowerPositiveIteTo(e *Ite, airModule *air.ModuleBuilder) []air.Term {
	var (
		terms []air.Term
	)
	// NOTE: using extractNormalisedCondition could be useful here.
	if e.TrueBranch != nil && e.FalseBranch != nil {
		trueCondition := p.lowerLogicalTo(true, e.Condition, airModule)
		falseCondition := p.lowerLogicalTo(false, e.Condition, airModule)
		trueBranch := p.lowerLogicalTo(true, e.TrueBranch, airModule)
		falseBranch := p.lowerLogicalTo(true, e.FalseBranch, airModule)
		// Check whether optimisation is possible
		if len(trueCondition) == 1 && len(falseCondition) == 1 &&
			len(falseBranch) == 1 && len(trueBranch) == 1 {
			// Yes, its safe to apply.
			fb := ir.Product(trueCondition[0], falseBranch[0])
			tb := ir.Product(falseCondition[0], trueBranch[0])
			//
			return []air.Term{ir.Sum(tb, fb)}
		}
		// No, optimisation does not apply
		terms = append(terms, disjunction(falseCondition, trueBranch)...)
		terms = append(terms, disjunction(trueCondition, falseBranch)...)
	} else if e.TrueBranch != nil {
		falseCondition := p.lowerLogicalTo(false, e.Condition, airModule)
		trueBranch := p.lowerLogicalTo(true, e.TrueBranch, airModule)
		terms = append(terms, disjunction(falseCondition, trueBranch)...)
	} else if e.FalseBranch != nil {
		trueCondition := p.lowerLogicalTo(true, e.Condition, airModule)
		falseBranch := p.lowerLogicalTo(true, e.FalseBranch, airModule)
		terms = append(terms, disjunction(trueCondition, falseBranch)...)
	}
	//
	return terms
}

// !ITE(A,B,C) => !((!A||B) && (A||C))
//
//	=> !(!A||B) || !(A||C)
//	=> (A&&!B) || (!A&&!C)
func (p *AirLowering) lowerNegativeIteTo(e *Ite, airModule *air.ModuleBuilder) []air.Term {
	// NOTE: using extractNormalisedCondition could be useful here.
	var (
		terms [][]air.Term
	)
	//
	if e.TrueBranch != nil {
		trueCondition := p.lowerLogicalTo(true, e.Condition, airModule)
		notTrueBranch := p.lowerLogicalTo(false, e.TrueBranch, airModule)
		terms = append(terms, conjunction(trueCondition, notTrueBranch))
	}
	//
	if e.FalseBranch != nil {
		falseCondition := p.lowerLogicalTo(false, e.Condition, airModule)
		notFalseBranch := p.lowerLogicalTo(false, e.FalseBranch, airModule)
		terms = append(terms, conjunction(falseCondition, notFalseBranch))
	}
	//
	return disjunction(terms...)
}

// Lower an expression into the Arithmetic Intermediate Representation.
// Essentially, this means eliminating normalising expressions by introducing
// new columns into the given table (with appropriate constraints).  This first
// performs constant propagation to ensure lowering is as efficient as possible.
// A module identifier is required to determine where any computed columns
// should be located.
func (p *AirLowering) lowerAndSimplifyTermTo(term Term, airModule *air.ModuleBuilder) air.Term {
	// Optimise normalisations
	term = eliminateNormalisationInTerm(term, airModule, p.config)
	// Apply all reasonable simplifications
	term = term.Simplify(false)
	// Lower properly
	return p.lowerTermTo(term, airModule)
}

// Inner form is used for recursive calls and does not repeat the constant
// propagation phase.
func (p *AirLowering) lowerTermTo(e Term, airModule *air.ModuleBuilder) air.Term {
	//
	switch e := e.(type) {
	case *Add:
		args := p.lowerTerms(e.Args, airModule)
		return ir.Sum(args...)
	case *Cast:
		return p.lowerTermTo(e.Arg, airModule)
	case *Constant:
		return ir.Const[bls12_377.Element, air.Term](e.Value)
	case *RegisterAccess:
		return ir.NewRegisterAccess[bls12_377.Element, air.Term](e.Register, e.Shift)
	case *Exp:
		return p.lowerExpTo(e, airModule)
	case *IfZero:
		return p.lowerIfZeroTo(e, airModule)
	case *LabelledConst:
		return ir.Const[bls12_377.Element, air.Term](e.Value)
	case *Mul:
		args := p.lowerTerms(e.Args, airModule)
		return ir.Product(args...)
	case *Norm:
		return p.lowerNormTo(e, airModule)
	case *Sub:
		args := p.lowerTerms(e.Args, airModule)
		return ir.Subtract(args...)
	case *VectorAccess:
		return p.lowerVectorAccess(e, airModule)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

// Lower a set of zero or more MIR expressions.
func (p *AirLowering) lowerTerms(exprs []Term, airModule *air.ModuleBuilder) []air.Term {
	nexprs := make([]air.Term, len(exprs))

	for i := range len(exprs) {
		nexprs[i] = p.lowerTermTo(exprs[i], airModule)
	}

	return nexprs
}

// LowerTo lowers an exponent expression to the AIR level by lowering the
// argument, and then constructing a multiplication.  This is because the AIR
// level does not support an explicit exponent operator.
func (p *AirLowering) lowerExpTo(e *Exp, airModule *air.ModuleBuilder) air.Term {
	// Lower the expression being raised
	le := p.lowerTermTo(e.Arg, airModule)
	// Multiply it out k times
	es := make([]air.Term, e.Pow)
	//
	for i := uint64(0); i < e.Pow; i++ {
		es[i] = le
	}
	// Done
	return ir.Product(es...)
}

func (p *AirLowering) lowerIfZeroTo(e *IfZero, airModule *air.ModuleBuilder) air.Term {
	var (
		trueCondition  = p.extractNormalisedCondition(true, e.Condition, airModule)
		falseCondition = p.extractNormalisedCondition(false, e.Condition, airModule)
		trueBranch     = p.lowerTermTo(e.TrueBranch, airModule)
		falseBranch    = p.lowerTermTo(e.FalseBranch, airModule)
	)
	//
	fb := ir.Product(trueCondition, falseBranch)
	tb := ir.Product(falseCondition, trueBranch)
	//
	return ir.Sum(tb, fb)
}

func (p *AirLowering) lowerNormTo(e *Norm, airModule *air.ModuleBuilder) air.Term {
	// Lower the expression being normalised
	arg := p.lowerTermTo(e.Arg, airModule)
	//
	return p.normalise(arg, airModule)
}

func (p *AirLowering) lowerVectorAccess(e *VectorAccess, airModule *air.ModuleBuilder) air.Term {
	var (
		terms []air.Term = make([]air.Term, len(e.Vars))
		shift            = uint(0)
	)
	//
	for i, v := range e.Vars {
		ith := ir.NewRegisterAccess[bls12_377.Element, air.Term](v.Register, v.Shift)
		// Apply shift
		terms[i] = ir.Product(shiftTerm(ith, shift))
		//
		shift = shift + airModule.Register(v.Register).Width
	}
	//
	return ir.Sum(terms...)
}

func shiftTerm(term air.Term, width uint) air.Term {
	if width == 0 {
		return term
	}
	// Compute 2^width
	n := field.TwoPowN[bls12_377.Element](width)
	//
	return ir.Product(ir.Const[bls12_377.Element, air.Term](n), term)
}

// Extract condition whilst ensuring it always evaluates to either 0 or 1.  This
// is useful for translating conditional terms.  For example, consider
// translating this:
//
// > 16 - (if (X == 0) 5 4)
//
// We translate this roughly as follows:
//
// > 16 - (X!=0)*5 - (X==0)*4
//
// Where we know that either X==0 or X!=0 will evaluate to 0.  However, if e.g.
// X==0 evaluates to 0 then we need X!=0 to evaluate to 1 (otherwise we've
// changed the meaning of our expression).
func (p *AirLowering) extractNormalisedCondition(sign bool, term LogicalTerm,
	airModule *air.ModuleBuilder) air.Term {
	//
	switch t := term.(type) {
	case *Conjunct:
		if sign {
			return p.extractNormalisedConjunction(sign, t.Args, airModule)
		}

		return p.extractNormalisedDisjunction(sign, t.Args, airModule)
	case *Disjunct:
		if sign {
			return p.extractNormalisedDisjunction(sign, t.Args, airModule)
		}

		return p.extractNormalisedConjunction(sign, t.Args, airModule)
	case *Equal:
		return p.extractNormalisedEquality(sign, t.Lhs, t.Rhs, airModule)
	case *Ite:
		panic("todo")
	case *Negate:
		return p.extractNormalisedCondition(!sign, t.Arg, airModule)
	case *NotEqual:
		return p.extractNormalisedEquality(!sign, t.Lhs, t.Rhs, airModule)
	default:
		name := reflect.TypeOf(t).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func (p *AirLowering) extractNormalisedConjunction(sign bool, terms []LogicalTerm,
	airModule *air.ModuleBuilder) air.Term {
	//
	args := p.extractNormalisedConditions(!sign, terms, airModule)
	// P && Q ==> !(!P || Q!) ==> 1 - ~(!P || !Q)
	return ir.Subtract(ir.Const64[bls12_377.Element, air.Term](1),
		p.normalise(ir.Product(args...), airModule))
}

func (p *AirLowering) extractNormalisedDisjunction(sign bool, terms []LogicalTerm,
	airModule *air.ModuleBuilder) air.Term {
	//
	ts := p.extractNormalisedConditions(sign, terms, airModule)
	// Easy case
	return ir.Product(ts...)
}

func (p *AirLowering) extractNormalisedEquality(sign bool, lhs Term, rhs Term,
	airModule *air.ModuleBuilder) air.Term {
	l := p.lowerTermTo(lhs, airModule)
	r := p.lowerTermTo(rhs, airModule)
	t := p.normalise(ir.Subtract(l, r), airModule)
	//
	if sign {
		return t
	}
	// Invert for not-equals
	return ir.Subtract(ir.Const64[bls12_377.Element, air.Term](1), t)
}

func (p *AirLowering) extractNormalisedConditions(sign bool, es []LogicalTerm,
	airModule *air.ModuleBuilder) []air.Term {
	//
	exprs := make([]air.Term, len(es))
	//
	for i, e := range es {
		exprs[i] = p.extractNormalisedCondition(sign, e, airModule)
	}
	//
	return exprs
}

func (p *AirLowering) normalise(arg air.Term, airModule *air.ModuleBuilder) air.Term {
	bounds := arg.ValueRange(airModule)
	// Check whether normalisation actually required.  For example, if the
	// argument is just a binary column then a normalisation is not actually
	// required.
	if p.config.InverseEliminiationLevel > 0 && bounds.Within(util_math.NewInterval64(0, 1)) {
		// arg ∈ {0,1} ==> normalised already :)
		return arg
	} else if p.config.InverseEliminiationLevel > 0 && bounds.Within(util_math.NewInterval64(-1, 1)) {
		// arg ∈ {-1,0,1} ==> (arg*arg) ∈ {0,1}
		return ir.Product(arg, arg)
	}
	// Determine appropriate shift
	shift := 0
	// Apply shift normalisation (if enabled)
	if p.config.ShiftNormalisation {
		// Determine shift ranges
		min, max := arg.ShiftRange()
		// determine shift amount
		if max < 0 {
			shift = max
		} else if min > 0 {
			shift = min
		}
	}
	// Construct an expression representing the normalised value of e.  That is,
	// an expression which is 0 when e is 0, and 1 when e is non-zero.
	arg = arg.ApplyShift(-shift).Simplify(false)
	norm := air_gadgets.Normalise(arg, airModule)
	//
	return norm.ApplyShift(shift)
}

// Simplify a bunch of logical terms
func simplify(terms []air.Term) []air.Term {
	var nterms []air.Term = make([]air.Term, len(terms))
	//
	for i, t := range terms {
		nterms[i] = t.Simplify(false)
	}
	//
	return nterms
}

// Construct the disjunction lhs v rhs, where both lhs and rhs can be
// conjunctions of terms.
func disjunction(terms ...[]air.Term) []air.Term {
	// Base cases
	switch len(terms) {
	case 0:
		return nil
	case 1:
		return terms[0]
	}
	//
	var (
		nterms []air.Term
		lhs    = terms[0]
		rhs    = disjunction(terms[1:]...)
	)
	// FIXME: this is where things can get expensive, and it would be useful to
	// explore whether extractNormalisedCondition could help here.
	for _, l := range lhs {
		for _, r := range rhs {
			disjunct := ir.Product(l, r)
			nterms = append(nterms, disjunct)
		}
	}
	//
	return nterms
}

func conjunction(terms ...[]air.Term) []air.Term {
	// FIXME: can we do better here in cases where the terms being conjuncted
	// can be safely summed?  This requires exploiting the ValueRange analysis
	// on the terms and check whether their sum fits within the field element.
	var nterms []air.Term
	// Combine conjuncts
	for _, ts := range terms {
		nterms = array.AppendAll(nterms, ts...)
	}
	//
	return nterms
}
