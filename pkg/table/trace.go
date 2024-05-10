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

// Testable captures the notion of a constraint which can be tested on a given
// row of a given trace.  It is very similar to Evaluable, except that it only
// indicates success or failure.  The reason for using this interface over
// Evaluable is that, for historical reasons, constraints at the HIR cannot be
// Evaluable (i.e. because they return multiple values, rather than a single
// value).  However, constraints at the HIR level remain testable.
type Testable interface {
	// TestAt evaluates this expression in a given tabular context and checks it
	// against zero. Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be undefined for
	// several reasons: firstly, if it accesses a row which does not exist (e.g.
	// at index -1); secondly, if it accesses a column which does not exist.
	TestAt(int, Trace) bool
}

// Acceptable represents an element which can "accept" a trace, or either reject
// with an error or report a warning.
type Acceptable interface {
	Accepts(Trace) error
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
	// ColumnByName returns the data of a given column in order that it can be
	// inspected.  If the given column does not exist, then nil is returned.
	ColumnByName(name string) []*fr.Element
	// Add a new column of data
	AddColumn(name string, data []*fr.Element)
}

// ForallAcceptTrace determines whether or not one or more groups of constraints
// accept a given trace.  It returns the first error or warning encountered.
func ForallAcceptTrace[T Acceptable](trace Trace, constraints []T) error {
	for _, c := range constraints {
		err := c.Accepts(trace)
		if err != nil {
			return err
		}
	}
	//
	return nil
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
	columns []*ArrayTraceColumn
}

// EmptyArrayTrace constructs an empty array trace into which column data can be
// added.
func EmptyArrayTrace() *ArrayTrace {
	p := new(ArrayTrace)
	// Initially empty columns
	p.columns = make([]*ArrayTraceColumn, 0)
	// Initialise height as 0
	p.height = 0
	// done
	return p
}

// Clone creates an identical clone of this trace.
func (p *ArrayTrace) Clone() *ArrayTrace {
	clone := new(ArrayTrace)
	clone.columns = make([]*ArrayTraceColumn, len(p.columns))
	clone.height = p.height
	//
	for i, c := range p.columns {
		// TODO: can this be avoided?
		clone.columns[i] = c.Clone()
	}
	// done
	return clone
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
	p.columns = append(p.columns, &column)
	// Update maximum height
	if len(data) > p.height {
		p.height = len(data)
	}
}

// Columns returns the set of columns in this trace.
func (p *ArrayTrace) Columns() []*ArrayTraceColumn {
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
	c := p.getColumnByName(name)
	if c != nil {
		// Matched column
		return c.Get(row)
	}

	// Failed to find column
	msg := fmt.Sprintf("Invalid column: {%s}", name)

	return nil, errors.New(msg)
}

// ColumnByName looks up a column based on its name.  If the column doesn't
// exist, then nil is returned.
func (p *ArrayTrace) ColumnByName(name string) []*fr.Element {
	for _, c := range p.columns {
		if name == c.name {
			// Matched column
			return c.data
		}
	}

	return nil
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

func (p *ArrayTrace) getColumnByName(name string) *ArrayTraceColumn {
	for _, c := range p.columns {
		if name == c.name {
			// Matched column
			return c
		}
	}

	return nil
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

// Clone an ArrayTraceColumn
func (p *ArrayTraceColumn) Clone() *ArrayTraceColumn {
	clone := new(ArrayTraceColumn)
	clone.name = p.name
	clone.data = make([]*fr.Element, len(p.data))
	copy(clone.data, p.data)

	return clone
}

// Get the value at the given row of this column.
func (p *ArrayTraceColumn) Get(row int) (*fr.Element, error) {
	if row < 0 || row >= len(p.data) {
		return nil, errors.New("Column access out-of-bounds")
	}

	return p.data[row], nil
}
