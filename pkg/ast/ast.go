package ast

import (
	"math/big"
)

// Expression Represents all of the different expression forms within the
// Abstract Syntax Tree (AST).
type Expression interface {
	/// Evaluate this expression at a specific row index within a
	/// given table.
	evalAt() big.Int
}

// Const is a constant value used within an expression tree.
type Const struct {
	value *big.Int
}

func (e *Const) evalAt() *big.Int {
	return e.value
}
