package air

import (
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// An Expression in the Arithmetic Intermediate Representation (AIR).
// Any expression in this form can be lowered into a polynomial.
type Expr interface {
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, trace.Trace) *fr.Element

	// Produce an string representing this as an S-Expression.
	String() string
}

type Nary struct {
	Arguments []Expr
}

type Add Nary
type Sub Nary
type Mul Nary

type Constant struct {
	Value *fr.Element
}

// Represents reading the value held at a given column in the tabular
// context.  Furthermore, the current row maybe shifted up (or down)
// by a given amount.  For example, consider this table:
//
//   +-----+-----+
// k |STAMP| CT  |
//   +-----+-----+
// 0 |  0  |  9  |
//   +-----+-----+
// 1 |  1  |  0  |
//   +-----+-----+
//
// Suppose we are evaluating a constraint on row k=1 which contains
// the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)=1 and
// CT(-1)=9.
type ColumnAccess struct {
	Column string;
	Shift int
}

// A computation-only expression which computes the multiplicative
// inverse of a given expression.
type Inverse struct {
	Expr Expr
}
