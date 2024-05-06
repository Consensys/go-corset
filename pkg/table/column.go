package table

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// DataColumn represents a column of user-provided values.
type DataColumn[T Type] struct {
	Name string
	Type T
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn[T Type](name string, base T) *DataColumn[T] {
	return &DataColumn[T]{name, base}
}

// Get the value of this column at a given row in a given trace.
func (c *DataColumn[T]) Get(row int, tr Trace) (*fr.Element, error) {
	return tr.GetByName(c.Name, row)
}

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
//
//nolint:revive
func (c *DataColumn[T]) Accepts(tr Trace) error {
	for i := 0; i < tr.Height(); i++ {
		val, err := tr.GetByName(c.Name, i)
		if err != nil {
			return err
		}

		if !c.Type.Accept(val) {
			// Construct useful error message
			msg := fmt.Sprintf("column %s value out-of-bounds (row %d, %s)", c.Name, i, val)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn struct {
	Name string
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	Expr Evaluable
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn(name string, expr Evaluable) *ComputedColumn {
	return &ComputedColumn{
		Name: name,
		Expr: expr,
	}
}

// Get reads the value at a given row in a data column. This amounts to
// looking up that value in the array of values which backs it.
func (c *ComputedColumn) Get(row int, tr Trace) (*fr.Element, error) {
	// Compute value at given row
	return c.Expr.EvalAt(row, tr), nil
}
