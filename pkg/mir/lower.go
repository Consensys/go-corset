package mir

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/air"
	air_gadgets "github.com/consensys/go-corset/pkg/air/gadgets"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
)

// LowerToAir lowers (or refines) an MIR table into an AIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToAir() *air.Schema {
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
		lowerConstraintToAir(c, p, airSchema)
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
func lowerConstraintToAir(c sc.Constraint, mirSchema *Schema, airSchema *air.Schema) {
	// Check what kind of constraint we have
	if v, ok := c.(LookupConstraint); ok {
		lowerLookupConstraintToAir(v, mirSchema, airSchema)
	} else if v, ok := c.(VanishingConstraint); ok {
		lowerVanishingConstraintToAir(v, mirSchema, airSchema)
	} else if v, ok := c.(RangeConstraint); ok {
		lowerRangeConstraintToAir(v, mirSchema, airSchema)
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
func lowerVanishingConstraintToAir(v VanishingConstraint, mirSchema *Schema, airSchema *air.Schema) {
	air_expr := lowerExprTo(v.Context, v.Constraint.Expr, mirSchema, airSchema)
	// Check whether this is a constant
	constant := air_expr.AsConstant()
	// Check for compile-time constants
	if constant != nil && !constant.IsZero() {
		panic(fmt.Sprintf("constraint %s cannot vanish!", v.Handle))
	} else if constant == nil {
		airSchema.AddVanishingConstraint(v.Handle, v.Context, v.Domain, air_expr)
	}
}

// Lower a range constraint to the AIR level.  The challenge here is that a
// range constraint at the AIR level cannot use arbitrary expressions; rather it
// can only constrain columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func lowerRangeConstraintToAir(v RangeConstraint, mirSchema *Schema, airSchema *air.Schema) {
	bitwidth := v.Expr.IntRange(mirSchema).BitWidth()
	// Lower target expression
	target := lowerExprTo(v.Context, v.Expr, mirSchema, airSchema)
	// Expand target expression (if necessary)
	column := air_gadgets.Expand(v.Context, bitwidth, target, airSchema)
	// Yes, a constraint is implied.  Now, decide whether to use a range
	// constraint or just a vanishing constraint.
	if v.BoundedAtMost(2) {
		// u1 => use vanishing constraint X * (X - 1)
		air_gadgets.ApplyBinaryGadget(column, airSchema)
	} else if v.BoundedAtMost(256) {
		// u2..8 use range constraints
		airSchema.AddRangeConstraint(column, v.Bound)
	} else {
		// u9+ use byte decompositions.
		var bi big.Int
		// Convert bound into big int
		elem := v.Bound
		elem.BigInt(&bi)
		// Apply bitwidth gadget
		air_gadgets.ApplyBitwidthGadget(column, uint(bi.BitLen()-1), airSchema)
	}
}

// Lower a lookup constraint to the AIR level.  The challenge here is that a
// lookup constraint at the AIR level cannot use arbitrary expressions; rather,
// it can only access columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func lowerLookupConstraintToAir(c LookupConstraint, mirSchema *Schema, airSchema *air.Schema) {
	targets := make([]uint, len(c.Targets))
	sources := make([]uint, len(c.Sources))
	//
	for i := 0; i < len(targets); i++ {
		targetBitwidth := c.Targets[i].IntRange(mirSchema).BitWidth()
		sourceBitwidth := c.Sources[i].IntRange(mirSchema).BitWidth()
		// Lower source and target expressions
		target := lowerExprTo(c.TargetContext, c.Targets[i], mirSchema, airSchema)
		source := lowerExprTo(c.SourceContext, c.Sources[i], mirSchema, airSchema)
		// Expand them
		targets[i] = air_gadgets.Expand(c.TargetContext, targetBitwidth, target, airSchema)
		sources[i] = air_gadgets.Expand(c.SourceContext, sourceBitwidth, source, airSchema)
	}
	// finally add the constraint
	airSchema.AddLookupConstraint(c.Handle, c.SourceContext, c.TargetContext, sources, targets)
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
	ncols := len(c_targets)
	targets := make([]uint, ncols)
	//
	builder.WriteString("permutation")
	// Add individual permutation constraints
	for i := 0; i < ncols; i++ {
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
	airSchema.AddPermutationConstraint(builder.String(), targets, c.Sources)
	// Add sorting constraints + computed columns as necessary.
	if ncols == 1 {
		// For a single column sort, its actually a bit easier because we don't
		// need to implement a multiplexor (i.e. to determine which column is
		// differs, etc).  Instead, we just need a delta column which ensures
		// there is a non-negative difference between consecutive rows.  This
		// also requires bitwidth constraints.
		bitwidth := mirSchema.Columns().Nth(c.Sources[0]).DataType.AsUint().BitWidth()
		// Add column sorting constraints
		air_gadgets.ApplyColumnSortGadget(targets[0], c.Signs[0], bitwidth, airSchema)
	} else {
		// For a multi column sort, its a bit harder as we need additional
		// logicl to ensure the target columns are lexicographally sorted.
		bitwidth := uint(0)

		for i := 0; i < ncols; i++ {
			// Extract bitwidth of ith column
			ith := mirSchema.Columns().Nth(c.Sources[i]).DataType.AsUint().BitWidth()
			if ith > bitwidth {
				bitwidth = ith
			}
		}
		// Add lexicographically sorted constraints
		air_gadgets.ApplyLexicographicSortingGadget(targets, c.Signs, bitwidth, airSchema)
	}
}

// Lower an expression into the Arithmetic Intermediate Representation.
// Essentially, this means eliminating normalising expressions by introducing
// new columns into the given table (with appropriate constraints).  This first
// performs constant propagation to ensure lowering is as efficient as possible.
// A module identifier is required to determine where any computed columns
// should be located.
func lowerExprTo(ctx trace.Context, e1 Expr, mirSchema *Schema, airSchema *air.Schema) air.Expr {
	// Apply constant propagation
	e2 := applyConstantPropagation(e1, airSchema)
	// Lower properly
	return lowerExprToInner(ctx, e2, mirSchema, airSchema)
}

// Inner form is used for recursive calls and does not repeat the constant
// propagation phase.
func lowerExprToInner(ctx trace.Context, e Expr, mirSchema *Schema, airSchema *air.Schema) air.Expr {
	if p, ok := e.(*Add); ok {
		args := lowerExprs(ctx, p.Args, mirSchema, airSchema)
		return air.Sum(args...)
	} else if p, ok := e.(*Constant); ok {
		return air.NewConst(p.Value)
	} else if p, ok := e.(*ColumnAccess); ok {
		return air.NewColumnAccess(p.Column, p.Shift)
	} else if p, ok := e.(*Mul); ok {
		args := lowerExprs(ctx, p.Args, mirSchema, airSchema)
		return air.Product(args...)
	} else if p, ok := e.(*Exp); ok {
		return lowerExpTo(ctx, p, mirSchema, airSchema)
	} else if p, ok := e.(*Normalise); ok {
		bounds := p.Arg.IntRange(mirSchema)
		// Lower the expression being normalised
		e := lowerExprToInner(ctx, p.Arg, mirSchema, airSchema)
		// Check whether normalisation actual required.  For example, if the
		// argument is just a binary column then a normalisation is not actually
		// required.
		if bounds.Within(big.NewInt(0), big.NewInt(1)) {
			return e
		}
		// Construct an expression representing the normalised value of e.  That is,
		// an expression which is 0 when e is 0, and 1 when e is non-zero.
		return air_gadgets.Normalise(e, airSchema)
	} else if p, ok := e.(*Sub); ok {
		args := lowerExprs(ctx, p.Args, mirSchema, airSchema)
		return air.Subtract(args...)
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", e.Lisp(airSchema).String(true)))
}

// LowerTo lowers an exponent expression to the AIR level by lowering the
// argument, and then constructing a multiplication.  This is because the AIR
// level does not support an explicit exponent operator.
func lowerExpTo(ctx trace.Context, e *Exp, mirSchema *Schema, airSchema *air.Schema) air.Expr {
	// Lower the expression being raised
	le := lowerExprToInner(ctx, e.Arg, mirSchema, airSchema)
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
func lowerExprs(ctx trace.Context, exprs []Expr, mirSchema *Schema, airSchema *air.Schema) []air.Expr {
	n := len(exprs)
	nexprs := make([]air.Expr, n)

	for i := 0; i < n; i++ {
		nexprs[i] = lowerExprToInner(ctx, exprs[i], mirSchema, airSchema)
	}

	return nexprs
}
