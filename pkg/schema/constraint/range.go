package constraint

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
)

// RangeFailure provides structural information about a failing range constraint.
type RangeFailure struct {
	msg string
}

// Message provides a suitable error message
func (p *RangeFailure) Message() string {
	return p.msg
}

func (p *RangeFailure) String() string {
	return p.msg
}

// RangeConstraint restricts all values in a given column to be within
// a range [0..n) for some bound n.  For example, a bound of 256 would
// restrict all values to be bytes.  At this time, range constraints
// are explicitly limited at the arithmetic level to bounds of at most
// 256 (i.e. to ensuring bytes).  This restriction is somewhat
// arbitrary and is determined by the underlying prover.
type RangeConstraint struct {
	// Column index to be constrained.
	column uint
	// The actual constraint itself, namely an expression which
	// should evaluate to zero.  NOTE: an fr.Element is used here
	// to store the bound simply to make the necessary comparison
	// against table data more direct.
	bound *fr.Element
}

// NewRangeConstraint constructs a new Range constraint!
func NewRangeConstraint(column uint, bound *fr.Element) *RangeConstraint {
	var n fr.Element = fr.NewElement(256)
	if bound.Cmp(&n) > 0 {
		panic("Range constraint for bitwidth above 8 not supported")
	}

	return &RangeConstraint{column, bound}
}

// Accepts checks whether a range constraint evaluates to zero on
// every row of a table. If so, return nil otherwise return an error.
func (p *RangeConstraint) Accepts(tr trace.Trace) schema.Failure {
	column := tr.Column(p.column)
	height := tr.Height(column.Context())
	// Iterate all rows of the module
	for k := 0; k < int(height); k++ {
		// Get the value on the kth row
		kth := column.Get(k)
		// Perform the bounds check
		if kth.Cmp(p.bound) >= 0 {
			name := column.Name()
			// Construct useful error message
			msg := fmt.Sprintf("value out-of-bounds (row %d, %s)", kth, name)
			// Evaluation failure
			return &RangeFailure{msg}
		}
	}
	// All good
	return nil
}

func (p *RangeConstraint) String() string {
	return fmt.Sprintf("(range #%d %s)", p.column, p.bound)
}

// Column returns the index of the column subjected to the constraint.
func (p *RangeConstraint) Column() uint {
	return p.column
}

// Bound returns the range boundary of the constraint.
//
// Note: the bound is returned in the form of a uint because this is simpler
// and more straightforward to understand.
func (p *RangeConstraint) Bound() uint64 {
	return p.bound.Uint64()
}
