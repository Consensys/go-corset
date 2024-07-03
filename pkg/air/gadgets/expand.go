package gadgets

import "github.com/consensys/go-corset/pkg/air"

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
	// No optimisation, therefore expand the column
	panic("todo")
}
