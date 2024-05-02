package table

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Column describes a given column and provides a mechanism for accessing its
// values at a given row.
type Column interface {
	// Get the name of this column.
	Name() string
	// Get the value at a given row in this column, or return an
	// error.
	Get(row int, tr Trace) (*fr.Element, error)
	// Determine whether this is a computed column or not
	Computable() bool
	// Check whether or not this column accepts a particular
	// trace.  A column might reject a trace if the values in that
	// trace do not meet some specific requirement (e.g. they are
	// all bytes).
	Accepts(Trace) error
}

// Schema describes the permitted "layout" of a given trace.  That includes
// identifying the required columns and the set of constraints which must hold
// over the trace.  Columns can be either data columns, or computed columns.  A
// data column is one whose values are expected to be provided by the user,
// whilst computed columns are derivatives whose values can be computed from the
// other columns of the trace. A trace of data values is said to be "accepted"
// by a schema if: (1) every data column in the schema exists in the trace; (2)
// every constraint in the schema holds for the trace.
type Schema[C Column, R Constraint] struct {
	// Column array (either data or computed).  Columns are stored
	// such that the dependencies of a column always come before
	// that column (i.e. have a lower index).  Thus, data columns
	// always precede computed columns, etc.
	columns []C
	// Constaint array.  For a trace of values to be well-formed
	// with respect to this schema, each constraint must hold.
	constraints []R
	// Property assertions.
	assertions []Assertion
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema[C Column, R Constraint]() *Schema[C, R] {
	p := new(Schema[C, R])
	// Initially empty columns
	p.columns = make([]C, 0)
	// Initially empty constraints
	p.constraints = make([]R, 0)
	// Initialise empty assertions
	p.assertions = make([]Assertion, 0)
	// Done
	return p
}

// NewSchema constructs a new Schema initialised with a given set of columns and
// constraints.
func NewSchema[C Column, R Constraint](columns []C, constraints []R) *Schema[C, R] {
	p := new(Schema[C, R])
	p.columns = columns
	p.constraints = constraints
	//
	return p
}

// AcceptsTrace determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
func (p *Schema[C, R]) AcceptsTrace(trace Trace) bool {
	// TODO: check that required columns are present.
	// TODO: check that each column accepts its data.
	for _, c := range p.Columns() {
		err := c.Accepts(trace)
		if err != nil {
			return false
		}
	}

	for _, c := range p.Constraints() {
		err := c.Accepts(trace)
		if err != nil {
			return false
		}
	}

	return true
}

// HasColumn checks whether a given schema has a given column.
func (p *Schema[C, R]) HasColumn(name string) bool {
	for _, c := range p.columns {
		if c.Name() == name {
			return true
		}
	}

	return false
}

// Columns returns the set of columns (data or computed) which are required by
// this schema.
func (p *Schema[C, R]) Columns() []C {
	return p.columns
}

// Constraints returns the set of constraints required by this schema.
func (p *Schema[C, R]) Constraints() []R {
	return p.constraints
}

// AddConstraint appends a new constraint onto the schema.
func (p *Schema[C, R]) AddConstraint(constraint R) {
	p.constraints = append(p.constraints, constraint)
}

// AddColumn appends a new column onto the schema.
func (p *Schema[C, R]) AddColumn(column C) {
	// TODO: check the column does not already exist?
	p.columns = append(p.columns, column)
}

// ExpandTrace expands a given trace according to this schema.  More
// specifically, that means computing the actual values for any computed
// columns. Observe that computed columns have to be computed in the correct
// order.
func (p *Schema[C, R]) ExpandTrace(tr Trace) {
	for _, c := range p.columns {
		if c.Computable() && !tr.HasColumn(c.Name()) {
			data := make([]*fr.Element, tr.Height())
			// Expand the trace
			for i := 0; i < len(data); i++ {
				var err error
				// NOTE: at the moment Get cannot return an error anyway
				data[i], err = c.Get(i, tr)
				// FIXME: we need proper error handling
				if err != nil {
					panic(err)
				}
			}
			// Colunm needs to be expanded.
			tr.AddColumn(c.Name(), data)
		}
	}
}
