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
	// Evaluate this expression in the context of a given table.
	EvalAt(int, trace.Table) *big.Int
}

// ============================================================================
// Definitions
// ============================================================================

type MirAdd Add[MirExpr]
type MirConstant = Constant
type MirNormalise Normalise[MirExpr]

// ============================================================================
// Lowering
// ============================================================================

func (e *MirAdd) LowerToAir() AirExpr {
	n := len(e.arguments)
	nargs := make([]AirExpr, n)
	for i := 0; i < n; i++ {
		nargs[i] = e.arguments[i].LowerToAir()
	}
	return &AirAdd{nargs}
}

func (e *MirNormalise) LowerToAir() AirExpr {
	panic("implement me!")
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirConstant) LowerToAir() AirExpr {
	return e
}

// ============================================================================
// Evaluation
// ============================================================================

func (e *MirAdd) EvalAt(k int, tbl trace.Table) *big.Int {
	// Evaluate first argument
	sum := e.arguments[0].EvalAt(k,tbl)
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		sum.Add(sum, e.arguments[i].EvalAt(k,tbl))
	}
	// Done
	return sum
}

func (e *MirNormalise) EvalAt(k int, tbl trace.Table) *big.Int {
	// Check whether argument evaluates to zero or not.
	if e.expr.EvalAt(k,tbl).BitLen() == 0 {
		return big.NewInt(0)
	} else {
		return big.NewInt(1)
	}
}
