package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
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
}

// Add represents the sum over zero or more expressions.
type Add struct{ Args []Expr }

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Expr }

// Constant represents a constant value within an expression.
type Constant struct{ Value *fr.Element }

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
