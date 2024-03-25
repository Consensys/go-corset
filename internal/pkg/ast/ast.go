package ast

import (
	"math/big"
)

// Represents all of the different expression forms within the
// Abstract Syntax Tree (AST).
type Expression interface {
	/// Evaluate this expression at a specific row index within a
	/// given table.
	eval_at() big.Int
}

// / A constant value used within an expression tree.
type Const struct {
	value *big.Int
}

func (e *Const) eval_at() *big.Int {
	return e.value
}
