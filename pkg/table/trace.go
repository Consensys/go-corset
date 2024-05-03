package table

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Evaluable captures something which can be evaluated on a given table row to
// produce an evaluation point.  For example, expressions in the
// Mid-Level or Arithmetic-Level IR can all be evaluated at rows of a
// table.
type Evaluable interface {
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, Trace) *fr.Element
}

// Acceptor represents an element which can "accept" a trace, or
// either reject with an error or report a warning.
type Acceptor interface {
	Accepts(Trace) (bool, error)
}

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace interface {
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() int
	// Get the value of a given column by its name.  If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	//
	// NOTE: this operation is expected to be slower than
	// GetByindex as, depending on the underlying data format,
	// this may first resolve the name into a physical column
	// index.
	GetByName(name string, row int) (*fr.Element, error)
	// Get the value of a given column by its index. If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	GetByIndex(col int, row int) (*fr.Element, error)
	// Check whether this trace contains data for the given column.
	HasColumn(name string) bool
	// Add a new column of data
	AddColumn(name string, data []*fr.Element)
}

// ForallAcceptTrace determines whether or not one or more groups of constraints
// accept a given trace.  It returns the first error or warning encountered.
func ForallAcceptTrace[T Acceptor](trace Trace, constraints []T) (bool, error) {
	for _, c := range constraints {
		warning, err := c.Accepts(trace)
		if err != nil {
			return warning, err
		}
	}
	//
	return false, nil
}

// ===================================================================
// Array Trace
// ===================================================================

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace struct {
	// Holds the maximum height of any column in the trace
	height int
	// Holds the name of each column
	columns []ArrayTraceColumn
}

// EmptyArrayTrace constructs an empty array trace into which column data can be
// added.
func EmptyArrayTrace() *ArrayTrace {
	p := new(ArrayTrace)
	// Initially empty columns
	p.columns = make([]ArrayTraceColumn, 0)
	// Initialise height as 0
	p.height = 0
	// done
	return p
}

// HasColumn checks whether the trace has a given column or not.
func (p *ArrayTrace) HasColumn(name string) bool {
	for _, c := range p.columns {
		if c.name == name {
			return true
		}
	}

	return false
}

// AddColumn adds a new column of data to this trace.
func (p *ArrayTrace) AddColumn(name string, data []*fr.Element) {
	// Sanity check the column does not already exist.
	if p.HasColumn(name) {
		panic("column already exists")
	}
	// Construct new column
	column := ArrayTraceColumn{name, data}
	// Append it
	p.columns = append(p.columns, column)
	// Update maximum height
	if len(data) > p.height {
		p.height = len(data)
	}
}

// Columns returns the set of columns in this trace.
func (p *ArrayTrace) Columns() []ArrayTraceColumn {
	return p.columns
}

// Height determines the maximum height of any column within this trace.
func (p *ArrayTrace) Height() int {
	return p.height
}

// GetByName gets the value of a given column (as identified by its name) at a
// given row.  If the column does not exist, an error is returned.
func (p *ArrayTrace) GetByName(name string, row int) (*fr.Element, error) {
	// NOTE: Could improve performance here if names were kept in
	// sorted order.
	for _, c := range p.columns {
		if name == c.name {
			// Matched column
			return c.Get(row)
		}
	}
	// Failed to find column
	msg := fmt.Sprintf("Invalid column: {%s}", name)

	return nil, errors.New(msg)
}

// GetByIndex returns the value of a given column (as identifier by its index or
// register) at a given row.  If the column is out-of-bounds an error is
// returned.
func (p *ArrayTrace) GetByIndex(col int, row int) (*fr.Element, error) {
	if col < 0 || col >= len(p.columns) {
		return nil, errors.New("Column access out-of-bounds")
	}

	return p.columns[col].Get(row)
}

// ===================================================================
// Array Trace Column
// ===================================================================

// ArrayTraceColumn represents a column of data within an array trace.
type ArrayTraceColumn struct {
	// Holds the name of this column
	name string
	// Holds the raw data making up this column
	data []*fr.Element
}

// Name returns the name of the given column.
func (p *ArrayTraceColumn) Name() string {
	return p.name
}

// Get the value at the given row of this column.
func (p *ArrayTraceColumn) Get(row int) (*fr.Element, error) {
	if row < 0 || row >= len(p.data) {
		return nil, errors.New("Column access out-of-bounds")
	}

	return p.data[row], nil
}
