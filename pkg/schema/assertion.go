package schema

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// AssertionFailure provides structural information about a failing vanishing constraint.
type AssertionFailure struct {
	// Handle of the failing constraint
	Handle string
	// Constraint expression
	Constraint Testable
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *AssertionFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("assertion \"%s\" does not hold (row %d)", p.Handle, p.Row)
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *AssertionFailure) RequiredCells(trace tr.Trace) *util.AnySortedSet[tr.CellRef] {
	return p.Constraint.RequiredCells(int(p.Row), trace)
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
	Handle string
	// Enclosing module for this assertion.  This restricts the asserted
	// property to access only columns from within this module.
	Context tr.Context
	// The actual assertion itself, namely an expression which
	// should hold (i.e. vanish) for every row of a trace.
	// Observe that this can be any function which is computable
	// on a given trace --- we are not restricted to expressions
	// which can be arithmetised.
	Property T
}

// NewPropertyAssertion constructs a new property assertion!
func NewPropertyAssertion[T Testable](handle string, ctx tr.Context, property T) *PropertyAssertion[T] {
	return &PropertyAssertion[T]{handle, ctx, property}
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table. If so, return nil otherwise return an error.
//
//nolint:revive
func (p *PropertyAssertion[T]) Accepts(tr tr.Trace) Failure {
	// Determine height of enclosing module
	height := tr.Height(p.Context)
	// Iterate every row in the module
	for k := uint(0); k < height; k++ {
		// Check whether property holds (or was undefined)
		if !p.Property.TestAt(int(k), tr) {
			// Evaluation failure
			return &AssertionFailure{p.Handle, p.Property, k}
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
		sexp.NewSymbol(p.Handle),
		p.Property.Lisp(schema),
	})
}
