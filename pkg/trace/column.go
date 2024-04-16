package trace

import (
	"errors"
	"math/big"
)

// Column describes a given column and provides a mechanism for accessing its
// values at a given row.
type Column interface {
	// Name the name of this column.
	Name() string
	// Get the value at a given row in this column, or return an
	// error.
	Get(row int) (*big.Int, error)
	// Height is the number of rows in this column.
	Height() int
}

// DataColumn describes a column which is backed by an array of data values.
// Such columns are fundamental and must be provided as part of the
// trace.  Despite this, such columns can still be manipulated in
// certain ways, such as by introducing padding to ensure they have a
// given length, etc.
type DataColumn struct {
	name string
	data []*big.Int
}

// NOTE: This is used for compile time type checking if the given type satisfies the given interface.
var _ Column = (*DataColumn)(nil)

// NewDataColumn constructs a new instance of DataColumn.
func NewDataColumn(name string, data []*big.Int) *DataColumn {
	return &DataColumn{
		name: name,
		data: data,
	}
}

// Name reads out the name of this column.
func (c *DataColumn) Name() string {
	return c.name
}

// Height returns the height of the DataColumn.
func (c *DataColumn) Height() int {
	return len(c.data)
}

// Get reads the value at a given row in a data column.  This amounts to
// looking up that value in the array of values which backs it.
func (c *DataColumn) Get(row int) (*big.Int, error) {
	if row < 0 || row >= len(c.data) {
		return nil, errors.New("data column access out-of-bounds")
	}

	return c.data[row], nil
}
