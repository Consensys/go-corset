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
	"strings"

	"github.com/consensys/go-corset/pkg/air"
	air_gadgets "github.com/consensys/go-corset/pkg/air/gadgets"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// LowerToAir lowers (or refines) an MIR table into an AIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToAir(cfg OptimisationConfig) *air.Schema {
	airSchema := air.EmptySchema[Expr]()
	// Copy modules
	for _, mod := range p.modules {
		airSchema.AddModule(mod.Name)
	}
	// Add data columns.
	for _, c := range p.inputs {
		col := c.(DataColumn)
		airSchema.AddColumn(col.Context(), col.Name(), col.Type())
	}
	// Add Assignments. Again this has to be done first for things to work.
	// Essentially to reflect the fact that these columns have been added above
	// before others.  Realistically, the overall design of this process is a
	// bit broken right now.
	for _, assign := range p.assignments {
		airSchema.AddAssignment(assign)
	}
	// Now, lower assignments.
	for _, assign := range p.assignments {
		lowerAssignmentToAir(assign, p, airSchema, cfg)
	}
	// Lower vanishing constraints
	for _, c := range p.constraints {
		lowerConstraintToAir(c, p, airSchema, cfg)
	}
	// Add assertions (these do not need to be lowered)
	for _, assertion := range p.assertions {
		airSchema.AddPropertyAssertion(assertion.Handle, assertion.Context, assertion.Property)
	}
	// Done
	return airSchema
}

// Lower an assignment to the AIR level.
func lowerAssignmentToAir(c sc.Assignment, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	if v, ok := c.(Permutation); ok {
		lowerPermutationToAir(v, mirSchema, airSchema, cfg)
	} else if _, ok := c.(Interleaving); ok {
		// Nothing to do for interleaving constraints, as they can be passed
		// directly down to the AIR level
		return
	} else if _, ok := c.(Computation); ok {
		// Nothing to do for computation, as they can be passed directly down to
		// the AIR level
		return
	} else {
		panic("unknown assignment")
	}
}

// Lower a constraint to the AIR level.
func lowerConstraintToAir(c sc.Constraint, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	// Check what kind of constraint we have
	if v, ok := c.(LookupConstraint); ok {
		lowerLookupConstraintToAir(v, mirSchema, airSchema, cfg)
	} else if v, ok := c.(VanishingConstraint); ok {
		lowerVanishingConstraintToAir(v, mirSchema, airSchema, cfg)
	} else if v, ok := c.(RangeConstraint); ok {
		lowerRangeConstraintToAir(v, mirSchema, airSchema, cfg)
	} else if v, ok := c.(SortedConstraint); ok {
		lowerSortedConstraintToAir(v, mirSchema, airSchema, cfg)
	} else {
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

// Lower a vanishing constraint to the AIR level.  This is relatively
// straightforward and simply relies on lowering the expression being
// constrained.  This may result in the generation of computed columns, e.g. to
// hold inverses, etc.
func lowerVanishingConstraintToAir(v VanishingConstraint, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) {
	//
	air_expr := lowerConstraintTo(v.Context, v.Constraint, mirSchema, airSchema, cfg)
	// Check whether this is a constant
	constant := air_expr.AsConstant()
	// Check for compile-time constants
	if constant != nil && !constant.IsZero() {
		panic(fmt.Sprintf("constraint %s cannot vanish!", v.Handle))
	} else if constant == nil {
		airSchema.AddVanishingConstraint(v.Handle, v.Case, v.Context, v.Domain, air_expr)
	}
}

// Lower a range constraint to the AIR level.  The challenge here is that a
// range constraint at the AIR level cannot use arbitrary expressions; rather it
// can only constrain columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func lowerRangeConstraintToAir(v RangeConstraint, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	bitwidth := rangeOfTerm(v.Expr.term, mirSchema).BitWidth()
	// Lower target expression
	target := lowerExprTo(v.Context, v.Expr, mirSchema, airSchema, cfg)
	// Expand target expression (if necessary)
	column := air_gadgets.Expand(v.Context, bitwidth, target, airSchema)
	// Yes, a constraint is implied.  Now, decide whether to use a range
	// constraint or just a vanishing constraint.
	// Constrict gadget
	gadget := air_gadgets.NewBitwidthGadget(airSchema).
		WithLegacyTypeProofs(cfg.LegacyTypeProofs).
		WithMaxRangeConstraint(cfg.MaxRangeConstraint)
	//
	gadget.Constrain(column, v.Bitwidth)
}

// Lower a lookup constraint to the AIR level.  The challenge here is that a
// lookup constraint at the AIR level cannot use arbitrary expressions; rather,
// it can only access columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func lowerLookupConstraintToAir(c LookupConstraint, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	var (
		source = lowerLookupVector(c.Source, mirSchema, airSchema, cfg)
		target = lowerLookupVector(c.Target, mirSchema, airSchema, cfg)
	)
	//
	airSchema.AddLookupConstraint(c.Handle, source, target)
}

