package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

// Column captures the essence of a column at the AIR level.  In reality, all
// columns at this level are reall just untyped data columns.  However, the
// notion of a computed column exists for the purposes of trace expansion.
type Column interface {
	table.Column
	// IsAir is a marker intended to signal that this a column at the lowest level.
	IsAir() bool
}

// DataColumn represents a column of user-provided values.
type DataColumn struct {
	name string
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn(name string) *DataColumn {
	return &DataColumn{name}
}

// IsAir is a marker that indicates this is an AIR column.
func (c *DataColumn) IsAir() bool {
	return true
}

// Name returns the name of this column.
func (c *DataColumn) Name() string {
	return c.name
}

// Computable determines whether or not this column can be computed from the
// existing columns of a trace.  That is, whether or not there is a known
// expression which determines the values for this column based on others in the
// trace.  Data columns are not computable.
func (c *DataColumn) Computable() bool {
	return false
}

// Get the value of this column at a given row in a given trace.
func (c *DataColumn) Get(row int, tr table.Trace) (*fr.Element, error) {
	return tr.GetByName(c.name, row)
}

// Accepts determines whether or not this column accepts the given trace.  At
// this AIR level, this doesn't make sense because data columns are
// untyped.  Thus, this always returns true (i.e. nil).
func (c *DataColumn) Accepts(tr table.Trace) error {
	return nil
}

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn struct {
	name string
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	expr Expr
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn(name string, expr Expr) *ComputedColumn {
	return &ComputedColumn{
		name: name,
		expr: expr,
	}
}

// IsAir is a marker that indicates this is an AIR column.
func (c *ComputedColumn) IsAir() bool {
	return true
}

// Name reads out the name of this column.
func (c *ComputedColumn) Name() string {
	return c.name
}

// Computable determines whether or not this column can be computed from the
// existing columns of a trace.  That is, whether or not there is a known
// expression which determines the values for this column based on others in the
// trace.  Computed columns are, of course, computable.
func (c *ComputedColumn) Computable() bool {
	return true
}

// Get reads the value at a given row in a data column. This amounts to
// looking up that value in the array of values which backs it.
func (c *ComputedColumn) Get(row int, tr table.Trace) (*fr.Element, error) {
	// Compute value at given row
	return c.expr.EvalAt(row, tr), nil
}

// Accepts determines whether or not this column accepts the given trace.  At
// this AIR level, this doesn't make sense because all columns are
// untyped.  Thus, this always returns true (i.e. nil).
func (c *ComputedColumn) Accepts(tr table.Trace) error {
	// NOTE: there are two ways to think about this.  On the one hand, we could
	// check that the given trace has the correct computed value for all rows of
	// this column and, if not, reject it.  However, on the other hand, the
	// prover cannot reject a trace based on the fact that one of its values
	// doesn't matched an expected computed result.  Rather, the prove can only
	// reject a trace for which a given constraint does not hold.
	return nil
}
