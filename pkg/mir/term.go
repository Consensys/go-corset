package mir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// Term represents a component of an AIR expression.
type Term interface {
	util.Boundable
}

// ============================================================================
// Addition
// ============================================================================

type Add struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Add) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Sub) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Mul) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Exponentiation
// ============================================================================

// Exp represents the a given value taken to a power.
type Exp struct {
	Arg Term
	Pow uint64
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Exp) Bounds() util.Bounds { return p.Arg.Bounds() }

// ============================================================================
// Constant
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Value fr.Element }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *Constant) Bounds() util.Bounds { return util.EMPTY_BOUND }

// ============================================================================
// Normalise
// ============================================================================

// Normalise reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Normalise struct{ Arg Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Normalise) Bounds() util.Bounds { return p.Arg.Bounds() }

// ============================================================================
// ColumnAccess
// ============================================================================

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount. Suppose we are evaluating a constraint on row k=5 which
// contains the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)
// accesses the STAMP column at row 5, whilst CT(-1) accesses the CT column at
// row 4.
type ColumnAccess struct {
	Column uint
	Shift  int
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *ColumnAccess) Bounds() util.Bounds {
	if p.Shift >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.Shift))
	}
	// Negative shift
	return util.NewBounds(uint(-p.Shift), 0)
}
