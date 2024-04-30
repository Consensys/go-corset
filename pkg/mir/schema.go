package mir

import (
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
)

type Schema = table.Schema[Column,Constraint]

// ===================================================================
// Constraints
// ===================================================================

// For now, all constraints are vanishing constraints.
type Constraint = *table.VanishingConstraint[Expr]

// Lower (or refine) an MIR table into an AIR table.  That means
// lowering all the columns and constraints, whilst adding additional
// columns / constraints as necessary to preserve the original
// semantics.
func LowerToAir(mirSchema *Schema, airSchema *air.Schema) {
	for _,col := range mirSchema.Columns() {
		airSchema.AddColumn(col.LowerTo(airSchema))
	}
	for _,c := range mirSchema.Constraints() {
		// FIXME: this is broken because its currently
		// assuming that an AirConstraint is always a
		// VanishingConstraint.  Eventually this will not be
		// true.
		air_expr := c.Expr.LowerTo(airSchema)
		airSchema.AddConstraint(&air.VanishingConstraint{Handle: c.Handle,Expr: air_expr})
	}
}