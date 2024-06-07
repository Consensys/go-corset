package table

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Evaluable captures something which can be evaluated on a given table row to
// produce an evaluation point.  For example, expressions in the
// Mid-Level or Arithmetic-Level IR can all be evaluated at rows of a
// table.
type Evaluable interface {
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, Trace) *fr.Element
}

// Testable captures the notion of a constraint which can be tested on a given
// row of a given trace.  It is very similar to Evaluable, except that it only
// indicates success or failure.  The reason for using this interface over
// Evaluable is that, for historical reasons, constraints at the HIR cannot be
// Evaluable (i.e. because they return multiple values, rather than a single
// value).  However, constraints at the HIR level remain testable.
type Testable interface {
	// TestAt evaluates this expression in a given tabular context and checks it
	// against zero. Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be undefined for
	// several reasons: firstly, if it accesses a row which does not exist (e.g.
	// at index -1); secondly, if it accesses a column which does not exist.
	TestAt(int, Trace) bool
}

// ===================================================================
// Vanishing Constraints
// ===================================================================

// ZeroTest is a wrapper which converts an Evaluable expression into a Testable
// constraint.  Specifically, by checking whether or not the given expression
// vanishes (i.e. evaluates to zero).
type ZeroTest[E Evaluable] struct {
	Expr E
}

// TestAt determines whether or not a given expression evaluates to zero.
// Observe that if the expression is undefined, then it is assumed to hold.
func (p ZeroTest[E]) TestAt(row int, tr Trace) bool {
	val := p.Expr.EvalAt(row, tr)
	return val == nil || val.IsZero()
}

func (p ZeroTest[E]) String() string {
	return fmt.Sprintf("%s", any(p.Expr))
}

// RowConstraint specifies a constraint which should hold on every row of the
// table.  The only exception is when the constraint is undefined (e.g. because
// it references a non-existent table cell).  In such case, the constraint is
// ignored.  This is parameterised by the type of the constraint expression.
// Thus, we can reuse this definition across the various intermediate
// representations (e.g. Mid-Level IR, Arithmetic IR, etc).
type RowConstraint[T Testable] struct {
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

// NewRowConstraint constructs a new vanishing constraint!
func NewRowConstraint[T Testable](handle string, domain *int, constraint T) *RowConstraint[T] {
	return &RowConstraint[T]{handle, domain, constraint}
}

// GetHandle returns the handle associated with this constraint.
func (p *RowConstraint[T]) GetHandle() string {
	return p.Handle
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p *RowConstraint[T]) Accepts(tr Trace) error {
	if p.Domain == nil {
		// Global Constraint
		return HoldsGlobally(p.Handle, p.Constraint, tr)
	}
	// Check specific row
	return HoldsLocally(*p.Domain, p.Handle, p.Constraint, tr)
}

// HoldsGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func HoldsGlobally[T Testable](handle string, constraint T, tr Trace) error {
	for k := 0; k < tr.Height(); k++ {
		if err := HoldsLocally(k, handle, constraint, tr); err != nil {
			return err
		}
	}
	// Success
	return nil
}

// HoldsLocally checks whether a given constraint holds (e.g. vanishes) on a
// specific row of a trace. If not, report an appropriate error.
func HoldsLocally[T Testable](k int, handle string, constraint T, tr Trace) error {
	// Negative rows calculated from end of trace.
	if k < 0 {
		k += tr.Height()
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

func (p *RowConstraint[T]) String() string {
	if p.Domain == nil {
		return fmt.Sprintf("(vanish %s %s)", p.Handle, any(p.Constraint))
	} else if *p.Domain == 0 {
		return fmt.Sprintf("(vanish:first %s %s)", p.Handle, any(p.Constraint))
	}
	//
	return fmt.Sprintf("(vanish:last %s %s)", p.Handle, any(p.Constraint))
}

// ===================================================================
// Range Constraint
// ===================================================================

// RangeConstraint restricts all values in a given column to be within
// a range [0..n) for some bound n.  For example, a bound of 256 would
// restrict all values to be bytes.  At this time, range constraints
// are explicitly limited at the arithmetic level to bounds of at most
// 256 (i.e. to ensuring bytes).  This restriction is somewhat
// arbitrary and is determined by the underlying prover.
type RangeConstraint struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// The actual constraint itself, namely an expression which
	// should evaluate to zero.  NOTE: an fr.Element is used here
	// to store the bound simply to make the necessary comparison
	// against table data more direct.
	Bound *fr.Element
}

// NewRangeConstraint constructs a new Range constraint!
func NewRangeConstraint(column string, bound *fr.Element) *RangeConstraint {
	var n fr.Element = fr.NewElement(256)
	if bound.Cmp(&n) > 0 {
		panic("Range constraint for bitwidth above 8 not supported")
	}

	return &RangeConstraint{column, bound}
}

// GetHandle returns the handle associated with this constraint.
func (p *RangeConstraint) GetHandle() string {
	return p.Handle
}

// IsAir is a marker that indicates this is an AIR column.
func (p *RangeConstraint) IsAir() bool { return true }

// Accepts checks whether a range constraint evaluates to zero on
// every row of a table. If so, return nil otherwise return an error.
func (p *RangeConstraint) Accepts(tr Trace) error {
	for k := 0; k < tr.Height(); k++ {
		// Get the value on the kth row
		kth := tr.GetByName(p.Handle, k)
		// Perform the bounds check
		if kth != nil && kth.Cmp(p.Bound) >= 0 {
			// Construct useful error message
			msg := fmt.Sprintf("value out-of-bounds (row %d, %s)", kth, p.Handle)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}

func (p *RangeConstraint) String() string {
	return fmt.Sprintf("(range %s %s)", p.Handle, p.Bound)
}

// ===================================================================
// Property Assertion
// ===================================================================

// PropertyAssertion is similar to a vanishing constraint but is used only for
// debugging / testing / verification.  Unlike vanishing constraints, property
// assertions do not represent something that the prover can enforce.  Rather,
// they represent properties which are expected to hold for every valid trace.
// That is, they should be implied by the actual constraints.  Thus, whilst the
// prover cannot enforce such properties, external tools (such as for formal
// verification) can attempt to ensure they do indeed always hold.
type PropertyAssertion[E Evaluable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// The actual assertion itself, namely an expression which
	// should hold (i.e. vanish) for every row of a trace.
	// Observe that this can be any function which is computable
	// on a given trace --- we are not restricted to expressions
	// which can be arithmetised.
	Expr E
}

// GetHandle returns the handle associated with this constraint.
//
//nolint:revive
func (p *PropertyAssertion[E]) GetHandle() string {
	return p.Handle
}

// NewPropertyAssertion constructs a new property assertion!
func NewPropertyAssertion[E Evaluable](handle string, expr E) *PropertyAssertion[E] {
	return &PropertyAssertion[E]{handle, expr}
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table. If so, return nil otherwise return an error.
//
//nolint:revive
func (p *PropertyAssertion[E]) Accepts(tr Trace) error {
	for k := 0; k < tr.Height(); k++ {
		// Determine kth evaluation point
		kth := p.Expr.EvalAt(k, tr)
		// Check whether it vanished (or was undefined)
		if kth != nil && !kth.IsZero() {
			// Construct useful error message
			msg := fmt.Sprintf("property assertion %s does not hold (row %d, %s)", p.Handle, k, kth)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}
