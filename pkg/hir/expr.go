package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/table"
)

// ============================================================================
// Expressions
// ============================================================================

// Expr is an expression in the High-Level Intermediate Representation (HIR).
// Expressions at this level have a many-2-one correspondance with expressions
// in the AIR level.  For example, an "if" expression at this level will be
// "compiled out" into one or more expressions at the MIR level.
type Expr interface {
	// LowerTo lowers this expression into the Mid-Level Intermediate
	// Representation.  Observe that a single expression at this
	// level can expand into *multiple* expressions at the MIR
	// level.
	LowerTo() []mir.Expr
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, table.Trace) *fr.Element
	// String produces a string representing this as an S-Expression.
	String() string
}

// Add represents the sum over zero or more expressions.
type Add struct{ Args []Expr }

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Expr }

// List represents a block of zero or more expressions.
type List struct{ Args []Expr }

// Constant represents a constant value within an expression.
type Constant struct{ Val *fr.Element }

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

// Normalise reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Normalise struct{ Arg Expr }

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
