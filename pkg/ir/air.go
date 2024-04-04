package ir

import (
	"math/big"
)

type Polynomial any // for now

// An Expression in the Arithmetic Intermediate Representation (AIR).
// Any expression in this form can be lowered into a polynomial.
type AirExpr interface {
	// Lower this expression into a minimal polynomial form.
	LowerToPolynomial() Polynomial
	// Evaluate this expression in the context of a given table.
	EvalAt() *big.Int
}

// ============================================================================
// Definitions
// ============================================================================

type AirAdd Add[AirExpr]
type AirConstant = Constant

// ============================================================================
// Lowering
// ============================================================================

func (*AirConstant) LowerToPolynomial() Polynomial {
	panic("to do")
}

func (*AirAdd) LowerToPolynomial() Polynomial {
	panic("to do")
}

// ============================================================================
// Evaluation
// ============================================================================

func (e *AirConstant) EvalAt() *big.Int {
	return e.Value
}

func (e *AirAdd) EvalAt() *big.Int {
	// Evaluate first argument
	sum := e.arguments[0].EvalAt()
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		sum.Add(sum, e.arguments[i].EvalAt())
	}
	// Done
	return sum
}
