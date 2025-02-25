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
	"math/big"
	"reflect"
	"strings"

	"github.com/consensys/go-corset/pkg/air"
	air_gadgets "github.com/consensys/go-corset/pkg/air/gadgets"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// OptimisationConfig provides a mechanism for controlling how optimisations are
// applied during MIR lowering.
type OptimisationConfig struct {
	// InverseEliminationLevel sets an upper bound on the range cardinality at
	// which inverses will be eliminated in favour of constraints.  A level of 0
	// means no inverses will be eliminated, a range of 1 means only trivial
	// ranges (i.e. {-1,0}, {0,1} and {-1,0,1}) will be eliminated; Otherwise,
	// the level indicates the range cardinality.  For example, level 2 means
	// any range of cardinality 2 is eliminated (e.g. {1,2}, {5,6}, etc).
	InverseEliminiationLevel uint
	// MaxRangeConstraint determines an upper bound on which MIR range
	// constraints are translated in AIR range constraints, versus using a
	// horizontal bitwidth gadget.
	MaxRangeConstraint uint
}

// OPTIMISATION_LEVELS provides a set of precanned optimisation configurations.
// Here 0 implies no optimisation and, otherwise, increasing levels implies
// increasingly aggressive optimisation (though that doesn't mean they will
// always improve performance).
var OPTIMISATION_LEVELS = []OptimisationConfig{
	// Level 0 == nothing enabled
	{0, 256},
	// Level 1 == minimal optimisations applied.
	{1, 256},
}

// DEFAULT_OPTIMISATION_LEVEL provides a default level of optimisation which
// should be used in most cases.
var DEFAULT_OPTIMISATION_LEVEL = OPTIMISATION_LEVELS[1]

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
		lowerAssignmentToAir(assign, p, airSchema)
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
func lowerAssignmentToAir(c sc.Assignment, mirSchema *Schema, airSchema *air.Schema) {
	if v, ok := c.(Permutation); ok {
		lowerPermutationToAir(v, mirSchema, airSchema)
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
	air_expr := lowerExprTo(v.Context, v.Constraint, mirSchema, airSchema, cfg)
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
	bitwidth := v.Expr.IntRange(mirSchema).BitWidth()
	// Lower target expression
	target := lowerExprTo(v.Context, v.Expr, mirSchema, airSchema, cfg)
	// Expand target expression (if necessary)
	column := air_gadgets.Expand(v.Context, bitwidth, target, airSchema)
	// Yes, a constraint is implied.  Now, decide whether to use a range
	// constraint or just a vanishing constraint.
	if v.BoundedAtMost(2) {
		// u1 => use vanishing constraint X * (X - 1)
		air_gadgets.ApplyBinaryGadget(column, airSchema)
	} else if v.BoundedAtMost(cfg.MaxRangeConstraint) {
		// u2..n use range constraints
		airSchema.AddRangeConstraint(column, v.Case, v.Bound)
	} else {
		// remainder use horizontal byte decompositions.
		var bi big.Int
		// Convert bound into big int
		elem := v.Bound
		elem.BigInt(&bi)
		// Apply bitwidth gadget
		air_gadgets.ApplyBitwidthGadget(column, uint(bi.BitLen()-1), air.NewConst64(1), airSchema)
	}
}

// Lower a lookup constraint to the AIR level.  The challenge here is that a
// lookup constraint at the AIR level cannot use arbitrary expressions; rather,
// it can only access columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func lowerLookupConstraintToAir(c LookupConstraint, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	targets := make([]uint, len(c.Targets))
	sources := make([]uint, len(c.Sources))
	//
	for i := 0; i < len(targets); i++ {
		targetBitwidth := c.Targets[i].IntRange(mirSchema).BitWidth()
		sourceBitwidth := c.Sources[i].IntRange(mirSchema).BitWidth()
		// Lower source and target expressions
		target := lowerExprTo(c.TargetContext, c.Targets[i], mirSchema, airSchema, cfg)
		source := lowerExprTo(c.SourceContext, c.Sources[i], mirSchema, airSchema, cfg)
		// Expand them
		targets[i] = air_gadgets.Expand(c.TargetContext, targetBitwidth, target, airSchema)
		sources[i] = air_gadgets.Expand(c.SourceContext, sourceBitwidth, source, airSchema)
	}
	// finally add the constraint
	airSchema.AddLookupConstraint(c.Handle, c.SourceContext, c.TargetContext, sources, targets)
}

