package trace

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Column describes an individual column of data within a trace table.
type Column interface {
	// Clone creates an identical clone of this column.
	Clone() Column
	// Return the raw data stored in this column.
	Data() []*fr.Element
	// Get the value at a given row in this column.  If the row is
	// out-of-bounds, then the column's padding value is returned instead.
	// Thus, this function always succeeds.
	Get(row int) *fr.Element
	// Return the height (i.e. number of rows) of this column.
	Height() uint
	// Get the name of this column
	Name() string
	// Pad this column n items at the front.
	Pad(n uint)
	// Return the value to use for padding this column.
	Padding() *fr.Element
	// Return the width (i.e. number of bytes per element) of this column.
	Width() uint
	// Write the raw bytes of this column to a given writer, returning an error
	// if this failed (for some reason).
	Write(io.Writer) error
}

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace interface {
	// Add a new column of data
	Add(Column)
	// Clone creates an identical clone of this trace.
	Clone() Trace
	// Column returns the ith column in this trace.
	Column(uint) Column
	// Determine the index of a particular column in this trace, or return false
	// if no such column exists.
	ColumnIndex(name string) (uint, bool)
	// Check whether this trace contains data for the given column.
	HasColumn(name string) bool
	// Pad each column in this trace with n items at the front.
	Pad(n uint)
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() uint
	// Swap the order of two columns in this trace.  This is needed, in
	// particular, for alignment.
	Swap(uint, uint)
	// Get the number of columns in this trace.
	Width() uint
}
