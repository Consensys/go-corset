package table

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// DataColumn represents a column of user-provided values.
type DataColumn struct {
	name string
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn(name string) *DataColumn {
	return &DataColumn{name}
}

// Name returns the name of this column.
func (c *DataColumn) Name() string {
	return c.name
}

// Get the value of this column at a given row in a given trace.
func (c *DataColumn) Get(row int, tr Trace) (*fr.Element, error) {
	return tr.GetByName(c.name, row)
}

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn[E Evaluable] struct {
	name string
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	expr E
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn[E Evaluable](name string, expr E) *ComputedColumn[E] {
	return &ComputedColumn[E]{
		name: name,
		expr: expr,
	}
}

// Name reads out the name of this column.
func (c *ComputedColumn[E]) Name() string {
	return c.name
}

// Get reads the value at a given row in a data column. This amounts to
// looking up that value in the array of values which backs it.
func (c *ComputedColumn[E]) Get(row int, tr Trace) (*fr.Element, error) {
	// Compute value at given row
	return c.expr.EvalAt(row, tr), nil
}