func lowerLookupVector(c LookupVector, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) air.LookupVector {
	// Make decision on whether to use legacy translation or optimal translation
	if cfg.LegacyLookups {
		return lowerLegacyLookupVector(c, mirSchema, airSchema, cfg)
	}
	// Optimial
	return lowerConditionalLookupVector(c, mirSchema, airSchema, cfg)
}

func lowerConditionalLookupVector(c LookupVector, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) air.LookupVector {
	//
	var terms = make([]*air.ColumnAccess, c.Len())
	// lower terms
	for i := range c.Len() {
		terms[i] = lowerLookupTerm(c.Context(), c.Ith(i), mirSchema, airSchema, cfg)
	}
	// lower selector (if applicable)
	if c.HasSelector() {
		selector := lowerLookupTerm(c.Context(), c.Selector.Unwrap(), mirSchema, airSchema, cfg)
		// Optimal translation
		return constraint.FilteredLookupVector(c.Context(), selector, terms...)
	}
	// no selector
	return constraint.UnfilteredLookupVector(c.Context(), terms...)
}

func lowerLegacyLookupVector(c LookupVector, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) air.LookupVector {
	//
	var terms = make([]*air.ColumnAccess, c.Len())
	// lower terms
	for i := range c.Len() {
		ith := c.Ith(i)
		// Multiply out selector (if applicable)
		if c.HasSelector() {
			ith = Product(c.Selector.Unwrap(), ith)
		}
		//
		terms[i] = lowerLookupTerm(c.Context(), ith, mirSchema, airSchema, cfg)
	}
	// no selector
	return constraint.UnfilteredLookupVector(c.Context(), terms...)
}

func lowerLookupTerm(context tr.Context, expr Expr, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) *air.ColumnAccess {
	// Determine bitwidth
	bitwidth := rangeOfTerm(expr.term, mirSchema).BitWidth()
	// Lower selector expression
	term := lowerExprTo(context, expr, mirSchema, airSchema, cfg)
	// Expand expression into a column identifier
	cid := air_gadgets.Expand(context, bitwidth, term, airSchema)
	//
	return &air.ColumnAccess{Column: cid, Shift: 0}
}

