package hir

import (
	"encoding/gob"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// Term represents a component of an AIR expression.
type Term interface {
	util.Boundable
	// multiplicity returns the number of underlying expressions that this
	// expression will expand to.
	multiplicity() uint
}

// ============================================================================
// Addition
// ============================================================================

// Add represents the addition of zero or more expressions.
type Add struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Add) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Add) multiplicity() uint {
	count := uint(1)
	//
	for _, e := range p.Args {
		count *= e.multiplicity()
	}
	//
	return count
}

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Sub) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Sub) multiplicity() uint {
	count := uint(1)
	//
	for _, e := range p.Args {
		count *= e.multiplicity()
	}
	//
	return count
}

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Mul) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// multiplicity returns the number of underlying expressions that this
// expression will expand to.
func (p *Mul) multiplicity() uint {
	count := uint(1)
	//
	for _, e := range p.Args {
		count *= e.multiplicity()
	}
	//
	return count
}

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

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Exp) multiplicity() uint {
	return p.Arg.multiplicity()
}

// ============================================================================
// IfZero
// ============================================================================

// IfZero returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero struct {
	// Elements contained within this list.
	Condition Term
	// True branch (optional).
	TrueBranch Term
	// False branch (optional).
	FalseBranch Term
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *IfZero) Bounds() util.Bounds {
	c := p.Condition.Bounds()
	// Get bounds for true branch (if applicable)
	if p.TrueBranch != nil {
		tbounds := p.TrueBranch.Bounds()
		c.Union(&tbounds)
	}
	// Get bounds for false branch (if applicable)
	if p.FalseBranch != nil {
		fbounds := p.FalseBranch.Bounds()
		c.Union(&fbounds)
	}
	// Done
	return c
}

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *IfZero) multiplicity() uint {
	cond := p.Condition.multiplicity()
	count := uint(0)
	// TrueBranch (if applicable)
	if p.TrueBranch != nil {
		count += cond * p.TrueBranch.multiplicity()
	}
	// FalseBranch (if applicable)
	if p.FalseBranch != nil {
		count += cond * p.FalseBranch.multiplicity()
	}
	// done
	return count
}

// ============================================================================
// List
// ============================================================================

// List represents a block of zero or more expressions.
type List struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *List) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *List) multiplicity() uint {
	count := uint(0)
	//
	for _, e := range p.Args {
		count += e.multiplicity()
	}
	//
	return count
}

// ============================================================================
// Constant
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Value fr.Element }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *Constant) Bounds() util.Bounds { return util.EMPTY_BOUND }

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Constant) multiplicity() uint { return 1 }

// ============================================================================
// Normalise
// ============================================================================

// Norm represents the normalisation operator which reduces the value of an
// expression to either zero (if it was zero) or one (otherwise).
type Norm struct{ Arg Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Norm) Bounds() util.Bounds { return p.Arg.Bounds() }

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Norm) multiplicity() uint { return p.Arg.multiplicity() }

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

// multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *ColumnAccess) multiplicity() uint { return 1 }

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(Term(&Add{}))
	gob.Register(Term(&Mul{}))
	gob.Register(Term(&Sub{}))
	gob.Register(Term(&Exp{}))
	gob.Register(Term(&IfZero{}))
	gob.Register(Term(&List{}))
	gob.Register(Term(&Constant{}))
	gob.Register(Term(&Norm{}))
	gob.Register(Term(&ColumnAccess{}))
}
