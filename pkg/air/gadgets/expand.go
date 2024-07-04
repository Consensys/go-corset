package gadgets

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/air"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
)

// Expand converts an arbitrary expression into a specific column index.  In
// general, this means adding a computed column to hold the value of the
// arbitrary expression and returning its index.  However, this can be optimised
// in the case the given expression is a direct column access by simply
// returning the accessed column index.
func Expand(e air.Expr, schema *air.Schema) uint {
	//
	if ca, ok := e.(*air.ColumnAccess); ok && ca.Shift == 0 {
		// Optimisation possible
		return ca.Column
	}
	// No optimisation, therefore expand using a computedcolumn
	ctx := sc.DetermineEnclosingModuleOfExpression(e, schema)
	// Determine computed column name
	name := e.String()
	// Look up column
	index, ok := sc.ColumnIndexOf(schema, ctx.Module, name)
	// Add new column (if it does not already exist)
	if !ok {
		// Add computed column
		index = schema.AddAssignment(assignment.NewComputedColumn(ctx.Module, name, ctx.Multiplier, e))
	}
	// Construct v == [e]
	v := air.NewColumnAccess(index, 0)
	// Construct 1 == e/e
	eq_e_v := v.Equate(e)
	// Ensure (e - v) == 0, where v is value of computed column.
	c_name := fmt.Sprintf("[%s]", e.String())
	schema.AddVanishingConstraint(c_name, ctx.Module, ctx.Multiplier, nil, eq_e_v)
	//
	return index
}
