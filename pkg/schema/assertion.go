package schema

import (
	"errors"
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
)

// PropertyAssertion is similar to a vanishing constraint but is used only for
// debugging / testing / verification.  Unlike vanishing constraints, property
// assertions do not represent something that the prover can enforce.  Rather,
// they represent properties which are expected to hold for every valid trace.
// That is, they should be implied by the actual constraints.  Thus, whilst the
// prover cannot enforce such properties, external tools (such as for formal
// verification) can attempt to ensure they do indeed always hold.
type PropertyAssertion[T Testable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// Enclosing module for this assertion.  This restricts the asserted
	// property to access only columns from within this module.
	module uint
	// The actual assertion itself, namely an expression which
	// should hold (i.e. vanish) for every row of a trace.
	// Observe that this can be any function which is computable
	// on a given trace --- we are not restricted to expressions
	// which can be arithmetised.
	Property T
}

// NewPropertyAssertion constructs a new property assertion!
func NewPropertyAssertion[T Testable](handle string, property T) *PropertyAssertion[T] {
	// FIXME: determine correct module index
	return &PropertyAssertion[T]{handle, 0, property}
}

// GetHandle returns the handle associated with this constraint.
//
//nolint:revive
func (p *PropertyAssertion[T]) GetHandle() string {
	return p.Handle
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table. If so, return nil otherwise return an error.
//
//nolint:revive
func (p *PropertyAssertion[T]) Accepts(tr tr.Trace) error {
	// Determine height of enclosing module
	height := tr.Modules().Get(p.module).Height()
	// Iterate every row in the module
	for k := uint(0); k < height; k++ {
		// Check whether property holds (or was undefined)
		if !p.Property.TestAt(int(k), tr) {
			// Construct useful error message
			msg := fmt.Sprintf("property assertion %s does not hold (row %d)", p.Handle, k)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}