// Lower a sorted constraint to the AIR level.  The challenge here is that there
// is not concept of sorting constraints at the AIR level.  Instead, we have to
// generate the necessary machinery to enforce the sorting constraint.
func lowerSortedConstraintToAir(c SortedConstraint, mirSchema *Schema, airSchema *air.Schema, cfg OptimisationConfig) {
	sources := make([]uint, len(c.Sources))
	//
	for i := 0; i < len(sources); i++ {
		sourceBitwidth := c.Sources[i].IntRange(mirSchema).BitWidth()
		// Lower source expression
		source := lowerExprTo(c.Context, c.Sources[i], mirSchema, airSchema, cfg)
		// Expand them
		sources[i] = air_gadgets.Expand(c.Context, sourceBitwidth, source, airSchema)
	}
	// Determine number of ordered columns
	numSignedCols := len(c.Signs)
	// finally add the constraint
	if numSignedCols == 1 {
		// For a single column sort, its actually a bit easier because we don't
		// need to implement a multiplexor (i.e. to determine which column is
		// differs, etc).  Instead, we just need a delta column which ensures
		// there is a non-negative difference between consecutive rows.  This
		// also requires bitwidth constraints.
		gadget := air_gadgets.NewColumnSortGadget(c.Handle, sources[0], c.BitWidth)
		gadget.SetSign(c.Signs[0])
		gadget.SetStrict(c.Strict)
		// Add (optional) selector
		if c.Selector.HasValue() {
			selector := lowerExprTo(c.Context, c.Selector.Unwrap(), mirSchema, airSchema, cfg)
			gadget.SetSelector(selector)
		}
		// Done!
		gadget.Apply(airSchema)
	} else {
		// For a multi column sort, its a bit harder as we need additional
		// logic to ensure the target columns are lexicographally sorted.
		gadget := air_gadgets.NewLexicographicSortingGadget(c.Handle, sources, c.BitWidth)
		gadget.SetSigns(c.Signs...)
		gadget.SetStrict(c.Strict)
		// Add (optional) selector
		if c.Selector.HasValue() {
			selector := lowerExprTo(c.Context, c.Selector.Unwrap(), mirSchema, airSchema, cfg)
			gadget.SetSelector(selector)
		}
		// Done
		gadget.Apply(airSchema)
	}
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
func lowerPermutationToAir(c Permutation, mirSchema *Schema, airSchema *air.Schema) {
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
	if numSignedCols == 1 {
		// For a single column sort, its actually a bit easier because we don't
		// need to implement a multiplexor (i.e. to determine which column is
		// differs, etc).  Instead, we just need a delta column which ensures
		// there is a non-negative difference between consecutive rows.  This
		// also requires bitwidth constraints.
		bitwidth := mirSchema.Columns().Nth(c.Sources[0]).DataType.AsUint().BitWidth()
		// Identify target column name
		target := mirSchema.Columns().Nth(targets[0]).Name
		// Add column sorting constraints
		gadget := air_gadgets.NewColumnSortGadget(target, targets[0], bitwidth)
		gadget.SetSign(c.Signs[0])
		// Done!
		gadget.Apply(airSchema)
	} else {
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
		// For a multi column sort, its a bit harder as we need additional
		// logic to ensure the target columns are lexicographally sorted.
		gadget := air_gadgets.NewLexicographicSortingGadget(prefix, targets, bitwidth)
		gadget.SetSigns(c.Signs...)
		// Done
		gadget.Apply(airSchema)
	}
}

// Lower an expression into the Arithmetic Intermediate Representation.
// Essentially, this means eliminating normalising expressions by introducing
// new columns into the given table (with appropriate constraints).  This first
// performs constant propagation to ensure lowering is as efficient as possible.
// A module identifier is required to determine where any computed columns
// should be located.
func lowerExprTo(ctx trace.Context, e1 Expr, mirSchema *Schema, airSchema *air.Schema,
	cfg OptimisationConfig) air.Expr {
	// Apply constant propagation
	t1 := constantPropagationForTerm(e1.term, airSchema)
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
		bounds := rangeOfTerm(e.Arg, mirSchema)
		// Lower the expression being normalised
		arg := lowerTermToInner(ctx, e.Arg, mirSchema, airSchema, cfg)
		// Check whether normalisation actually required.  For example, if the
		// argument is just a binary column then a normalisation is not actually
		// required.
		if cfg.InverseEliminiationLevel > 0 && bounds.Within(util.NewInterval64(0, 1)) {
			// arg ∈ {0,1} ==> normalised already :)
			return arg
		} else if cfg.InverseEliminiationLevel > 0 && bounds.Within(util.NewInterval64(-1, 1)) {
			// arg ∈ {-1,0,1} ==> (arg*arg) ∈ {0,1}
			return air.Product(arg, arg)
		}
		// Construct an expression representing the normalised value of e.  That is,
		// an expression which is 0 when e is 0, and 1 when e is non-zero.
		return air_gadgets.Normalise(arg, airSchema)
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
