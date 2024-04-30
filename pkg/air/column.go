package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

type Column interface {
	table.Column
	// IsAir is a marker intended to signal that this a column at the lowest level.
	IsAir() bool
}

// ===================================================================
// Data Column
// ===================================================================

type DataColumn struct {
	name string
}

func NewDataColumn(name string) *DataColumn {
	return &DataColumn{name}
}

func (c *DataColumn) IsAir() bool {
	return true
}

func (c *DataColumn) Name() string {
	return c.name
}

func (c *DataColumn) Computable() bool {
	return false
}

func (c *DataColumn) Get(row int, tr table.Trace) (*fr.Element, error) {
	return tr.GetByName(c.name, row)
}

func (c *DataColumn) Accepts(tr table.Trace) error {
	return nil
}

// ===================================================================
// Computed Column
// ===================================================================

// ComputedColumn describes a column whose values are computed on-demand, rather than
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
	return &ComputedColumn{
		name: name,
		expr: expr,
	}
}

func (c *ComputedColumn) IsAir() bool {
	return true
}

// Name reads out the name of this column.
func (c *ComputedColumn) Name() string {
	return c.name
}

func (c *ComputedColumn) MinHeight() int {
	return 0
}

func (c *ComputedColumn) Computable() bool {
	return true
}

// Get reads the value at a given row in a data column. This amounts to
// looking up that value in the array of values which backs it.
func (c *ComputedColumn) Get(row int, tr table.Trace) (*fr.Element, error) {
	// Compute value at given row
	return c.expr.EvalAt(row, tr), nil
}

func (c *ComputedColumn) Accepts(tr table.Trace) error {
	// FIXME: does this make sense?
	return nil
}
