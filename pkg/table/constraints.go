package table

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ===================================================================
// Vanishing Constraints
// ===================================================================

// VanishingConstraint on every row of the table, a vanishing
// constraint must evaluate to zero.  The only exception is when the
// constraint is undefined (e.g. because it references a non-existent
// table cell).  In such case, the constraint is ignored.  This is
// parameterised by the type of the constraint expression.  Thus, we
// can reuse this definition across the various intermediate
// representations (e.g. Mid-Level IR, Arithmetic IR, etc).
type VanishingConstraint[T Evaluable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// Indicates (when nil) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given here.
	Domain *int
	// The actual constraint itself, namely an expression which
	// should evaluate to zero.
	Expr T
}

// NewVanishingConstraint constructs a new vanishing constraint!
func NewVanishingConstraint[T Evaluable](handle string, domain *int, expr T) *VanishingConstraint[T] {
	return &VanishingConstraint[T]{handle, domain, expr}
}

// GetHandle returns the handle associated with this constraint.
func (p *VanishingConstraint[T]) GetHandle() string {
	return p.Handle
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p *VanishingConstraint[T]) Accepts(tr Trace) (bool, error) {
	if p.Domain == nil {
		// Global Constraint
		return false, VanishesGlobally(p.Handle, p.Expr, tr)
	}
	// Check specific row
	return false, VanishesLocally(*p.Domain, p.Handle, p.Expr, tr)
}

// VanishesGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func VanishesGlobally[E Evaluable](handle string, expr E, tr Trace) error {
	for k := 0; k < tr.Height(); k++ {
		if err := VanishesLocally(k, handle, expr, tr); err != nil {
			return err
		}
	}
	// Success
	return nil
}

// VanishesLocally checks whether a given expression vanishes (i.e. evaluates to zero)
// on a specific row of a trace. If not, report an appropriate error.
func VanishesLocally[E Evaluable](k int, handle string, expr E, tr Trace) error {
	// Negative rows calculated from end of trace.
	if k < 0 {
		k += tr.Height()
	}
	// Determine kth evaluation point
	kth := expr.EvalAt(k, tr)
	// Check whether it vanished (or was undefined)
	if kth != nil && !kth.IsZero() {
		// Construct useful error message
		msg := fmt.Sprintf("constraint %s does not vanish (row %d, %s)", handle, k, kth)
		// Evaluation failure
		return errors.New(msg)
	}
	// Success
	return nil
}

func (p *VanishingConstraint[T]) String() string {
	if p.Domain == nil {
		return fmt.Sprintf("(vanishes %s %s)", p.Handle, any(p.Expr))
	} else if *p.Domain == 0 {
		return fmt.Sprintf("(vanishes:first %s %s)", p.Handle, any(p.Expr))
	}
	//
	return fmt.Sprintf("(vanishes:last %s %s)", p.Handle, any(p.Expr))
}

// ===================================================================
// Range Constraint
// ===================================================================

// RangeConstraint restricts all values in a given column to be
// within a range [0..n) for some bound n.  For example, a bound of
// 256 would restrict all values to be bytes.
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
	return &RangeConstraint{column, bound}
}

// GetHandle returns the handle associated with this constraint.
func (p *RangeConstraint) GetHandle() string {
	return p.Handle
}

// IsAir is a marker that indicates this is an AIR column.
func (p *RangeConstraint) IsAir() bool { return true }

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table. If so, return nil otherwise return an error.
func (p *RangeConstraint) Accepts(tr Trace) (bool, error) {
	for k := 0; k < tr.Height(); k++ {
		// Get the value on the kth row
		kth, err := tr.GetByName(p.Handle, k)
		// Sanity check column exists!
		if err != nil {
			return false, err
		}
		// Perform the bounds check
		if kth != nil && kth.Cmp(p.Bound) >= 0 {
			// Construct useful error message
			msg := fmt.Sprintf("value out-of-bounds (row %d, %s)", kth, p.Handle)
			// Evaluation failure
			return false, errors.New(msg)
		}
	}
	// All good
	return false, nil
}

// ===================================================================
// Property Assertion
// ===================================================================

// Assertion is similar to a vanishing constraint but is used only for
// debugging / testing / verification.  Unlike vanishing constraints,
// property assertions do not represent something that the prover can
// enforce.  Rather, they represent properties which are expected to
// hold for every valid trace.  That is, they should be implied by the
// actual constraints.  Thus, whilst the prover cannot enforce such
// properties, external tools (such as for formal verification) can
// attempt to ensure they do indeed always hold.
type Assertion struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// The actual assertion itself, namely an expression which
	// should hold (i.e. vanish) for every row of a trace.
	// Observe that this can be any function which is computable
	// on a given trace --- we are not restricted to expressions
	// which can be arithmetised.
	Expr Evaluable
}

// GetHandle returns the handle associated with this constraint.
func (p *Assertion) GetHandle() string {
	return p.Handle
}
