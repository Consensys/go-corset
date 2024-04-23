package hir

import (
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/trace"
)

// ============================================================================
// Table
// ============================================================================

type HirTable = trace.Table[Column,Constraint]

// Lower (or refine) an HIR table into an MIR table.  That means
// lowering all the columns and constraints, whilst adding additional
// columns / constraints as necessary to preserve the original
// semantics.
func LowerToMir(hirTbl HirTable, mirTbl mir.Table) {
	// First, lower columns
	for _,col := range hirTbl.Columns() {
		mirTbl.AddColumn(col.LowerTo())
	}
	// Second, lower constraints
	for _,c := range hirTbl.Constraints() {
		mir_exprs := c.Expr.LowerTo()
		// Add individual constraints arising
		for _,mir_expr := range mir_exprs {
			mirTbl.AddConstraint(&mir.VanishingConstraint{Handle: c.Handle,Expr: mir_expr})
		}
	}
}
