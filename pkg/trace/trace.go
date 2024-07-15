package trace

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace interface {
	// Access the columns of this trace.
	Columns() ColumnSet
	// Clone creates an identical clone of this trace.
	Clone() Trace
	// Access the modules of this trace.
	Modules() ModuleSet
}

// ColumnSet provides an interface to the declared columns within this trace.
type ColumnSet interface {
	// Add a new column to this column set.
	Add(column Column) uint
	// Get the ith module in this set.
	Get(uint) Column
	// Determine index of given column, or return false if this fails.
	IndexOf(module uint, column string) (uint, bool)
	// Returns the number of items in this array.
	Len() uint
	// Swap two columns in this column set
	Swap(uint, uint)
	// Reduce the number of columns to a given length by removing columns from
	// the end.
	Trim(uint)
}

// Column describes an individual column of data within a trace table.
type Column interface {
	// Clone this column
	Clone() Column
	// Get the value at a given row in this column.  If the row is
	// out-of-bounds, then the column's padding value is returned instead.
	// Thus, this function always succeeds.
	Get(row int) *fr.Element
	// Return the height (i.e. number of rows) of this column.
	Height() uint
	// Returns the evaluation context for this column.  That identifies the
	// enclosing module, and then length multiplier (which must be a factor of
	// the height). For example, if the multiplier is 2 then the height must
	// always be a multiple of 2, etc.  This affects padding also, as we must
	// pad to this multiplier, etc.
	Context() Context
	// Get the name of this column
	Name() string
	// Return the value to use for padding this column.
	Padding() *fr.Element
	// Pad this column with n copies of the column's padding value.
	Pad(n uint)
	// Reseat updates the module index of this column (e.g. as a result of a
	// realignment).
	Reseat(mid uint)
	// Return the width (i.e. number of bytes per element) of this column.
	Width() uint
	// Write the raw bytes of this column to a given writer, returning an error
	// if this failed (for some reason).
	Write(io.Writer) error
}

// ModuleSet provides an interface to the declared moules within this trace.
type ModuleSet interface {
	// Register a new module with a given name and height, returning the module
	// index.
	Add(string, uint) uint
	// Get the ith module in this set.
	Get(uint) *Module
	// Determine index of given module, or return false if this fails.
	IndexOf(string) (uint, bool)
	// Returns the number of items in this array.
	Len() uint
	// Pad the ith module in this set with n items at the front of each column
	Pad(mid uint, n uint)
	// Swap order of modules.  Note columns are updated accordingly.
	Swap(uint, uint)
}

// Module describes an individual module within the trace table, and
// permits actions on the columns of this module as a whole.
type Module struct {
	// Name of this module.
	name string
	// Determine the columns contained in this module by their column index.
	columns []uint
	// Height (in rows) of this module.  Specifically, every column in this
	// module must have this height.
	height uint
}

// Name of this module.
func (p *Module) Name() string {
	return p.name
}

// Columns identifies the columns contained in this module by their column index.
func (p *Module) Columns() []uint {
	return p.columns
}

// Copy creates a copy of this module, such that mutations to the copy will not
// affect the original.
func (p *Module) Copy() Module {
	var clone Module
	clone.name = p.name
	clone.height = p.height
	clone.columns = make([]uint, len(p.columns))
	// Copy column indices
	copy(clone.columns, p.columns)
	// Done
	return clone
}

// Height (in rows) of this module.  Specifically, every column in this
// module must have this height.
func (p *Module) Height() uint {
	return p.height
}

// Register a new columnd contained within this module.
func (p *Module) registerColumn(cid uint) {
	p.columns = append(p.columns, cid)
}
