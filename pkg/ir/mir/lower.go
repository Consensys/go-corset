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
	"github.com/consensys/go-corset/pkg/ir/air"
	air_gadgets "github.com/consensys/go-corset/pkg/ir/air/gadgets"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
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
		airSchema = ir.NewSchemaBuilder[air.Constraint, air.Term]()
	)
	// Initialise AIR modules
	for _, m := range mirSchema.RawModules() {
		airSchema.NewModule(m.Name())
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
		p.LowerModule(uint(i))
	}
	// Done
	return schema.NewUniformSchema(p.airSchema.Build())
}

// LowerModule lowers the given MIR module into the correspondind AIR module.
// This includes all registers, constraints and assignments.
func (p *AirLowering) LowerModule(index uint) {
	var (
		mirModule = p.mirSchema.Module(index)
		airModule = p.airSchema.Module(index)
	)
	// Initialise registers in AIR module
	airModule.NewRegisters(mirModule.Registers()...)
	// Lower constraints
	for iter := mirModule.Constraints(); iter.HasNext(); {
		// Following should always hold
		constraint := iter.Next().(Constraint)
		//
		p.lowerConstraintToAir(constraint, airModule)
	}
	// Lower assignments
	return
}

// // Lower an assignment to the AIR level.
// func lowerAssignmentToAir(c sc.Assignment, mirSchema *Schema, airSchema *air.Schema) {
// 	if v, ok := c.(Permutation); ok {
// 		lowerPermutationToAir(v, mirSchema, airSchema)
// 	} else if _, ok := c.(Interleaving); ok {
// 		// Nothing to do for interleaving constraints, as they can be passed
// 		// directly down to the AIR level
// 		return
// 	} else if _, ok := c.(Computation); ok {
// 		// Nothing to do for computation, as they can be passed directly down to
// 		// the AIR level
// 		return
// 	} else {
// 		panic("unknown assignment")
// 	}
// }

// Lower a constraint to the AIR level.
func (p *AirLowering) lowerConstraintToAir(c Constraint, airModule *air.ModuleBuilder) {
	// Check what kind of constraint we have
	switch v := c.constraint.(type) {
	case Assertion:
		p.lowerAssertionToAir(v, airModule)
	case LookupConstraint:
		p.lowerLookupConstraintToAir(v, airModule)
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
		terms = p.lowerLogicalTo(true, v.Constraint, v.Context, airModule)
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

// Lower a range constraint to the AIR level.  The challenge here is that a
// range constraint at the AIR level cannot use arbitrary expressions; rather it
// can only constrain columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func (p *AirLowering) lowerRangeConstraintToAir(v RangeConstraint, airModule *air.ModuleBuilder) {
	mirModule := p.mirSchema.Module(v.Context.ModuleId)
	bitwidth := v.Expr.ValueRange(mirModule).BitWidth()
	// Lower target expression
	target := p.lowerTermTo(v.Context, v.Expr, airModule)
	// Expand target expression (if necessary)
	register := air_gadgets.Expand(v.Context, bitwidth, target, airModule)
	// Yes, a constraint is implied.  Now, decide whether to use a range
	// constraint or just a vanishing constraint.
	if v.Bitwidth == 1 {
		// u1 => use vanishing constraint X * (X - 1)
		air_gadgets.ApplyBinaryGadget(register, v.Context, airModule)
	} else if v.Bitwidth <= p.config.MaxRangeConstraint {
		// u2..n use range constraints
		column := ir.RawRegisterAccess[air.Term](register, 0)
		//
		airModule.AddConstraint(air.NewRangeConstraint("", v.Context, *column, v.Bitwidth))
	} else {
		// Apply bitwidth gadget
		air_gadgets.ApplyBitwidthGadget(register, v.Bitwidth, ir.Const64[air.Term](1), airModule)
	}
}

// Lower a lookup constraint to the AIR level.  The challenge here is that a
// lookup constraint at the AIR level cannot use arbitrary expressions; rather,
// it can only access columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func (p *AirLowering) lowerLookupConstraintToAir(c LookupConstraint, airModule *air.ModuleBuilder) {
	targets := make([]*air.ColumnAccess, len(c.Targets))
	sources := make([]*air.ColumnAccess, len(c.Sources))
	//
	for i := 0; i < len(targets); i++ {
		targetBitwidth := c.Targets[i].ValueRange(airModule).BitWidth()
		sourceBitwidth := c.Sources[i].ValueRange(airModule).BitWidth()
		// Lower source and target expressions
		target := p.lowerTermTo(c.TargetContext, c.Targets[i], airModule)
		source := p.lowerTermTo(c.SourceContext, c.Sources[i], airModule)
		// Expand them
		target_register := air_gadgets.Expand(c.TargetContext, targetBitwidth, target, airModule)
		source_register := air_gadgets.Expand(c.SourceContext, sourceBitwidth, source, airModule)
		//
		targets[i] = ir.RawRegisterAccess[air.Term](target_register, 0)
		sources[i] = ir.RawRegisterAccess[air.Term](source_register, 0)
	}
	// finally add the constraint
	airModule.AddConstraint(air.NewLookupConstraint(c.Handle, c.SourceContext, c.TargetContext, sources, targets))
}

