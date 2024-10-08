package schema

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// AssertionFailure provides structural information about a failing vanishing constraint.
type AssertionFailure struct {
	// Handle of the failing constraint
	handle string
	// Constraint expression
	constraint Testable
	// Row on which the constraint failed
	row uint
}

// Handle returns the handle of this constraint
func (p *AssertionFailure) Handle() string {
	// Construct useful error message
	return p.handle
}

// Message provides a suitable error message
func (p *AssertionFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("assertion \"%s\" does not hold (row %d)", p.handle, p.row)
}

// Constraint returns the constraint expression itself.
func (p *AssertionFailure) Constraint() Testable {
	return p.constraint
}

// Row identifies the row on which this constraint failed.
func (p *AssertionFailure) Row() uint {
	return p.row
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *AssertionFailure) RequiredCells(trace tr.Trace) *util.AnySortedSet[tr.CellRef] {
	return p.constraint.RequiredCells(int(p.row), trace)
}

func (p *AssertionFailure) String() string {
	return p.Message()
}

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
	handle string
	// Enclosing module for this assertion.  This restricts the asserted
	// property to access only columns from within this module.
	context tr.Context
	// The actual assertion itself, namely an expression which
	// should hold (i.e. vanish) for every row of a trace.
	// Observe that this can be any function which is computable
	// on a given trace --- we are not restricted to expressions
	// which can be arithmetised.
	property T
}

// NewPropertyAssertion constructs a new property assertion!
func NewPropertyAssertion[T Testable](handle string, ctx tr.Context, property T) *PropertyAssertion[T] {
	return &PropertyAssertion[T]{handle, ctx, property}
}

// Handle returns the handle associated with this constraint.
//
//nolint:revive
func (p *PropertyAssertion[T]) Handle() string {
	return p.handle
}

// Context returns the handle associated with this constraint.
//
//nolint:revive
func (p *PropertyAssertion[T]) Context() tr.Context {
	return p.context
}

// Property returns the handle associated with this constraint.
//
//nolint:revive
func (p *PropertyAssertion[T]) Property() T {
	return p.property
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table. If so, return nil otherwise return an error.
//
//nolint:revive
func (p *PropertyAssertion[T]) Accepts(tr tr.Trace) Failure {
	// Determine height of enclosing module
	height := tr.Height(p.context)
	// Iterate every row in the module
	for k := uint(0); k < height; k++ {
		// Check whether property holds (or was undefined)
		if !p.property.TestAt(int(k), tr) {
			// Evaluation failure
			return &AssertionFailure{p.handle, p.property, k}
		}
	}
	// All good
	return nil
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p *PropertyAssertion[T]) Lisp(schema Schema) sexp.SExp {
	// Construct the list
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("assert"),
		sexp.NewSymbol(p.handle),
		p.property.Lisp(schema),
	})
}
