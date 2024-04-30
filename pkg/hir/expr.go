package hir

import (
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ============================================================================
// Expressions
// ============================================================================

// An expression in the High-Level Intermediate Representation (HIR).
type Expr interface {
	// Lower this expression into the Mid-Level Intermediate
	// Representation.  Observe that a single expression at this
	// level can expand into *multiple* expressions at the MIR
	// level.
	LowerTo() []mir.Expr
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, table.Trace) *fr.Element
	// Produce an string representing this as an S-Expression.
	String() string
}

type Nary struct { Args[]Expr }
type Add Nary
type Sub Nary
type Mul Nary
type List Nary
type Constant struct { Val *fr.Element }
// Returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero struct {
	// Elements contained within this list.
	Condition Expr
	// True branch (optional).
	TrueBranch Expr
	// False branch (optional).
	FalseBranch Expr
}
//
type Normalise struct { Arg Expr }
type ColumnAccess struct { Column string; Shift int}
