package trace

import (
	"errors"
	"math/big"
)

// Describes a given column and provides a mechanism for accessing its
// values at a given row.
type Column interface {
	// Get the name of this column.
	Name() string
	// Get the value at a given row in this column, or return an
	// error.
	Get(row int) (*big.Int,error)
	// Get the number of rows in this column
	Height() int
}

// ===================================================================
// Data Column
// ===================================================================

// Describes a column which is backed by an array of data values.
// Such columns are fundamental and must be provided as part of the
// trace.  Despite this, such columns can still be manipulated in
// certain ways, such as by introducing padding to ensure they have a
// given length, etc.
type DataColumn struct {
	name string
	data []*big.Int
}

func NewDataColumn(name string, data []*big.Int) *DataColumn {
	var c DataColumn
	c.name = name
	c.data = data
	return &c
}

// Read out the name of this column
func (c *DataColumn) Name() string {
	return c.name
}

func (c *DataColumn) Height() int {
	return len(c.data)
}

// Read the value at a given row in a data column.  This amounts to
// looking up that value in the array of values which backs it.
func (c *DataColumn) Get(row int) (*big.Int,error) {
	if row < 0 || row >= len(c.data) {
		return nil,errors.New("data column access out-of-bounds")
	} else {
		return c.data[row],nil
	}
}

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
	// The pre-determined height of a computed column.  This is
	// typically derived from the height of those columns it
	// depends upon.  However, compute columns can also have fixed
	// heights, etc.
	height int
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	fn func(int) *big.Int
}

func NewComputedColumn(name string, height int, fn func(int) *big.Int) *ComputedColumn {
	var c ComputedColumn
	c.name = name
	c.height = height
	c.fn = fn
	return &c
}

// Read out the name of this column
func (c *ComputedColumn) Name() string {
	return c.name
}

func (c *ComputedColumn) Height() int {
	return c.height
}

// Read the value at a given row in a data column.  This amounts to
// looking up that value in the array of values which backs it.
func (c *ComputedColumn) Get(row int) (*big.Int,error) {
	if row < 0 || row >= c.height {
		return nil,errors.New("data column access out-of-bounds")
	} else {
		return c.fn(row),nil
	}
}
