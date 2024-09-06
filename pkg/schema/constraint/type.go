package constraint

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
)

// TypeFailure provides structural information about a failing type constraint.
type TypeFailure struct {
	msg string
}

// Message provides a suitable error message
func (p *TypeFailure) Message() string {
	return p.msg
}

func (p *TypeFailure) String() string {
	return p.msg
}

// TypeConstraint restricts all values in a given column to be within
// a range [0..n) for some bound n.  Any bound is supported, and the system will
// choose the best underlying implementation as needed.
type TypeConstraint struct {
	// Column to be constrained.
	column uint
	// The actual constraint itself, namely an expression which
	// should evaluate to zero.  NOTE: an fr.Element is used here
	// to store the bound simply to make the necessary comparison
	// against table data more direct.
	expected schema.Type
}

// NewTypeConstraint constructs a new Range constraint!
func NewTypeConstraint(column uint, expected schema.Type) *TypeConstraint {
	return &TypeConstraint{column, expected}
}

// Target returns the target column for this constraint.
func (p *TypeConstraint) Target() uint {
	return p.column
}

// Type returns the expected for all values in the target column.
func (p *TypeConstraint) Type() schema.Type {
	return p.expected
}

// Accepts checks whether a range constraint evaluates to zero on
// every row of a table. If so, return nil otherwise return an error.
func (p *TypeConstraint) Accepts(tr trace.Trace) schema.Failure {
	column := tr.Column(p.column)
	// Determine height
	height := tr.Height(column.Context())
	// Iterate every row
	for k := 0; k < int(height); k++ {
		// Get the value on the kth row
		kth := column.Get(k)
		// Perform the type check
		if !p.expected.Accept(kth) {
			name := column.Name()
			// Construct useful error message
			msg := fmt.Sprintf("value out-of-bounds (row %d, %s)", kth, name)
			// Evaluation failure
			return &TypeFailure{msg}
		}
	}
	// All good
	return nil
}

func (p *TypeConstraint) String() string {
	return fmt.Sprintf("(type %d %s)", p.column, p.expected.String())
}
