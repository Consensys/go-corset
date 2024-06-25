package mir

import (
	"github.com/consensys/go-corset/pkg/air"
	air_gadgets "github.com/consensys/go-corset/pkg/air/gadgets"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
)

// LowerToAir lowers (or refines) an MIR table into an AIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToAir() *air.Schema {
	airSchema := air.EmptySchema[Expr]()
	// Add data columns.
	for _, c := range p.inputs {
		col := c.(DataColumn)
		airSchema.AddColumn(col.Name(), col.Type())
	}
	// Add Assignments. Again this has to be done first for things to work.
	// Essentially to reflect the fact that these columns have been added above
	// before others.  Realistically, the overall design of this process is a
	// bit broken right now.
	for _, perm := range p.assignments {
		airSchema.AddAssignment(perm.(Permutation))
	}
	// Lower permutations columns
	for _, perm := range p.assignments {
		lowerPermutationToAir(perm.(Permutation), p, airSchema)
	}
	// Lower vanishing constraints
	for _, c := range p.constraints {
		lowerConstraintToAir(c, airSchema)
	}
	// Done
	return airSchema
}

// Lower a constraint to the AIR level.
func lowerConstraintToAir(c sc.Constraint, schema *air.Schema) {
	// Check what kind of constraint we have
	if v, ok := c.(VanishingConstraint); ok {
		air_expr := v.Constraint().Expr.LowerTo(schema)
		schema.AddVanishingConstraint(v.Handle(), v.Domain(), air_expr)
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
	sources := make([]uint, ncols)
	// Add individual permutation constraints
	for i := 0; i < ncols; i++ {
		var ok1, ok2 bool
		// TODO: REPLACE
		sources[i], ok1 = sc.ColumnIndexOf(airSchema, c.Sources[i])
		targets[i], ok2 = sc.ColumnIndexOf(airSchema, c_targets[i].Name())

		if !ok1 || !ok2 {
			panic("missing column")
		}
	}
	//
	airSchema.AddPermutationConstraint(targets, sources)
	// Add sorting constraints + computed columns as necessary.
	if ncols == 1 {
		// For a single column sort, its actually a bit easier because we don't
		// need to implement a multiplexor (i.e. to determine which column is
		// differs, etc).  Instead, we just need a delta column which ensures
		// there is a non-negative difference between consecutive rows.  This
		// also requires bitwidth constraints.
		bitwidth := schema.ColumnByName(mirSchema, c.Sources[0]).Type().AsUint().BitWidth()
		// Add column sorting constraints
		air_gadgets.ApplyColumnSortGadget(targets[0], c.Signs[0], bitwidth, airSchema)
	} else {
		// For a multi column sort, its a bit harder as we need additional
		// logicl to ensure the target columns are lexicographally sorted.
		bitwidth := uint(0)

		for i := 0; i < ncols; i++ {
			// Extract bitwidth of ith column
			ith := schema.ColumnByName(mirSchema, c.Sources[i]).Type().AsUint().BitWidth()
			if ith > bitwidth {
				bitwidth = ith
			}
		}
		// Add lexicographically sorted constraints
		air_gadgets.ApplyLexicographicSortingGadget(targets, c.Signs, bitwidth, airSchema)
	}
}

// LowerTo lowers a sum expression to the AIR level by lowering the arguments.
func (e *Add) LowerTo(schema *air.Schema) air.Expr {
	return &air.Add{Args: lowerExprs(e.Args, schema)}
}

// LowerTo lowers a subtract expression to the AIR level by lowering the arguments.
func (e *Sub) LowerTo(schema *air.Schema) air.Expr {
	return &air.Sub{Args: lowerExprs(e.Args, schema)}
}

// LowerTo lowers a product expression to the AIR level by lowering the arguments.
func (e *Mul) LowerTo(schema *air.Schema) air.Expr {
	return &air.Mul{Args: lowerExprs(e.Args, schema)}
}

// LowerTo lowers a normalise expression to the AIR level by "compiling it out"
// using a computed column.
func (p *Normalise) LowerTo(schema *air.Schema) air.Expr {
	// Lower the expression being normalised
	e := p.Arg.LowerTo(schema)
	// Construct an expression representing the normalised value of e.  That is,
	// an expression which is 0 when e is 0, and 1 when e is non-zero.
	return air_gadgets.Normalise(e, schema)
}

// LowerTo lowers a column access to the AIR level.  This is straightforward as
// it is already in the correct form.
func (e *ColumnAccess) LowerTo(schema *air.Schema) air.Expr {
	return &air.ColumnAccess{Column: e.Column, Shift: e.Shift}
}

// LowerTo lowers a constant to the AIR level.  This is straightforward as it is
// already in the correct form.
func (e *Constant) LowerTo(schema *air.Schema) air.Expr {
	return &air.Constant{Value: e.Value}
}

// Lower a set of zero or more MIR expressions.
func lowerExprs(exprs []Expr, schema *air.Schema) []air.Expr {
	n := len(exprs)
	nexprs := make([]air.Expr, n)

	for i := 0; i < n; i++ {
		nexprs[i] = exprs[i].LowerTo(schema)
	}

	return nexprs
}
