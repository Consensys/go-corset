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

func NewDataColumn(name string, data []*big.Int) DataColumn {
	var c DataColumn
	c.name = name
	c.data = data
	return c
}

// Read out the name of this column
func (c DataColumn) Name() string {
	return c.name
}

func (c DataColumn) Height() int {
	return len(c.data)
}

// Read the value at a given row in a data column.  This amounts to
// looking up that value in the array of values which backs it.
func (c DataColumn) Get(row int) (*big.Int,error) {
	if row < 0 || row >= len(c.data) {
		return nil,errors.New("data column access out-of-bounds")
	} else {
		return c.data[row],nil
	}
}
