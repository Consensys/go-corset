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
	util_math "github.com/consensys/go-corset/pkg/util/math"
)

// LowerToAir lowers (or refines) an MIR schema into an AIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func LowerToAir[F field.Element[F]](schema Schema[F], config OptimisationConfig) air.Schema[F] {
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
type AirLowering[F field.Element[F]] struct {
	config OptimisationConfig
	// Modules we are lowering from
	mirSchema Schema[F]
	// Modules we are lowering to
	airSchema air.SchemaBuilder[F]
}

// NewAirLowering constructs an initial state for lowering a given MIR schema.
func NewAirLowering[F field.Element[F]](mirSchema Schema[F]) AirLowering[F] {
	var (
		airSchema = ir.NewSchemaBuilder[F, air.Constraint[F], air.Term[F], air.Module[F]]()
	)
	// Initialise AIR modules
	for _, m := range mirSchema.RawModules() {
		airSchema.NewModule(m.Name(), m.LengthMultiplier(), m.AllowPadding(), m.IsSynthetic())
	}
	//
	return AirLowering[F]{
		DEFAULT_OPTIMISATION_LEVEL,
		mirSchema,
		airSchema,
	}
}

// ConfigureOptimisation configures the amount of optimisation to apply during
// the lowering process.
func (p *AirLowering[F]) ConfigureOptimisation(config OptimisationConfig) {
	p.config = config
}

// Lower the MIR schema provide when this lowering instance was created into an
// equivalent AIR schema.
func (p *AirLowering[F]) Lower() air.Schema[F] {
	// Initialise modules
	for i := 0; i < int(p.mirSchema.Width()); i++ {
		p.InitialiseModule(uint(i))
	}
	// Lower modules
	for i := 0; i < int(p.mirSchema.Width()); i++ {
		p.LowerModule(uint(i))
	}
	// Build concrete modules from schema
	modules := ir.BuildSchema[air.Module[F]](p.airSchema)
	// Done
	return schema.NewUniformSchema(modules)
}

// InitialiseModule simply initialises all registers within the module, but does
// not lower any constraint or assignments.
func (p *AirLowering[F]) InitialiseModule(index uint) {
	var (
		mirModule = p.mirSchema.Module(index)
		airModule = p.airSchema.Module(index)
	)
	// Initialise registers in AIR module
	airModule.NewRegisters(mirModule.Registers()...)
}

// LowerModule lowers the given MIR module into the corresponding AIR module.
// This includes all constraints and assignments.
func (p *AirLowering[F]) LowerModule(index uint) {
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
		constraint := iter.Next().(Constraint[F])
		//
		p.lowerConstraintToAir(constraint, airModule)
	}
}