// Lower a sorted constraint to the AIR level.  The challenge here is that there
// is not concept of sorting constraints at the AIR level.  Instead, we have to
// generate the necessary machinery to enforce the sorting constraint.
func (p *AirLowering) lowerSortedConstraintToAir(c SortedConstraint, airModule *air.ModuleBuilder) {
	// sources := make([]uint, len(c.Sources))
	// //
	// for i := 0; i < len(sources); i++ {
	// 	sourceBitwidth := rangeOfTerm(c.Sources[i].term, mirSchema).BitWidth()
	// 	// Lower source expression
	// 	source := lowerExprTo(c.Context, c.Sources[i], mirSchema, airSchema, cfg)
	// 	// Expand them
	// 	sources[i] = air_gadgets.Expand(c.Context, sourceBitwidth, source, airSchema)
	// }
	// // Determine number of ordered columns
	// numSignedCols := len(c.Signs)
	// // finally add the constraint
	// if numSignedCols == 1 {
	// 	// For a single column sort, its actually a bit easier because we don't
	// 	// need to implement a multiplexor (i.e. to determine which column is
	// 	// differs, etc).  Instead, we just need a delta column which ensures
	// 	// there is a non-negative difference between consecutive rows.  This
	// 	// also requires bitwidth constraints.
	// 	gadget := air_gadgets.NewColumnSortGadget(c.Handle, sources[0], c.BitWidth)
	// 	gadget.SetSign(c.Signs[0])
	// 	gadget.SetStrict(c.Strict)
	// 	// Add (optional) selector
	// 	if c.Selector.HasValue() {
	// 		selector := lowerExprTo(c.Context, c.Selector.Unwrap(), mirSchema, airSchema, cfg)
	// 		gadget.SetSelector(selector)
	// 	}
	// 	// Done!
	// 	gadget.Apply(airSchema)
	// } else {
	// 	// For a multi column sort, its a bit harder as we need additional
	// 	// logic to ensure the target columns are lexicographally sorted.
	// 	gadget := air_gadgets.NewLexicographicSortingGadget(c.Handle, sources, c.BitWidth)
	// 	gadget.SetSigns(c.Signs...)
	// 	gadget.SetStrict(c.Strict)
	// 	// Add (optional) selector
	// 	if c.Selector.HasValue() {
	// 		selector := lowerExprTo(c.Context, c.Selector.Unwrap(), mirSchema, airSchema, cfg)
	// 		gadget.SetSelector(selector)
	// 	}
	// 	// Done
	// 	gadget.Apply(airSchema)
	// }
	// // Sanity check bitwidth
	// bitwidth := uint(0)

	// for i := 0; i < numSignedCols; i++ {
	// 	// Extract bitwidth of ith column
	// 	ith := mirSchema.Columns().Nth(sources[i]).DataType.AsUint().BitWidth()
	// 	if ith > bitwidth {
	// 		bitwidth = ith
	// 	}
	// }
	// //
	// if bitwidth != c.BitWidth {
	// 	// Should be unreachable.
	// 	msg := fmt.Sprintf("incompatible bitwidths (%d vs %d)", bitwidth, c.BitWidth)
	// 	panic(msg)
	// }
	panic("todo")
}

