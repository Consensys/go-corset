package mir

import (
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
)

// Schema for MIR constraints and columns.
type Schema = table.Schema[Column, Constraint]

// Constraint identifies the essence of a constraint at the MIR level.  For now,
// all constraints are vanishing constraints.
type Constraint = *table.VanishingConstraint[Expr]

// LowerToAir lowers (or refines) an MIR table into an AIR table.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func LowerToAir(mirSchema *Schema, airSchema *air.Schema) {
	for _, col := range mirSchema.Columns() {
		dc := col.(*DataColumn)
		dc.LowerTo(airSchema)
	}

	for _, c := range mirSchema.Constraints() {
		// FIXME: this is broken because its currently
		// assuming that an AirConstraint is always a
		// VanishingConstraint.  Eventually this will not be
		// true.
		air_expr := c.Expr.LowerTo(airSchema)
		airSchema.AddVanishingConstraint(c.Handle, c.Domain, air_expr)
	}
}
