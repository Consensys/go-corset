package constraint

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
)

// TypeFailure provides structural information about a failing type constraint.
type TypeFailure struct {
	// Handle of the failing constraint
	handle string
	// Constraint expression
	expr sc.Evaluable
	// Row on which the constraint failed
	row uint
}

// Message provides a suitable error message
func (p *TypeFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("expression \"%s\" out-of-bounds (row %d)", p.handle, p.row)
}

func (p *TypeFailure) String() string {
	return p.Message()
}

// TypeConstraint restricts all values for a given expression to be within
// a range [0..n) for some bound n.  Any bound is supported, and the system will
// choose the best underlying implementation as needed.
type TypeConstraint[E sc.Evaluable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	handle string
	// Evaluation context for this constraint which must match that of the
	// constrained expression itself.
	context trace.Context
	// Indicates (when nil) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given here.
	expr E
	// The actual constraint itself, namely an expression which
	// should evaluate to zero.  NOTE: an fr.Element is used here
	// to store the bound simply to make the necessary comparison
	// against table data more direct.
	expected schema.Type
}

// NewTypeConstraint constructs a new Range constraint!
func NewTypeConstraint[E sc.Evaluable](handle string, context trace.Context,
	expr E, expected schema.Type) *TypeConstraint[E] {
	return &TypeConstraint[E]{handle, context, expr, expected}
}

// Handle returns a unique identifier for this constraint.
//
//nolint:revive
func (p *TypeConstraint[E]) Handle() string {
	return p.handle
}

// Context returns the evaluation context for this constraint.
//
//nolint:revive
func (p *TypeConstraint[E]) Context() trace.Context {
	return p.context
}

// Target returns the target expression for this constraint.
func (p *TypeConstraint[E]) Target() E {
	return p.expr
}

// Type returns the expected for all values in the target column.
func (p *TypeConstraint[E]) Type() schema.Type {
	return p.expected
}

// Accepts checks whether a range constraint evaluates to zero on
// every row of a table. If so, return nil otherwise return an error.
//
//nolint:revive
func (p *TypeConstraint[E]) Accepts(tr trace.Trace) schema.Failure {
	// Determine height of enclosing module
	height := tr.Height(p.context)
	// Iterate every row
	for k := 0; k < int(height); k++ {
		// Get the value on the kth row
		kth := p.expr.EvalAt(k, tr)
		// Perform the type check
		if !p.expected.Accept(kth) {
			// Evaluation failure
			return &TypeFailure{p.handle, p.expr, uint(k)}
		}
	}
	// All good
	return nil
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *TypeConstraint[E]) Lisp(schema sc.Schema) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("definrange"),
		p.expr.Lisp(schema),
		sexp.NewSymbol(p.expected.String()),
	})
}
