package constraint

import (
	"errors"
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ZeroTest is a wrapper which converts an Evaluable expression into a Testable
// constraint.  Specifically, by checking whether or not the given expression
// vanishes (i.e. evaluates to zero).
type ZeroTest[E schema.Evaluable] struct {
	Expr E
}

// TestAt determines whether or not a given expression evaluates to zero.
// Observe that if the expression is undefined, then it is assumed not to hold.
func (p ZeroTest[E]) TestAt(row int, tr trace.Trace) bool {
	val := p.Expr.EvalAt(row, tr)
	return val != nil && val.IsZero()
}

// Bounds determines the bounds for this zero test.
func (p ZeroTest[E]) Bounds() util.Bounds {
	return p.Expr.Bounds()
}

// String generates a human-readble string.
//
//nolint:revive
func (p ZeroTest[E]) String() string {
	return fmt.Sprintf("%s", any(p.Expr))
}

// VanishingConstraint specifies a constraint which should hold on every row of the
// table.  The only exception is when the constraint is undefined (e.g. because
// it references a non-existent table cell).  In such case, the constraint is
// ignored.  This is parameterised by the type of the constraint expression.
// Thus, we can reuse this definition across the various intermediate
// representations (e.g. Mid-Level IR, Arithmetic IR, etc).
type VanishingConstraint[T schema.Testable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// Indicates (when nil) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given here.
	Domain *int
	// The actual constraint itself (e.g. an expression which
	// should evaluate to zero, etc)
	Constraint T
}

// NewVanishingConstraint constructs a new vanishing constraint!
func NewVanishingConstraint[T schema.Testable](handle string, domain *int, constraint T) *VanishingConstraint[T] {
	return &VanishingConstraint[T]{handle, domain, constraint}
}

// GetHandle returns the handle associated with this constraint.
func (p *VanishingConstraint[T]) GetHandle() string {
	return p.Handle
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p *VanishingConstraint[T]) Accepts(tr trace.Trace) error {
	if p.Domain == nil {
		// Global Constraint
		return HoldsGlobally(p.Handle, p.Constraint, tr)
	}
	// Check specific row
	return HoldsLocally(*p.Domain, p.Handle, p.Constraint, tr)
}

// HoldsGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func HoldsGlobally[T schema.Testable](handle string, constraint T, tr trace.Trace) error {
	// Determine well-definedness bounds for this constraint
	bounds := constraint.Bounds()
	// Sanity check enough rows
	if bounds.End < tr.Height() {
		// Check all in-bounds values
		for k := bounds.Start; k < (tr.Height() - bounds.End); k++ {
			if err := HoldsLocally(int(k), handle, constraint, tr); err != nil {
				return err
			}
		}
	}
	// Success
	return nil
}

// HoldsLocally checks whether a given constraint holds (e.g. vanishes) on a
// specific row of a trace. If not, report an appropriate error.
func HoldsLocally[T schema.Testable](k int, handle string, constraint T, tr trace.Trace) error {
	// Negative rows calculated from end of trace.
	if k < 0 {
		k += int(tr.Height())
	}
	// Check whether it holds or not
	if !constraint.TestAt(k, tr) {
		// Construct useful error message
		msg := fmt.Sprintf("constraint \"%s\" does not hold (row %d)", handle, k)
		// Evaluation failure
		return errors.New(msg)
	}
	// Success
	return nil
}

// String generates a human-readble string.
//
//nolint:revive
func (p *VanishingConstraint[T]) String() string {
	if p.Domain == nil {
		return fmt.Sprintf("(vanish %s %s)", p.Handle, any(p.Constraint))
	} else if *p.Domain == 0 {
		return fmt.Sprintf("(vanish:first %s %s)", p.Handle, any(p.Constraint))
	}
	//
	return fmt.Sprintf("(vanish:last %s %s)", p.Handle, any(p.Constraint))
}
