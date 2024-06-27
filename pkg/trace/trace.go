package trace

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
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
	// // Get the module index of the module which contains this column.
	// Module() uint
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
	// Access the columns of this trace.
	Columns() util.Array[Column]
	// Clone creates an identical clone of this trace.
	Clone() Trace
	// Determine the index of a particular column in this trace, or return false
	// if no such column exists.
	ColumnIndex(name string) (uint, bool)
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() uint
}
