package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Expr represents an expression in the Arithmetic Intermediate Representation
// (AIR). Any expression in this form can be lowered into a polynomial.
// Expressions at this level are split into those which can be arithmetised and
// those which cannot.  The latter represent expressions which cannot be
// expressed within a polynomial but can be computed externally (e.g. during
// trace expansion).
type Expr interface {
	util.Boundable
	sc.Evaluable

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
}

// ============================================================================
// Addition
// ============================================================================

// Add represents the sum over zero or more expressions.
type Add struct{ Args []Expr }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Add) Context(schema sc.Schema) trace.Context {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// Add two expressions together, producing a third.
func (p *Add) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Add) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Add) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Add) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Add) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Sub) Context(schema sc.Schema) trace.Context {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// Add two expressions together, producing a third.
func (p *Sub) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Sub) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Sub) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Sub) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Sub) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Expr }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Mul) Context(schema sc.Schema) trace.Context {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// Add two expressions together, producing a third.
func (p *Mul) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Mul) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Mul) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Mul) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Mul) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Constant
// ============================================================================

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

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Constant) Context(schema sc.Schema) trace.Context {
	return trace.VoidContext()
}

// Add two expressions together, producing a third.
func (p *Constant) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *Constant) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *Constant) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *Constant) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *Constant) Bounds() util.Bounds { return util.EMPTY_BOUND }

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

// NewColumnAccess constructs an AIR expression representing the value of a given
// column on the current row.
func NewColumnAccess(column uint, shift int) Expr {
	return &ColumnAccess{column, shift}
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *ColumnAccess) Context(schema sc.Schema) trace.Context {
	col := schema.Columns().Nth(p.Column)
	return col.Context()
}

// Add two expressions together, producing a third.
func (p *ColumnAccess) Add(other Expr) Expr { return &Add{Args: []Expr{p, other}} }

// Sub (subtract) one expression from another.
func (p *ColumnAccess) Sub(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

// Mul (multiply) two expressions together, producing a third.
func (p *ColumnAccess) Mul(other Expr) Expr { return &Mul{Args: []Expr{p, other}} }

// Equate one expression with another (equivalent to subtraction).
func (p *ColumnAccess) Equate(other Expr) Expr { return &Sub{Args: []Expr{p, other}} }

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
