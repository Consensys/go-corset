package mir

import (
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// An expression in the Mid-Level Intermediate Representation (MIR).
type Expr interface {
	// Lower this expression into the Arithmetic Intermediate
	// Representation.  Essentially, this means eliminating
	// normalising expressions by introducing new columns into the
	// given table (with appropriate constraints).
	LowerTo(air.Schema) air.Expr
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, trace.Trace) *fr.Element
}

type Nary struct { Arguments[]Expr }
type Add Nary
type Sub Nary
type Mul Nary

type Constant struct {
	Value *fr.Element
}

type Normalise struct {
	Expr Expr
}

type ColumnAccess struct {
	Column string;
	Shift int
}
