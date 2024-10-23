package constraint

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ZeroTest is a wrapper which converts an Evaluable expression into a Testable
// constraint.  Specifically, by checking whether or not the given expression
// vanishes (i.e. evaluates to zero).
type ZeroTest[E sc.Evaluable] struct {
	Expr E
}

// TestAt determines whether or not a given expression evaluates to zero.
// Observe that if the expression is undefined, then it is assumed not to hold.
func (p ZeroTest[E]) TestAt(row int, tr tr.Trace) bool {
	val := p.Expr.EvalAt(row, tr)
	return val.IsZero()
}

// Bounds determines the bounds for this zero test.
func (p ZeroTest[E]) Bounds() util.Bounds {
	return p.Expr.Bounds()
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p ZeroTest[E]) Context(schema sc.Schema) tr.Context {
	return p.Expr.Context(schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p ZeroTest[E]) RequiredColumns() *util.SortedSet[uint] {
	return p.Expr.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which evaluation of this
// constraint element depends.
func (p ZeroTest[E]) RequiredCells(row int, tr tr.Trace) *util.AnySortedSet[tr.CellRef] {
	return p.Expr.RequiredCells(row, tr)
}

// String generates a human-readble string.
//
//nolint:revive
func (p ZeroTest[E]) String() string {
	return fmt.Sprintf("%s", any(p.Expr))
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p ZeroTest[E]) Lisp(schema sc.Schema) sexp.SExp {
	return p.Expr.Lisp(schema)
}

// VanishingFailure provides structural information about a failing vanishing constraint.
type VanishingFailure struct {
	// Handle of the failing constraint
	handle string
	// Constraint expression
	constraint sc.Testable
	// Row on which the constraint failed
	row uint
}

// Handle returns the handle of this constraint
func (p *VanishingFailure) Handle() string {
	// Construct useful error message
	return p.handle
}

// Message provides a suitable error message
func (p *VanishingFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("constraint \"%s\" does not hold (row %d)", p.handle, p.row)
}

// Constraint returns the constraint expression itself.
func (p *VanishingFailure) Constraint() sc.Testable {
	return p.constraint
}

// Row identifies the row on which this constraint failed.
func (p *VanishingFailure) Row() uint {
	return p.row
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *VanishingFailure) RequiredCells(trace tr.Trace) *util.AnySortedSet[tr.CellRef] {
	return p.constraint.RequiredCells(int(p.row), trace)
}

func (p *VanishingFailure) String() string {
	return p.Message()
}

// VanishingConstraint specifies a constraint which should hold on every row of the
// table.  The only exception is when the constraint is undefined (e.g. because
// it references a non-existent table cell).  In such case, the constraint is
// ignored.  This is parameterised by the type of the constraint expression.
// Thus, we can reuse this definition across the various intermediate
// representations (e.g. Mid-Level IR, Arithmetic IR, etc).
type VanishingConstraint[T sc.Testable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	handle string
	// Evaluation context for this constraint which must match that of the
	// constrained expression itself.
	context tr.Context
	// Indicates (when nil) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given here.
	domain *int
	// The actual constraint itself (e.g. an expression which
	// should evaluate to zero, etc)
	constraint T
}

// NewVanishingConstraint constructs a new vanishing constraint!
func NewVanishingConstraint[T sc.Testable](handle string, context tr.Context,
	domain *int, constraint T) *VanishingConstraint[T] {
	return &VanishingConstraint[T]{handle, context, domain, constraint}
}

// Handle returns the handle associated with this constraint.
//
//nolint:revive
func (p *VanishingConstraint[T]) Handle() string {
	return p.handle
}

// Constraint returns the vanishing expression itself.
func (p *VanishingConstraint[T]) Constraint() T {
	return p.constraint
}

// Domain returns the domain of this constraint.  If the domain is nil, then
// this is a global constraint.  Otherwise this signals a local constraint which
// applies to a specific row (e.g. the first or last).
func (p *VanishingConstraint[T]) Domain() *int {
	return p.domain
}

// Context returns the evaluation context for this constraint.  Every constraint
// must be situated within exactly one module in order to be well-formed.
func (p *VanishingConstraint[T]) Context() tr.Context {
	return p.context
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p *VanishingConstraint[T]) Accepts(tr tr.Trace) schema.Failure {
	if p.domain == nil {
		// Global Constraint
		return HoldsGlobally(p.handle, p.context, p.constraint, tr)
	}
	// Local constraint
	var start uint
	// Handle negative domains
	if *p.domain < 0 {
		// Determine height of enclosing module
		height := tr.Height(p.context)
		// Negative rows calculated from end of trace.
		start = height + uint(*p.domain)
	} else {
		start = uint(*p.domain)
	}
	// Check specific row
	return HoldsLocally(start, p.handle, p.constraint, tr)
}

// HoldsGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func HoldsGlobally[T sc.Testable](handle string, ctx tr.Context, constraint T, tr tr.Trace) schema.Failure {
	// Determine height of enclosing module
	height := tr.Height(ctx)
	// Determine well-definedness bounds for this constraint
	bounds := constraint.Bounds()
	// Sanity check enough rows
	if bounds.End < height {
		// Check all in-bounds values
		for k := bounds.Start; k < (height - bounds.End); k++ {
			if err := HoldsLocally(k, handle, constraint, tr); err != nil {
				return err
			}
		}
	}
	// Success
	return nil
}

// HoldsLocally checks whether a given constraint holds (e.g. vanishes) on a
// specific row of a trace. If not, report an appropriate error.
func HoldsLocally[T sc.Testable](k uint, handle string, constraint T, tr tr.Trace) schema.Failure {
	// Check whether it holds or not
	if !constraint.TestAt(int(k), tr) {
		// Evaluation failure
		return &VanishingFailure{handle, constraint, k}
	}
	// Success
	return nil
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p *VanishingConstraint[T]) Lisp(schema sc.Schema) sexp.SExp {
	attributes := sexp.EmptyList()
	// Handle attributes
	if p.domain == nil {
		// Skip
	} else if *p.domain == 0 {
		attributes.Append(sexp.NewSymbol(":first"))
	} else {
		attributes.Append(sexp.NewSymbol(":last"))
	}
	// Construct the list
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("defconstraint"),
		sexp.NewSymbol(p.handle),
		attributes,
		p.constraint.Lisp(schema),
	})
}
