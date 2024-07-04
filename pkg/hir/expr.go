package hir

import (
	"math"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ============================================================================
// Expressions
// ============================================================================

// Expr is an expression in the High-Level Intermediate Representation (HIR).
// Expressions at this level have a many-2-one correspondance with expressions
// in the AIR level.  For example, an "if" expression at this level will be
// "compiled out" into one or more expressions at the MIR level.
type Expr interface {
	util.Boundable
	sc.Contextual
	// LowerTo lowers this expression into the Mid-Level Intermediate
	// Representation.  Observe that a single expression at this
	// level can expand into *multiple* expressions at the MIR
	// level.
	LowerTo(*mir.Schema) []mir.Expr
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAllAt(int, trace.Trace) []*fr.Element
	// String produces a string representing this as an S-Expression.
	String() string
}

// ============================================================================
// Addition
// ============================================================================

// Add represents the sum over zero or more expressions.
type Add struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Add) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Add) Context(schema sc.Schema) (uint, uint, bool) {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Sub) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Sub) Context(schema sc.Schema) (uint, uint, bool) {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Mul) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Mul) Context(schema sc.Schema) (uint, uint, bool) {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// ============================================================================
// List
// ============================================================================

// List represents a block of zero or more expressions.
type List struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *List) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *List) Context(schema sc.Schema) (uint, uint, bool) {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// ============================================================================
// Constant
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Val *fr.Element }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *Constant) Bounds() util.Bounds { return util.EMPTY_BOUND }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Constant) Context(schema sc.Schema) (uint, uint, bool) {
	return math.MaxUint, math.MaxUint, true
}

// ============================================================================
// IfZero
// ============================================================================

// IfZero returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero struct {
	// Elements contained within this list.
	Condition Expr
	// True branch (optional).
	TrueBranch Expr
	// False branch (optional).
	FalseBranch Expr
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

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *IfZero) Context(schema sc.Schema) (uint, uint, bool) {
	if p.TrueBranch != nil && p.FalseBranch != nil {
		args := []Expr{p.Condition, p.TrueBranch, p.FalseBranch}
		return sc.JoinContexts[Expr](args, schema)
	} else if p.TrueBranch != nil {
		// FalseBranch == nil
		args := []Expr{p.Condition, p.TrueBranch}
		return sc.JoinContexts[Expr](args, schema)
	}
	// TrueBranch == nil
	args := []Expr{p.Condition, p.FalseBranch}

	return sc.JoinContexts[Expr](args, schema)
}

// ============================================================================
// Normalise
// ============================================================================

// Normalise reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Normalise struct{ Arg Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Normalise) Bounds() util.Bounds {
	return p.Arg.Bounds()
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Normalise) Context(schema sc.Schema) (uint, uint, bool) {
	return p.Arg.Context(schema)
}

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

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *ColumnAccess) Context(schema sc.Schema) (uint, uint, bool) {
	col := schema.Columns().Nth(p.Column)
	return col.Module(), col.LengthMultiplier(), true
}