// Lower a sorted constraint to the AIR level.  The challenge here is that there
// is not concept of sorting constraints at the AIR level.  Instead, we have to
// generate the necessary machinery to enforce the sorting constraint.
func lowerSortedConstraintToAir(c SortedConstraint, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	sources := make([]uint, len(c.Sources))
	//
	for i := 0; i < len(sources); i++ {
		sourceBitwidth := rangeOfTerm(c.Sources[i].term, mirSchema).BitWidth()
		// Lower source expression
		source := lowerExprTo(c.Context, c.Sources[i], mirSchema, airSchema, cfg)
		// Expand them
		sources[i] = air_gadgets.Expand(c.Context, sourceBitwidth, source, airSchema)
	}
	// Determine number of ordered columns
	numSignedCols := len(c.Signs)
	// For a multi column sort, its a bit harder as we need additional
	// logic to ensure the target columns are lexicographally sorted.
	gadget := air_gadgets.NewLexicographicSortingGadget(c.Handle, sources, c.BitWidth)
	gadget.SetSigns(c.Signs...)
	gadget.SetStrict(c.Strict)
	gadget.SetMaxRangeConstraint(cfg.MaxRangeConstraint)
	gadget.SetLegacyTypeProofs(cfg.LegacyTypeProofs)
	// Add (optional) selector
	if c.Selector.HasValue() {
		selector := lowerExprTo(c.Context, c.Selector.Unwrap(), mirSchema, airSchema, cfg)
		gadget.SetSelector(selector)
	}
	// Done
	gadget.Apply(airSchema)
	// Sanity check bitwidth
	bitwidth := uint(0)

	for i := 0; i < numSignedCols; i++ {
		// Extract bitwidth of ith column
		ith := mirSchema.Columns().Nth(sources[i]).DataType.AsUint().BitWidth()
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

// Lower a permutation to the AIR level.  This has quite a few
// effects.  Firstly, permutation constraints are added for all of the
// new columns.  Secondly, sorting constraints (and their associated
// computed columns) must also be added.  Finally, a trace
// computation is required to ensure traces are correctly expanded to
// meet the requirements of a sorted permutation.
func lowerPermutationToAir(c Permutation, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	builder := strings.Builder{}
	c_targets := c.Targets
	targets := make([]uint, len(c_targets))
	//
	builder.WriteString("permutation")
	// Add individual permutation constraints
	for i := 0; i < len(c_targets); i++ {
		var ok bool
		// TODO: how best to avoid this lookup?
		targets[i], ok = sc.ColumnIndexOf(airSchema, c.Module(), c_targets[i].Name)
		//
		if !ok {
			panic("internal failure")
		}
		//
		builder.WriteString(fmt.Sprintf(":%s", c_targets[i].Name))
	}
	//
	airSchema.AddPermutationConstraint(builder.String(), c.Context(), targets, c.Sources)
	// Determine number of ordered columns
	numSignedCols := len(c.Signs)
	// Add sorting constraints + computed columns as necessary.
	// For a multi column sort, its a bit harder as we need additional
	// logic to ensure the target columns are lexicographally sorted.
	bitwidth := uint(0)

	for i := 0; i < numSignedCols; i++ {
		// Extract bitwidth of ith column
		ith := mirSchema.Columns().Nth(c.Sources[i]).DataType.AsUint().BitWidth()
		if ith > bitwidth {
			bitwidth = ith
		}
	}
	// Construct a unique prefix for this sort.
	prefix := constructLexicographicSortingPrefix(targets, c.Signs, airSchema)
	// Add lexicographically sorted constraints
	gadget := air_gadgets.NewLexicographicSortingGadget(prefix, targets, bitwidth)
	gadget.SetSigns(c.Signs...)
	gadget.SetMaxRangeConstraint(cfg.MaxRangeConstraint)
	gadget.SetLegacyTypeProofs(cfg.LegacyTypeProofs)
	// Done
	gadget.Apply(airSchema)
}

func lowerConstraintTo(ctx trace.Context, c Constraint, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) air.Expr {
	//
	es := make([]air.Expr, len(c.terms))
	//
	for i, t := range c.terms {
		// Optimise normalisations
		t1 := eliminateNormalisationInTerm(t, mirSchema, cfg)
		// Apply constant propagation
		t1 = constantPropagationForTerm(t1, false, airSchema)
		// Lower properly
		es[i] = lowerTermToInner(ctx, t1, mirSchema, airSchema, cfg)
	}
	// Simple optimisation; we could do more here.
	if len(es) == 1 {
		return es[0]
	}
	//
	return air.Product(es...)
}

// Lower an expression into the Arithmetic Intermediate Representation.
// Essentially, this means eliminating normalising expressions by introducing
// new columns into the given table (with appropriate constraints).  This first
// performs constant propagation to ensure lowering is as efficient as possible.
// A module identifier is required to determine where any computed columns
// should be located.
func lowerExprTo(ctx trace.Context, e1 Expr, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) air.Expr {
	// Optimise normalisations
	t1 := eliminateNormalisationInTerm(e1.term, mirSchema, cfg)
	// Apply constant propagation
	t1 = constantPropagationForTerm(t1, false, airSchema)
	// Lower properly
	return lowerTermToInner(ctx, t1, mirSchema, airSchema, cfg)
}

// Inner form is used for recursive calls and does not repeat the constant
// propagation phase.
func lowerTermToInner(ctx trace.Context, e Term, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) air.Expr {
	//
	switch e := e.(type) {
	case *Add:
		args := lowerTerms(ctx, e.Args, mirSchema, airSchema, cfg)
		return air.Sum(args...)
	case *Cast:
		return lowerTermToInner(ctx, e.Arg, mirSchema, airSchema, cfg)
	case *Constant:
		return air.NewConst(e.Value)
	case *ColumnAccess:
		return air.NewColumnAccess(e.Column, e.Shift)
	case *Exp:
		return lowerExpTo(ctx, e, mirSchema, airSchema, cfg)
	case *Mul:
		args := lowerTerms(ctx, e.Args, mirSchema, airSchema, cfg)
		return air.Product(args...)
	case *Norm:
		// Lower the expression being normalised
		arg := lowerTermToInner(ctx, e.Arg, mirSchema, airSchema, cfg)
		// Determine appropriate shift
		shift := 0
		//  Apply shift normalisation (if enabled)
		if cfg.ShiftNormalisation {
			// Determine shift ranges
			min, max := shiftRangeOfTerm(e.Arg)
			// determine shift amount
			if max < 0 {
				shift = max
			} else if min > 0 {
				shift = min
			}
		}
		// Construct an expression representing the normalised value of e.  That is,
		// an expression which is 0 when e is 0, and 1 when e is non-zero.
		norm := air_gadgets.Normalise(arg.Shift(-shift), airSchema)
		//
		return norm.Shift(shift)
	case *Sub:
		args := lowerTerms(ctx, e.Args, mirSchema, airSchema, cfg)
		return air.Subtract(args...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

// LowerTo lowers an exponent expression to the AIR level by lowering the
// argument, and then constructing a multiplication.  This is because the AIR
// level does not support an explicit exponent operator.
func lowerExpTo(ctx trace.Context, e *Exp, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) air.Expr {
	// Lower the expression being raised
	le := lowerTermToInner(ctx, e.Arg, mirSchema, airSchema, cfg)
	// Multiply it out k times
	es := make([]air.Expr, e.Pow)
	//
	for i := uint64(0); i < e.Pow; i++ {
		es[i] = le
	}
	// Done
	return air.Product(es...)
}

// Lower a set of zero or more MIR expressions.
func lowerTerms(ctx trace.Context, exprs []Term, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) []air.Expr {
	//
	n := len(exprs)
	nexprs := make([]air.Expr, n)

	for i := 0; i < n; i++ {
		nexprs[i] = lowerTermToInner(ctx, exprs[i], mirSchema, airSchema, cfg)
	}

	return nexprs
}

// Construct a unique identifier for the given sort.  This should not conflict
// with the identifier for any other sort.
func constructLexicographicSortingPrefix(columns []uint, signs []bool, schema *air.Schema) string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder
	// Concatenate column names with their signs.
	for i := 0; i < len(columns); i++ {
		ith := schema.Columns().Nth(columns[i])
		id.WriteString(ith.Name)

		if i >= len(signs) {

		} else if signs[i] {
			id.WriteString("+")
		} else {
			id.WriteString("-")
		}
	}
	// Done
	return id.String()
}
