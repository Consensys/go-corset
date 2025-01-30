package constraint

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
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
	Handle string
	// Constraint expression
	Constraint sc.Testable
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *VanishingFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("constraint \"%s\" does not hold (row %d)", p.Handle, p.Row)
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *VanishingFailure) RequiredCells(trace tr.Trace) *util.AnySortedSet[tr.CellRef] {
	return p.Constraint.RequiredCells(int(p.Row), trace)
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
	Handle string
	// Evaluation Context for this constraint which must match that of the
	// constrained expression itself.
	Context tr.Context
	// Indicates (when empty) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given.
	Domain util.Option[int]
	// The actual Constraint itself (e.g. an expression which
	// should evaluate to zero, etc)
	Constraint T
}

// NewVanishingConstraint constructs a new vanishing constraint!
func NewVanishingConstraint[T sc.Testable](handle string, context tr.Context,
	domain util.Option[int], constraint T) *VanishingConstraint[T] {
	return &VanishingConstraint[T]{handle, context, domain, constraint}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p *VanishingConstraint[E]) Bounds(module uint) util.Bounds {
	if p.Context.Module() == module {
		return p.Constraint.Bounds()
	}
	//
	return util.EMPTY_BOUND
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p *VanishingConstraint[T]) Accepts(tr tr.Trace) sc.Failure {
	if p.Domain.IsEmpty() {
		// Global Constraint
		return HoldsGlobally(p.Handle, p.Context, p.Constraint, tr)
	}
	// Extract domain
	domain := p.Domain.Unwrap()
	// Local constraint
	var start uint
	// Handle negative domains
	if domain < 0 {
		// Determine height of enclosing module
		height := tr.Height(p.Context)
		// Negative rows calculated from end of trace.
		start = height + uint(domain)
	} else {
		start = uint(domain)
	}
	// Check specific row
	return HoldsLocally(start, p.Handle, p.Constraint, tr)
}

// HoldsGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func HoldsGlobally[T sc.Testable](handle string, ctx tr.Context, constraint T, tr tr.Trace) sc.Failure {
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
func HoldsLocally[T sc.Testable](k uint, handle string, constraint T, tr tr.Trace) sc.Failure {
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
	var name string
	// Construct qualified name
	if module := schema.Modules().Nth(p.Context.Module()); module.Name != "" {
		name = fmt.Sprintf("%s:%s", module.Name, p.Handle)
	} else {
		name = p.Handle
	}
	// Handle attributes
	if p.Domain.HasValue() {
		domain := p.Domain.Unwrap()
		if domain == 0 {
			name = fmt.Sprintf("%s:first", name)
		} else if domain == -1 {
			name = fmt.Sprintf("%s:last", name)
		} else {
			panic(fmt.Sprintf("domain value %d not supported for local constraint", domain))
		}
	}
	// Determine multiplier
	multiplier := fmt.Sprintf("x%d", p.Context.LengthMultiplier())
	// Construct the list
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("vanish"),
		sexp.NewList([]sexp.SExp{sexp.NewSymbol(name), sexp.NewSymbol(multiplier)}),
		p.Constraint.Lisp(schema),
	})
}
