package mir

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/air"
	air_gadgets "github.com/consensys/go-corset/pkg/air/gadgets"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
)

// LowerToAir lowers (or refines) an MIR table into an AIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToAir() *air.Schema {
	airSchema := air.EmptySchema[Expr]()
	// Copy modules
	for _, mod := range p.modules {
		airSchema.AddModule(mod.Name())
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
		lowerConstraintToAir(c, airSchema)
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
	} else {
		panic("unknown assignment")
	}
}

// Lower a constraint to the AIR level.
func lowerConstraintToAir(c sc.Constraint, schema *air.Schema) {
	// Check what kind of constraint we have
	if v, ok := c.(LookupConstraint); ok {
		lowerLookupConstraintToAir(v, schema)
	} else if v, ok := c.(VanishingConstraint); ok {
		air_expr := lowerExprTo(v.Constraint().Expr, schema)
		schema.AddVanishingConstraint(v.Handle(), v.Context(), v.Domain(), air_expr)
	} else if v, ok := c.(*constraint.TypeConstraint); ok {
		if t := v.Type().AsUint(); t != nil {
			// Yes, a constraint is implied.  Now, decide whether to use a range
			// constraint or just a vanishing constraint.
			if t.HasBound(2) {
				// u1 => use vanishing constraint X * (X - 1)
				air_gadgets.ApplyBinaryGadget(v.Target(), schema)
			} else if t.HasBound(256) {
				// u2..8 use range constraints
				schema.AddRangeConstraint(v.Target(), t.Bound())
			} else {
				// u9+ use byte decompositions.
				air_gadgets.ApplyBitwidthGadget(v.Target(), t.BitWidth(), schema)
			}
		}
	} else {
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

// Lower a lookup constraint to the AIR level.  The challenge here is that a
// lookup constraint at the AIR level cannot use arbitrary expressions; rather,
// it can only access columns directly.  Therefore, whenever a general
// expression is encountered, we must generate a computed column to hold the
// value of that expression, along with appropriate constraints to enforce the
// expected value.
func lowerLookupConstraintToAir(c LookupConstraint, schema *air.Schema) {
	targets := make([]uint, len(c.Targets()))
	sources := make([]uint, len(c.Sources()))
	//
	for i := 0; i < len(targets); i++ {
		// Lower source and target expressions
		target := lowerExprTo(c.Targets()[i], schema)
		source := lowerExprTo(c.Sources()[i], schema)
		// Expand them
		targets[i] = air_gadgets.Expand(target, schema)
		sources[i] = air_gadgets.Expand(source, schema)
	}
	// finally add the constraint
	schema.AddLookupConstraint(c.Handle(), c.SourceContext(), c.TargetContext(), sources, targets)
}

// Lower a permutation to the AIR level.  This has quite a few
// effects.  Firstly, permutation constraints are added for all of the
// new columns.  Secondly, sorting constraints (and their associated
// computed columns) must also be added.  Finally, a trace
// computation is required to ensure traces are correctly expanded to
// meet the requirements of a sorted permutation.
func lowerPermutationToAir(c Permutation, mirSchema *Schema, airSchema *air.Schema) {
	c_targets := c.Targets()
	ncols := len(c_targets)
	//
	targets := make([]uint, ncols)
	// Add individual permutation constraints
	for i := 0; i < ncols; i++ {
		var ok bool
		// TODO: how best to avoid this lookup?
		targets[i], ok = sc.ColumnIndexOf(airSchema, c.Module(), c_targets[i].Name())

		if !ok {
			panic("internal failure")
		}
	}
	//
	airSchema.AddPermutationConstraint(targets, c.Sources())
	// Add sorting constraints + computed columns as necessary.
	if ncols == 1 {
		// For a single column sort, its actually a bit easier because we don't
		// need to implement a multiplexor (i.e. to determine which column is
		// differs, etc).  Instead, we just need a delta column which ensures
		// there is a non-negative difference between consecutive rows.  This
		// also requires bitwidth constraints.
		bitwidth := mirSchema.Columns().Nth(c.Sources()[0]).Type().AsUint().BitWidth()
		// Add column sorting constraints
		air_gadgets.ApplyColumnSortGadget(targets[0], c.Signs()[0], bitwidth, airSchema)
	} else {
		// For a multi column sort, its a bit harder as we need additional
		// logicl to ensure the target columns are lexicographally sorted.
		bitwidth := uint(0)

		for i := 0; i < ncols; i++ {
			// Extract bitwidth of ith column
			ith := mirSchema.Columns().Nth(c.Sources()[i]).Type().AsUint().BitWidth()
			if ith > bitwidth {
				bitwidth = ith
			}
		}
		// Add lexicographically sorted constraints
		air_gadgets.ApplyLexicographicSortingGadget(targets, c.Signs(), bitwidth, airSchema)
	}
}

// Lower an expression into the Arithmetic Intermediate Representation.
// Essentially, this means eliminating normalising expressions by introducing
// new columns into the given table (with appropriate constraints).  This first
// performs constant propagation to ensure lowering is as efficient as possible.
func lowerExprTo(e1 Expr, schema *air.Schema) air.Expr {
	// Apply constant propagation
	e2 := applyConstantPropagation(e1)
	// Lower properly
	return lowerExprToInner(e2, schema)
}

// Inner form is used for recursive calls and does not repeat the constant
// propagation phase.
func lowerExprToInner(e Expr, schema *air.Schema) air.Expr {
	if p, ok := e.(*Add); ok {
		return &air.Add{Args: lowerExprs(p.Args, schema)}
	} else if p, ok := e.(*Constant); ok {
		return &air.Constant{Value: p.Value}
	} else if p, ok := e.(*ColumnAccess); ok {
		return &air.ColumnAccess{Column: p.Column, Shift: p.Shift}
	} else if p, ok := e.(*Mul); ok {
		return &air.Mul{Args: lowerExprs(p.Args, schema)}
	} else if p, ok := e.(*Exp); ok {
		return lowerExpTo(p, schema)
	} else if p, ok := e.(*Normalise); ok {
		// Lower the expression being normalised
		e := lowerExprToInner(p.Arg, schema)
		// Construct an expression representing the normalised value of e.  That is,
		// an expression which is 0 when e is 0, and 1 when e is non-zero.
		return air_gadgets.Normalise(e, schema)
	} else if p, ok := e.(*Sub); ok {
		return &air.Sub{Args: lowerExprs(p.Args, schema)}
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", e.String()))
}

// LowerTo lowers an exponent expression to the AIR level by lowering the
// argument, and then constructing a multiplication.  This is because the AIR
// level does not support an explicit exponent operator.
func lowerExpTo(e *Exp, schema *air.Schema) air.Expr {
	// Lower the expression being raised
	le := lowerExprToInner(e.Arg, schema)
	// Multiply it out k times
	es := make([]air.Expr, e.Pow)
	//
	for i := uint64(0); i < e.Pow; i++ {
		es[i] = le
	}
	// Done
	return &air.Mul{Args: es}
}

// Lower a set of zero or more MIR expressions.
func lowerExprs(exprs []Expr, schema *air.Schema) []air.Expr {
	n := len(exprs)
	nexprs := make([]air.Expr, n)

	for i := 0; i < n; i++ {
		nexprs[i] = lowerExprToInner(exprs[i], schema)
	}

	return nexprs
}
