package ir

import (
	"math/big"
	"github.com/Consensys/go-corset/pkg/trace"
)

// An MirExpression in the Mid-Level Intermediate Representation (MIR).
type MirExpr interface {
	// Lower this MirExpression into the Arithmetic Intermediate
	// Representation.  Essentially, this means eliminating normalising
	// expressions by introducing new columns into the enclosing table (with
	// appropriate constraints).
	LowerToAir() AirExpr
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, trace.Table) *big.Int
}

// ============================================================================
// Definitions
// ============================================================================

type MirAdd Add[MirExpr]
type MirSub Sub[MirExpr]
type MirMul Mul[MirExpr]
type MirConstant = Constant
type MirNormalise Normalise[MirExpr]

// ============================================================================
// Lowering
// ============================================================================

func (e *MirAdd) LowerToAir() AirExpr {
	return &AirAdd{LowerMirExprs(e.arguments)}
}

func (e *MirSub) LowerToAir() AirExpr {
	return &AirSub{LowerMirExprs(e.arguments)}
}

func (e *MirMul) LowerToAir() AirExpr {
	return &AirMul{LowerMirExprs(e.arguments)}
}

func (e *MirNormalise) LowerToAir() AirExpr {
	panic("implement me!")
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirConstant) LowerToAir() AirExpr {
	return e
}

// Lower a set of zero or more MIR expressions.
func LowerMirExprs(exprs []MirExpr) []AirExpr {
	n := len(exprs)
	nexprs := make([]AirExpr, n)
	for i := 0; i < n; i++ {
		nexprs[i] = exprs[i].LowerToAir()
	}
	return nexprs
}

// ============================================================================
// Evaluation
// ============================================================================

func (e *MirAdd) EvalAt(k int, tbl trace.Table) *big.Int {
	// Evaluate first argument
	val := e.arguments[0].EvalAt(k,tbl)
	if val == nil { return nil }
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		ith := e.arguments[i].EvalAt(k,tbl)
		if ith == nil { return ith }
		val.Add(val, ith)
	}
	// Done
	return val
}

func (e *MirSub) EvalAt(k int, tbl trace.Table) *big.Int {
	// Evaluate first argument
	val := e.arguments[0].EvalAt(k,tbl)
	if val == nil { return nil }
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		ith := e.arguments[i].EvalAt(k,tbl)
		if ith == nil { return ith }
		val.Sub(val, ith)
	}
	// Done
	return val
}

func (e *MirMul) EvalAt(k int, tbl trace.Table) *big.Int {
	// Evaluate first argument
	val := e.arguments[0].EvalAt(k,tbl)
	if val == nil { return nil }
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		ith := e.arguments[i].EvalAt(k,tbl)
		if ith == nil { return ith }
		val.Mul(val, ith)
	}
	// Done
	return val
}

func (e *MirNormalise) EvalAt(k int, tbl trace.Table) *big.Int {
	// Check whether argument evaluates to zero or not.
	if e.expr.EvalAt(k,tbl).BitLen() == 0 {
		return big.NewInt(0)
	} else {
		return big.NewInt(1)
	}
}

// Evaluate all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalMirExprsAt(k int, tbl trace.Table, exprs []MirExpr, fn func(*big.Int,*big.Int)) *big.Int {
	// Evaluate first argument
	val := exprs[0].EvalAt(k,tbl)
	if val == nil { return nil }
	// Continue evaluating the rest
	for i := 1; i < len(exprs); i++ {
		ith := exprs[i].EvalAt(k,tbl)
		if ith == nil { return ith }
		fn(val,ith)
	}
	// Done
	return val
}