// // Lower a permutation to the AIR level.  This has quite a few
// // effects.  Firstly, permutation constraints are added for all of the
// // new columns.  Secondly, sorting constraints (and their associated
// // computed columns) must also be added.  Finally, a trace
// // computation is required to ensure traces are correctly expanded to
// // meet the requirements of a sorted permutation.
// func lowerPermutationToAir(c Permutation, mirSchema *Schema, airSchema *air.Schema) {
// 	builder := strings.Builder{}
// 	c_targets := c.Targets
// 	targets := make([]uint, len(c_targets))
// 	//
// 	builder.WriteString("permutation")
// 	// Add individual permutation constraints
// 	for i := 0; i < len(c_targets); i++ {
// 		var ok bool
// 		// TODO: how best to avoid this lookup?
// 		targets[i], ok = sc.ColumnIndexOf(airSchema, c.Module(), c_targets[i].Name)
// 		//
// 		if !ok {
// 			panic("internal failure")
// 		}
// 		//
// 		builder.WriteString(fmt.Sprintf(":%s", c_targets[i].Name))
// 	}
// 	//
// 	airSchema.AddPermutationConstraint(builder.String(), c.Context(), targets, c.Sources)
// 	// Determine number of ordered columns
// 	numSignedCols := len(c.Signs)
// 	// Add sorting constraints + computed columns as necessary.
// 	if numSignedCols == 1 {
// 		// For a single column sort, its actually a bit easier because we don't
// 		// need to implement a multiplexor (i.e. to determine which column is
// 		// differs, etc).  Instead, we just need a delta column which ensures
// 		// there is a non-negative difference between consecutive rows.  This
// 		// also requires bitwidth constraints.
// 		bitwidth := mirSchema.Columns().Nth(c.Sources[0]).DataType.AsUint().BitWidth()
// 		// Identify target column name
// 		target := mirSchema.Columns().Nth(targets[0]).Name
// 		// Add column sorting constraints
// 		gadget := air_gadgets.NewColumnSortGadget(target, targets[0], bitwidth)
// 		gadget.SetSign(c.Signs[0])
// 		// Done!
// 		gadget.Apply(airSchema)
// 	} else {
// 		// For a multi column sort, its a bit harder as we need additional
// 		// logic to ensure the target columns are lexicographally sorted.
// 		bitwidth := uint(0)

// 		for i := 0; i < numSignedCols; i++ {
// 			// Extract bitwidth of ith column
// 			ith := mirSchema.Columns().Nth(c.Sources[i]).DataType.AsUint().BitWidth()
// 			if ith > bitwidth {
// 				bitwidth = ith
// 			}
// 		}
// 		// Construct a unique prefix for this sort.
// 		prefix := constructLexicographicSortingPrefix(targets, c.Signs, airSchema)
// 		// Add lexicographically sorted constraints
// 		// For a multi column sort, its a bit harder as we need additional
// 		// logic to ensure the target columns are lexicographally sorted.
// 		gadget := air_gadgets.NewLexicographicSortingGadget(prefix, targets, bitwidth)
// 		gadget.SetSigns(c.Signs...)
// 		// Done
// 		gadget.Apply(airSchema)
// 	}
// }

