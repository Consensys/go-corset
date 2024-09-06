package trace

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace interface {
	// Access a given column in this trace.
	Column(uint) Column
	// Returns the number of columns in this trace.
	Width() uint
	// Returns the height of the given context (i.e. module).
	Height(Context) uint
	// Module returns the list of assigned modules and their respective heights
	Modules() util.Iterator[ArrayModule]
}

// Column describes an individual column of data within a trace table.
type Column interface {
	// Evaluation context of this column
	Context() Context
	// Holds the name of this column
	Name() string
	// Get the value at a given row in this column.  If the row is
	// out-of-bounds, then the column's padding value is returned instead.
	// Thus, this function always succeeds.
	Get(row int) fr.Element
	// Access the underlying data array for this column.  This is useful in
	// situations where we want to clone the entire column, etc.
	Data() util.FrArray
	// Value to be used when padding this column
	Padding() fr.Element
}

// RawColumn represents a raw column of data which has not (yet) been indexed as
// part of a trace, etc.  Raw columns are typically read directly from trace
// files, and subsequently indexed into a trace during the expansion process.
type RawColumn struct {
	// Name of the enclosing module
	Module string
	// Name of the column
	Name string
	// Data held in the column
	Data util.FrArray
}

// QualifiedName returns the fully qualified name of this column.
func (p *RawColumn) QualifiedName() string {
	return QualifiedColumnName(p.Module, p.Name)
}

// CellRef identifies a unique cell within a given table.
type CellRef struct {
	// Column index for the cell
	Column uint
	// Row index for the cell
	Row int
}
