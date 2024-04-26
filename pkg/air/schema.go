package air

import (
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

type Schema = table.Schema[Column,Constraint]

type Column = table.Column

// ===================================================================
// Computed Column
// ===================================================================

// Describes a column whose values are computed on-demand, rather than
// being stored in a backing array.  Typically computed columns read
// values from other columns in a trace in order to calculate their
// value.  There is an expectation that this computation is not
// cyclic.
type ComputedColumn struct {
	name string
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	expr Expr
}

func NewComputedColumn(name string, expr Expr) *ComputedColumn {
	var c ComputedColumn
	c.name = name
	c.expr = expr
	return &c
}

// Read out the name of this column
func (c *ComputedColumn) Name() string {
	return c.name
}

func (c *ComputedColumn) MinHeight() int {
	return 0
}

func (c *ComputedColumn) Computable() bool {
	return true
}

// Read the value at a given row in a data column.  This amounts to
// looking up that value in the array of values which backs it.
func (c *ComputedColumn) Get(row int, tr table.Trace) (*fr.Element,error) {
	// Compute value at given row
	return c.expr.EvalAt(row,tr), nil
}