// Lower a constraint to the AIR level.
func (p *AirLowering[F]) lowerConstraintToAir(c Constraint[F], airModule *air.ModuleBuilder[F]) {
	// Check what kind of constraint we have
	switch v := c.constraint.(type) {
	case Assertion[F]:
		p.lowerAssertionToAir(v, airModule)
	case InterleavingConstraint[F]:
		p.lowerInterleavingConstraintToAir(v, airModule)
	case LookupConstraint[F]:
		p.lowerLookupConstraintToAir(v, airModule)
	case PermutationConstraint[F]:
		p.lowerPermutationConstraintToAir(v, airModule)
	case RangeConstraint[F]:
		p.lowerRangeConstraintToAir(v, airModule)
	case SortedConstraint[F]:
		p.lowerSortedConstraintToAir(v, airModule)
	case VanishingConstraint[F]:
		p.lowerVanishingConstraintToAir(v, airModule)
	default:
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

// Lowering an assertion is straightforward since its not a true constraint.
func (p *AirLowering[F]) lowerAssertionToAir(v Assertion[F], airModule *air.ModuleBuilder[F]) {
	airModule.AddConstraint(air.NewAssertion(v.Handle, v.Context, v.Domain, v.Property))
}

// Lower a vanishing constraint to the AIR level.  This is relatively
// straightforward and simply relies on lowering the expression being
// constrained.  This may result in the generation of computed columns, e.g. to
// hold inverses, etc.
func (p *AirLowering[F]) lowerVanishingConstraintToAir(v VanishingConstraint[F], airModule *air.ModuleBuilder[F]) {
	//
	var (
		terms = p.lowerAndSimplifyLogicalTo(v.Constraint, airModule)
	)
	//
	for i, air_expr := range terms {
		// Construct suitable handle to distinguish this case
		handle := fmt.Sprintf("%s#%d", v.Handle, i)
		// Add constraint
		airModule.AddConstraint(
			air.NewVanishingConstraint(handle, v.Context, v.Domain, air_expr))
	}
}

// Lower a permutation constraint to the AIR level.  This is trivial because
// permutation constraints do not currently support complex forms.
func (p *AirLowering[F]) lowerPermutationConstraintToAir(v PermutationConstraint[F], airModule *air.ModuleBuilder[F]) {
	airModule.AddConstraint(
		air.NewPermutationConstraint[F](v.Handle, v.Context, v.Targets, v.Sources),
	)
}

// Lower a range constraint to the AIR level.  The challenge here is that a
// range constraint at the AIR level cannot use arbitrary expressions; rather it
// can only constrain columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func (p *AirLowering[F]) lowerRangeConstraintToAir(v RangeConstraint[F], airModule *air.ModuleBuilder[F]) {
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
func (p *AirLowering[F]) lowerInterleavingConstraintToAir(c InterleavingConstraint[F],
	airModule *air.ModuleBuilder[F]) {
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
func (p *AirLowering[F]) lowerLookupConstraintToAir(c LookupConstraint[F], airModule *air.ModuleBuilder[F]) {
	var (
		sources = make([]lookup.Vector[F, *air.ColumnAccess[F]], len(c.Sources))
		targets = make([]lookup.Vector[F, *air.ColumnAccess[F]], len(c.Targets))
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

func (p *AirLowering[F]) expandLookupVectorToAir(vector lookup.Vector[F, Term[F]],
) lookup.Vector[F, *air.ColumnAccess[F]] {
	var (
		terms    = p.expandTerms(vector.Module, vector.Terms...)
		selector util.Option[*air.ColumnAccess[F]]
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
func (p *AirLowering[F]) lowerSortedConstraintToAir(c SortedConstraint[F], airModule *air.ModuleBuilder[F]) {
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
	gadget := air_gadgets.NewLexicographicSortingGadget[F](c.Handle, sources, c.BitWidth).
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

func (p *AirLowering[F]) expandTerms(context schema.ModuleId, terms ...Term[F]) []*air.ColumnAccess[F] {
	var nterms = make([]*air.ColumnAccess[F], len(terms))
	//
	for i, ith := range terms {
		nterms[i] = p.expandTerm(context, ith)
	}
	//
	return nterms
}

func (p *AirLowering[F]) expandTerm(context schema.ModuleId, term Term[F]) *air.ColumnAccess[F] {
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
	return ir.RawRegisterAccess[F, air.Term[F]](source_register, 0)
}

func (p *AirLowering[F]) lowerAndSimplifyLogicalTo(term LogicalTerm[F],
	airModule *air.ModuleBuilder[F]) []air.Term[F] {
	// Apply all reasonable simplifications
	term = term.Simplify(false)
	// Lower properly
	return simplify(p.lowerLogicalTo(true, term, airModule))
}

func (p *AirLowering[F]) lowerLogicalTo(sign bool, e LogicalTerm[F], airModule *air.ModuleBuilder[F]) []air.Term[F] {
	//
	switch e := e.(type) {
	case *Conjunct[F]:
		return p.lowerConjunctionTo(sign, e, airModule)
	case *Disjunct[F]:
		return p.lowerDisjunctionTo(sign, e, airModule)
	case *Equal[F]:
		return p.lowerEqualityTo(sign, e.Lhs, e.Rhs, airModule)
	case *Ite[F]:
		return p.lowerIteTo(sign, e, airModule)
	case *Negate[F]:
		return p.lowerLogicalTo(!sign, e.Arg, airModule)
	case *NotEqual[F]:
		return p.lowerEqualityTo(!sign, e.Lhs, e.Rhs, airModule)
	case *Inequality[F]:
		panic("inequalities cannot (currently) be lowered to AIR")
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func (p *AirLowering[F]) lowerLogicalsTo(sign bool, airModule *air.ModuleBuilder[F], terms ...LogicalTerm[F],
) [][]air.Term[F] {
	//
	nexprs := make([][]air.Term[F], len(terms))

	for i := range len(terms) {
		nexprs[i] = p.lowerLogicalTo(sign, terms[i], airModule)
	}

	return nexprs
}

func (p *AirLowering[F]) lowerConjunctionTo(sign bool, e *Conjunct[F], airModule *air.ModuleBuilder[F]) []air.Term[F] {
	var terms = p.lowerLogicalsTo(sign, airModule, e.Args...)
	//
	if sign {
		return conjunction(terms...)
	}
	//
	return disjunction(terms...)
}

func (p *AirLowering[F]) lowerDisjunctionTo(sign bool, e *Disjunct[F], airModule *air.ModuleBuilder[F]) []air.Term[F] {
	var terms = p.lowerLogicalsTo(sign, airModule, e.Args...)
	//
	if sign {
		//
		return disjunction(terms...)
	}
	//
	return conjunction(terms...)
}

func (p *AirLowering[F]) lowerEqualityTo(sign bool, left Term[F], right Term[F], airModule *air.ModuleBuilder[F],
) []air.Term[F] {
	//
	var (
		lhs air.Term[F] = p.lowerTermTo(left, airModule)
		rhs air.Term[F] = p.lowerTermTo(right, airModule)
		eq              = ir.Subtract(lhs, rhs)
	)
	//
	if sign {
		return []air.Term[F]{eq}
	}
	//
	one := ir.Const64[F, air.Term[F]](1)
	// construct norm(eq)
	norm_eq := p.normalise(eq, airModule)
	// construct 1 - norm(eq)
	return []air.Term[F]{ir.Subtract(one, norm_eq)}
}

func (p *AirLowering[F]) lowerIteTo(sign bool, e *Ite[F], airModule *air.ModuleBuilder[F]) []air.Term[F] {
	if sign {
		return p.lowerPositiveIteTo(e, airModule)
	}
	//
	return p.lowerNegativeIteTo(e, airModule)
}

func (p *AirLowering[F]) lowerPositiveIteTo(e *Ite[F], airModule *air.ModuleBuilder[F]) []air.Term[F] {
	var (
		terms []air.Term[F]
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
			return []air.Term[F]{ir.Sum(tb, fb)}
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
func (p *AirLowering[F]) lowerNegativeIteTo(e *Ite[F], airModule *air.ModuleBuilder[F]) []air.Term[F] {
	// NOTE: using extractNormalisedCondition could be useful here.
	var (
		terms [][]air.Term[F]
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
func (p *AirLowering[F]) lowerAndSimplifyTermTo(term Term[F], airModule *air.ModuleBuilder[F]) air.Term[F] {
	// Apply all reasonable simplifications
	term = term.Simplify(false)
	// Lower properly
	return p.lowerTermTo(term, airModule)
}

// Inner form is used for recursive calls and does not repeat the constant
// propagation phase.
func (p *AirLowering[F]) lowerTermTo(e Term[F], airModule *air.ModuleBuilder[F]) air.Term[F] {
	//
	switch e := e.(type) {
	case *Add[F]:
		args := p.lowerTerms(e.Args, airModule)
		return ir.Sum(args...)
	case *Cast[F]:
		return p.lowerTermTo(e.Arg, airModule)
	case *Constant[F]:
		return ir.Const[F, air.Term[F]](e.Value)
	case *RegisterAccess[F]:
		return ir.NewRegisterAccess[F, air.Term[F]](e.Register, e.Shift)
	case *Exp[F]:
		return p.lowerExpTo(e, airModule)
	case *IfZero[F]:
		return p.lowerIfZeroTo(e, airModule)
	case *LabelledConst[F]:
		return ir.Const[F, air.Term[F]](e.Value)
	case *Mul[F]:
		args := p.lowerTerms(e.Args, airModule)
		return ir.Product(args...)
	case *Norm[F]:
		return p.lowerNormTo(e, airModule)
	case *Sub[F]:
		args := p.lowerTerms(e.Args, airModule)
		return ir.Subtract(args...)
	case *VectorAccess[F]:
		return p.lowerVectorAccess(e, airModule)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

// Lower a set of zero or more MIR expressions.
func (p *AirLowering[F]) lowerTerms(exprs []Term[F], airModule *air.ModuleBuilder[F]) []air.Term[F] {
	nexprs := make([]air.Term[F], len(exprs))

	for i := range len(exprs) {
		nexprs[i] = p.lowerTermTo(exprs[i], airModule)
	}

	return nexprs
}

// LowerTo lowers an exponent expression to the AIR level by lowering the
// argument, and then constructing a multiplication.  This is because the AIR
// level does not support an explicit exponent operator.
func (p *AirLowering[F]) lowerExpTo(e *Exp[F], airModule *air.ModuleBuilder[F]) air.Term[F] {
	// Lower the expression being raised
	le := p.lowerTermTo(e.Arg, airModule)
	// Multiply it out k times
	es := make([]air.Term[F], e.Pow)
	//
	for i := uint64(0); i < e.Pow; i++ {
		es[i] = le
	}
	// Done
	return ir.Product(es...)
}

func (p *AirLowering[F]) lowerIfZeroTo(e *IfZero[F], airModule *air.ModuleBuilder[F]) air.Term[F] {
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

func (p *AirLowering[F]) lowerNormTo(e *Norm[F], airModule *air.ModuleBuilder[F]) air.Term[F] {
	// Lower the expression being normalised
	arg := p.lowerTermTo(e.Arg, airModule)
	//
	return p.normalise(arg, airModule)
}

func (p *AirLowering[F]) lowerVectorAccess(e *VectorAccess[F], airModule *air.ModuleBuilder[F]) air.Term[F] {
	var (
		terms []air.Term[F] = make([]air.Term[F], len(e.Vars))
		shift               = uint(0)
	)
	//
	for i, v := range e.Vars {
		ith := ir.NewRegisterAccess[F, air.Term[F]](v.Register, v.Shift)
		// Apply shift
		terms[i] = ir.Product(shiftTerm(ith, shift))
		//
		shift = shift + airModule.Register(v.Register).Width
	}
	//
	return ir.Sum(terms...)
}

func shiftTerm[F field.Element[F]](term air.Term[F], width uint) air.Term[F] {
	if width == 0 {
		return term
	}
	// Compute 2^width
	n := field.TwoPowN[F](width)
	//
	return ir.Product(ir.Const[F, air.Term[F]](n), term)
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
func (p *AirLowering[F]) extractNormalisedCondition(sign bool, term LogicalTerm[F],
	airModule *air.ModuleBuilder[F]) air.Term[F] {
	//
	switch t := term.(type) {
	case *Conjunct[F]:
		if sign {
			return p.extractNormalisedConjunction(sign, t.Args, airModule)
		}

		return p.extractNormalisedDisjunction(sign, t.Args, airModule)
	case *Disjunct[F]:
		if sign {
			return p.extractNormalisedDisjunction(sign, t.Args, airModule)
		}

		return p.extractNormalisedConjunction(sign, t.Args, airModule)
	case *Equal[F]:
		return p.extractNormalisedEquality(sign, t.Lhs, t.Rhs, airModule)
	case *Ite[F]:
		panic("todo")
	case *Negate[F]:
		return p.extractNormalisedCondition(!sign, t.Arg, airModule)
	case *NotEqual[F]:
		return p.extractNormalisedEquality(!sign, t.Lhs, t.Rhs, airModule)
	default:
		name := reflect.TypeOf(t).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func (p *AirLowering[F]) extractNormalisedConjunction(sign bool, terms []LogicalTerm[F],
	airModule *air.ModuleBuilder[F]) air.Term[F] {
	//
	args := p.extractNormalisedConditions(!sign, terms, airModule)
	// P && Q ==> !(!P || Q!) ==> 1 - ~(!P || !Q)
	return ir.Subtract(ir.Const64[F, air.Term[F]](1),
		p.normalise(ir.Product(args...), airModule))
}

func (p *AirLowering[F]) extractNormalisedDisjunction(sign bool, terms []LogicalTerm[F],
	airModule *air.ModuleBuilder[F]) air.Term[F] {
	//
	ts := p.extractNormalisedConditions(sign, terms, airModule)
	// Easy case
	return ir.Product(ts...)
}

func (p *AirLowering[F]) extractNormalisedEquality(sign bool, lhs Term[F], rhs Term[F],
	airModule *air.ModuleBuilder[F]) air.Term[F] {
	l := p.lowerTermTo(lhs, airModule)
	r := p.lowerTermTo(rhs, airModule)
	t := p.normalise(ir.Subtract(l, r), airModule)
	//
	if sign {
		return t
	}
	// Invert for not-equals
	return ir.Subtract(ir.Const64[F, air.Term[F]](1), t)
}

func (p *AirLowering[F]) extractNormalisedConditions(sign bool, es []LogicalTerm[F],
	airModule *air.ModuleBuilder[F]) []air.Term[F] {
	//
	exprs := make([]air.Term[F], len(es))
	//
	for i, e := range es {
		exprs[i] = p.extractNormalisedCondition(sign, e, airModule)
	}
	//
	return exprs
}

func (p *AirLowering[F]) normalise(arg air.Term[F], airModule *air.ModuleBuilder[F]) air.Term[F] {
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
func simplify[F field.Element[F]](terms []air.Term[F]) []air.Term[F] {
	var nterms []air.Term[F] = make([]air.Term[F], len(terms))
	//
	for i, t := range terms {
		nterms[i] = t.Simplify(false)
	}
	//
	return nterms
}

// Construct the disjunction lhs v rhs, where both lhs and rhs can be
// conjunctions of terms.
func disjunction[F field.Element[F]](terms ...[]air.Term[F]) []air.Term[F] {
	// Base cases
	switch len(terms) {
	case 0:
		// NOTE: return non-zero value to indicate a failure.
		return []air.Term[F]{ir.Const64[F, air.Term[F]](1)}
	case 1:
		return terms[0]
	}
	//
	var (
		nterms []air.Term[F]
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

func conjunction[F field.Element[F]](terms ...[]air.Term[F]) []air.Term[F] {
	// FIXME: can we do better here in cases where the terms being conjuncted
	// can be safely summed?  This requires exploiting the ValueRange analysis
	// on the terms and check whether their sum fits within the field element.
	var nterms []air.Term[F]
	// Combine conjuncts
	for _, ts := range terms {
		nterms = array.AppendAll(nterms, ts...)
	}
	//
	return nterms
}
