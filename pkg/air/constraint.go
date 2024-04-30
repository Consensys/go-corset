package air

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

// Constraint for now, all constraints are vanishing constraints.
type Constraint interface {
	table.Constraint
	// IsAir is a marker intended to signal that this a column at the lowest level.
	IsAir() bool
}

// ===================================================================
// Vanishing Constraint
// ===================================================================

// VanishingConstraint on every row of the table, a vanishing constraint must evaluate to
// zero.  The only exception is when the constraint is undefined
// (e.g. because it references a non-existent table cell).  In such
// case, the constraint is ignored.  This is parameterised by the type
// of the constraint expression.  Thus, we can reuse this definition
// across the various intermediate representations (e.g. Mid-Level IR,
// Arithmetic IR, etc).
type VanishingConstraint struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// The actual constraint itself, namely an expression which
	// should evaluate to zero.
	Expr Expr
}

func (p *VanishingConstraint) GetHandle() string {
	return p.Handle
}

func (p *VanishingConstraint) IsAir() bool { return true }

// Check whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
func (p *VanishingConstraint) Accepts(tr table.Trace) error {
	for k := 0; k < tr.Height(); k++ {
		// Determine kth evaluation point
		kth := p.Expr.EvalAt(k, tr)
		// Check whether it vanished (or was undefined)
		if kth != nil && !kth.IsZero() {
			// Construct useful error message
			msg := fmt.Sprintf("constraint %s does not vanish (row %d, %s)", p.Handle, k, kth)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// Success!
	return nil
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

func (p *RangeConstraint) GetHandle() string {
	return p.Handle
}

func (p *RangeConstraint) IsAir() bool { return true }

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table. If so, return nil otherwise return an error.
func (p *RangeConstraint) Accepts(tr table.Trace) error {
	for k := 0; k < tr.Height(); k++ {
		// Get the value on the kth row
		kth, err := tr.GetByName(p.Handle, k)
		// Sanity check column exists!
		if err != nil {
			return err
		}
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