func (p *AirLowering) lowerLogicalTo(sign bool, e LogicalTerm, ctx trace.Context,
	airModule *air.ModuleBuilder) []air.Term {
	//
	switch e := e.(type) {
	case *Conjunct:
		return p.lowerConjunctionTo(sign, e, ctx, airModule)
	case *Disjunct:
		return p.lowerDisjunctionTo(sign, e, ctx, airModule)
	case *Equal:
		return p.lowerEqualityTo(sign, e.Lhs, e.Rhs, ctx, airModule)
	case *Ite:
		return p.lowerIteTo(sign, e, ctx, airModule)
	case *Negate:
		return p.lowerLogicalTo(!sign, e.Arg, ctx, airModule)
	case *NotEqual:
		return p.lowerEqualityTo(!sign, e.Lhs, e.Rhs, ctx, airModule)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func (p *AirLowering) lowerLogicalsTo(sign bool, ctx trace.Context,
	airModule *air.ModuleBuilder, terms ...LogicalTerm) [][]air.Term {
	nexprs := make([][]air.Term, len(terms))

	for i := range len(terms) {
		nexprs[i] = p.lowerLogicalTo(sign, terms[i], ctx, airModule)
	}

	return nexprs
}

func (p *AirLowering) lowerConjunctionTo(sign bool, e *Conjunct, ctx trace.Context,
	airModule *air.ModuleBuilder) []air.Term {
	var terms = p.lowerLogicalsTo(sign, ctx, airModule, e.Args...)
	//
	if sign {
		return conjunction(terms...)
	}
	//
	return disjunction(terms...)
}

func (p *AirLowering) lowerDisjunctionTo(sign bool, e *Disjunct, ctx trace.Context,
	airModule *air.ModuleBuilder) []air.Term {
	var terms = p.lowerLogicalsTo(sign, ctx, airModule, e.Args...)
	//
	if sign {
		return disjunction(terms...)
	}
	//
	return conjunction(terms...)
}

func (p *AirLowering) lowerEqualityTo(sign bool, left Term, right Term, ctx trace.Context,
	airModule *air.ModuleBuilder) []air.Term {
	//
	var (
		lhs air.Term = p.lowerTermTo(ctx, left, airModule)
		rhs air.Term = p.lowerTermTo(ctx, right, airModule)
		eq           = ir.Subtract(lhs, rhs)
	)
	//
	if sign {
		return []air.Term{eq}
	}
	//
	one := ir.Const64[air.Term](1)
	// construct norm(eq)
	norm_eq := p.lowerNormToInner(eq, ctx, airModule)
	// construct 1 - norm(eq)
	return []air.Term{ir.Subtract(one, norm_eq)}
}

func (p *AirLowering) lowerIteTo(sign bool, e *Ite, ctx trace.Context, airModule *air.ModuleBuilder) []air.Term {
	if sign {
		return p.lowerPositiveIteTo(e, ctx, airModule)
	}
	//
	return p.lowerNegativeIteTo(e, ctx, airModule)
}

func (p *AirLowering) lowerPositiveIteTo(e *Ite, ctx trace.Context, airModule *air.ModuleBuilder) []air.Term {
	var (
		terms          []air.Term
		trueCondition  = p.lowerLogicalTo(true, e.Condition, ctx, airModule)
		falseCondition = p.lowerLogicalTo(false, e.Condition, ctx, airModule)
	)

	//
	if e.TrueBranch != nil {
		trueBranch := p.lowerLogicalTo(true, e.TrueBranch, ctx, airModule)
		terms = append(terms, disjunction(falseCondition, trueBranch)...)
	}
	//
	if e.FalseBranch != nil {
		falseBranch := p.lowerLogicalTo(true, e.FalseBranch, ctx, airModule)
		terms = append(terms, disjunction(trueCondition, falseBranch)...)
	}
	//
	return terms
}

func (p *AirLowering) lowerNegativeIteTo(e *Ite, ctx trace.Context, airModule *air.ModuleBuilder) []air.Term {
	panic("todo")
}

// // Lower an expression into the Arithmetic Intermediate Representation.
// // Essentially, this means eliminating normalising expressions by introducing
// // new columns into the given table (with appropriate constraints).  This first
// // performs constant propagation to ensure lowering is as efficient as possible.
// // A module identifier is required to determine where any computed columns
// // should be located.
// func lowerExprTo(ctx trace.Context, e1 Expr, mirSchema *Schema, airSchema *air.Schema,
// 	cfg OptimisationConfig) air.Expr {
// 	return lowerTermTo(ctx, e1.term, mirSchema, airSchema, cfg)
// }

func (p *AirLowering) lowerTermTo(ctx trace.Context, term Term, airModule *air.ModuleBuilder) air.Term {
	// Optimise normalisations
	// term = eliminateNormalisationInTerm(term, mirSchema, cfg)
	// Apply constant propagation
	//term = constantPropagationForTerm(term, false, airSchema)
	// Lower properly
	return p.lowerTermToInner(ctx, term, airModule)
}

// Inner form is used for recursive calls and does not repeat the constant
// propagation phase.
func (p *AirLowering) lowerTermToInner(ctx trace.Context, e Term, airModule *air.ModuleBuilder) air.Term {
	//
	switch e := e.(type) {
	case *Add:
		args := p.lowerTerms(ctx, e.Args, airModule)
		return ir.Sum(args...)
	case *Cast:
		// 	return lowerTermToInner(ctx, e.Arg, airModule)
		panic("got here")
	case *Constant:
		return ir.Const[air.Term](e.Value)
	case *RegisterAccess:
		return ir.NewRegisterAccess[air.Term](e.Register, e.Shift)
	case *Exp:
		return p.lowerExpTo(ctx, e, airModule)
	case *IfZero:
		return p.lowerIfZeroTo(ctx, e, airModule)
	case *LabelledConst:
		return ir.Const[air.Term](e.Value)
	case *Mul:
		args := p.lowerTerms(ctx, e.Args, airModule)
		return ir.Product(args...)
	case *Norm:
		return p.lowerNormTo(ctx, e, airModule)
	case *Sub:
		args := p.lowerTerms(ctx, e.Args, airModule)
		return ir.Subtract(args...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

// Lower a set of zero or more MIR expressions.
func (p *AirLowering) lowerTerms(ctx trace.Context, exprs []Term, airModule *air.ModuleBuilder) []air.Term {
	nexprs := make([]air.Term, len(exprs))

	for i := range len(exprs) {
		nexprs[i] = p.lowerTermToInner(ctx, exprs[i], airModule)
	}

	return nexprs
}

// LowerTo lowers an exponent expression to the AIR level by lowering the
// argument, and then constructing a multiplication.  This is because the AIR
// level does not support an explicit exponent operator.
func (p *AirLowering) lowerExpTo(ctx trace.Context, e *Exp, airModule *air.ModuleBuilder) air.Term {
	// Lower the expression being raised
	le := p.lowerTermToInner(ctx, e.Arg, airModule)
	// Multiply it out k times
	es := make([]air.Term, e.Pow)
	//
	for i := uint64(0); i < e.Pow; i++ {
		es[i] = le
	}
	// Done
	return ir.Product(es...)
}

func (p *AirLowering) lowerIfZeroTo(ctx trace.Context, e *IfZero, airModule *air.ModuleBuilder) air.Term {
	// var (
	// 	condition   = p.lowerLogicalTo(ctx, e.Condition, airModule)
	// 	trueBranch  = p.lowerTermToInner(ctx, e.TrueBranch, airModule)
	// 	falseBranch = p.lowerTermToInner(ctx, e.FalseBranch, airModule)
	// )
	// fb := ir.Product(condition, falseBranch)
	panic("todo")
}

func (p *AirLowering) lowerNormTo(ctx trace.Context, e *Norm, airModule *air.ModuleBuilder) air.Term {
	// Lower the expression being normalised
	arg := p.lowerTermToInner(ctx, e.Arg, airModule)
	//
	return p.lowerNormToInner(arg, ctx, airModule)
}

func (p *AirLowering) lowerNormToInner(arg air.Term, ctx trace.Context, airModule *air.ModuleBuilder) air.Term {
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
	norm := air_gadgets.Normalise(arg.ApplyShift(-shift), ctx, airModule)
	//
	return norm.ApplyShift(shift)
}

// // Construct a unique identifier for the given sort.  This should not conflict
// // with the identifier for any other sort.
// func constructLexicographicSortingPrefix(columns []uint, signs []bool, schema *air.Schema) string {
// 	// Use string builder to try and make this vaguely efficient.
// 	var id strings.Builder
// 	// Concatenate column names with their signs.
// 	for i := 0; i < len(columns); i++ {
// 		ith := schema.Columns().Nth(columns[i])
// 		id.WriteString(ith.Name)

// 		if i >= len(signs) {

// 		} else if signs[i] {
// 			id.WriteString("+")
// 		} else {
// 			id.WriteString("-")
// 		}
// 	}
// 	// Done
// 	return id.String()
// }

// Construct the disjunction lhs v rhs, where both lhs and rhs can be
// conjunctions of terms.
func disjunction(terms ...[]air.Term) []air.Term {
	if len(terms) == 1 {
		return terms[0]
	}
	//
	var (
		nterms []air.Term
		lhs    = terms[0]
		rhs    = disjunction(terms[1:]...)
	)
	// FIXME: this is where things can get expensive!
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
	var nterms []air.Term
	// Combine conjuncts
	for _, ts := range terms {
		nterms = util.AppendAll(nterms, ts...)
	}
	//
	return nterms
}
