package trace

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Describes a set of named data columns.  Columns are not
// required to have the same height.
type Trace interface {
	// Get the value of a given column by its name.  If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	//
	// NOTE: this operation is expected to be slower than
	// GetByindex as, depending on the underlying data format,
	// this may first resolve the name into a physical column
	// index.
	GetByName(name string, row int) (*fr.Element,error)
	// Get the value of a given column by its index. If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	GetByIndex(col int, row int)  (*fr.Element,error)
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() int
}
