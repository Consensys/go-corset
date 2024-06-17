package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/go-corset/pkg/util"
)

// Expr represents an expression in the Arithmetic Intermediate Representation
// (AIR). Any expression in this form can be lowered into a polynomial.
// Expressions at this level are split into those which can be arithmetised and
// those which cannot.  The latter represent expressions which cannot be
// expressed within a polynomial but can be computed externally (e.g. during
// trace expansion).
type Expr interface {
	// EvalAt evaluates this expression in a given tabular context. Observe that
	// if this expression is *undefined* within this context then it returns
	// "nil".  An expression can be undefined for several reasons: firstly, if
	// it accesses a row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, table.Trace) *fr.Element

	// String produces a string representing this as an S-Expression.
	String() string

	// Add two expressions together, producing a third.
	Add(Expr) Expr

	// Subtract one expression from another
	Sub(Expr) Expr

	// Multiply two expressions together, producing a third.
	Mul(Expr) Expr

	// Equate one expression with another
	Equate(Expr) Expr

	// Determine the maximum shift in this expression in either the negative
	// (left) or positive direction (right).
	MaxShift() util.Pair[uint, uint]
}

// Add represents the sum over zero or more expressions.
type Add struct{ Args []Expr }

// Add two expressions together, producing a third.
func (p *Add) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Add) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Add) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Add) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// MaxShift returns max shift in either the negative (left) or positive
// direction (right).
func (p *Add) MaxShift() util.Pair[uint, uint] { return maxShiftOfArray(p.Args) }

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// Add two expressions together, producing a third.
func (p *Sub) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Sub) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Sub) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Sub) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// MaxShift returns max shift in either the negative (left) or positive
// direction (right).
func (p *Sub) MaxShift() util.Pair[uint, uint] { return maxShiftOfArray(p.Args) }

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Expr }

// Add two expressions together, producing a third.
func (p *Mul) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Mul) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Mul) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Mul) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// MaxShift returns max shift in either the negative (left) or positive
// direction (right).
func (p *Mul) MaxShift() util.Pair[uint, uint] { return maxShiftOfArray(p.Args) }

// Constant represents a constant value within an expression.
type Constant struct{ Value *fr.Element }

// NewConst construct an AIR expression representing a given constant.
func NewConst(val *fr.Element) Expr {
	return &Constant{val}
}

// NewConst64 construct an AIR expression representing a given constant from a
// uint64.
func NewConst64(val uint64) Expr {
	element := fr.NewElement(val)
	return &Constant{&element}
}

// NewConstCopy construct an AIR expression representing a given constant,
// and also clones that constant.
func NewConstCopy(val *fr.Element) Expr {
	// Create ith term (for final sum)
	var clone fr.Element
	// Clone coefficient
	clone.Set(val)
	// DOne
	return &Constant{&clone}
}

// Add two expressions together, producing a third.
func (p *Constant) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Constant) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Constant) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Constant) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// MaxShift returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *Constant) MaxShift() util.Pair[uint, uint] { return util.NewPair[uint, uint](0, 0) }

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount. Suppose we are evaluating a constraint on row k=5 which
// contains the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)
// accesses the STAMP column at row 5, whilst CT(-1) accesses the CT column at
// row 4.
type ColumnAccess struct {
	Column string
	Shift  int
}

// NewColumnAccess constructs an AIR expression representing the value of a given
// column on the current row.
func NewColumnAccess(name string, shift int) Expr {
	return &ColumnAccess{name, shift}
}

// Add two expressions together, producing a third.
func (p *ColumnAccess) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *ColumnAccess) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *ColumnAccess) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *ColumnAccess) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// MaxShift returns max shift in either the negative (left) or positive
// direction (right).
func (p *ColumnAccess) MaxShift() util.Pair[uint, uint] {
	if p.Shift >= 0 {
		// Positive shift
		return util.NewPair[uint, uint](0, uint(p.Shift))
	}
	// Negative shift
	return util.NewPair[uint, uint](uint(-p.Shift), 0)
}

// ==========================================================================
// Helpers
// ==========================================================================

func maxShiftOfArray(args []Expr) util.Pair[uint, uint] {
	neg := uint(0)
	pos := uint(0)

	for _, e := range args {
		mx := e.MaxShift()
		neg = max(neg, mx.Left)
		pos = max(pos, mx.Right)
	}
	// Done
	return util.NewPair(neg, pos)
}
