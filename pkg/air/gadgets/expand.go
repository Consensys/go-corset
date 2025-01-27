package gadgets

import (
	"github.com/consensys/go-corset/pkg/air"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Expand converts an arbitrary expression into a specific column index.  In
// general, this means adding a computed column to hold the value of the
// arbitrary expression and returning its index.  However, this can be optimised
// in the case the given expression is a direct column access by simply
// returning the accessed column index.
func Expand(ctx trace.Context, bitwidth uint, e air.Expr, schema *air.Schema) uint {
	if ctx.IsVoid() || ctx.IsConflicted() {
		panic("conflicting (or void) context")
	}
	//
	if ca, ok := e.(*air.ColumnAccess); ok && ca.Shift == 0 {
		// Optimisation possible
		return ca.Column
	}
	// Determine computed column name
	name := e.Lisp(schema).String(false)
	// Look up column
	index, ok := sc.ColumnIndexOf(schema, ctx.Module(), name)
	// Add new column (if it does not already exist)
	if !ok {
		// Add computed column
		index = schema.AddAssignment(assignment.NewComputedColumn[air.Expr](ctx, name, sc.NewUintType(bitwidth), e))
		// Construct v == [e]
		v := air.NewColumnAccess(index, 0)
		// Construct 1 == e/e
		eq_e_v := v.Equate(e)
		// Ensure (e - v) == 0, where v is value of computed column.
		schema.AddVanishingConstraint(name, ctx, util.None[int](), eq_e_v)
	}
	//
	return index
}
