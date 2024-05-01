package table

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Constraint is an abstract notion of a constraint which must hold true for a given
// table.
type Constraint interface {
	// GetHandle gets the handle for this constraint (i.e. its name).
	GetHandle() string
	// Accepts checks whether this constraint accepts a particular trace.
	Accepts(Trace) error
}

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

// ===================================================================
// Vanishing Constraints
// ===================================================================

// VanishingConstraint on every row of the table, a vanishing constraint must evaluate to
// zero.  The only exception is when the constraint is undefined
// (e.g. because it references a non-existent table cell).  In such
// case, the constraint is ignored.  This is parameterised by the type
// of the constraint expression.  Thus, we can reuse this definition
// across the various intermediate representations (e.g. Mid-Level IR,
// Arithmetic IR, etc).
type VanishingConstraint[T Evaluable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// The actual constraint itself, namely an expression which
	// should evaluate to zero.
	Expr T
}

// GetHandle returns the handle associated with this constraint.
func (p *VanishingConstraint[T]) GetHandle() string {
	return p.Handle
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
func (p *VanishingConstraint[T]) Accepts(tr Trace) error {
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

func (p *VanishingConstraint[T]) String() string {
	return fmt.Sprintf("(vanishes %s %s)", p.Handle, any(p.Expr))
}
