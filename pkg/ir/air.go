package ir

import (
	"math/big"
)

// An Expression in the Arithmetic Intermediate Representation (AIR)
type AirExpr interface {
	// Evaluate this AirExpression at a specific row index within a
	// given table.
	EvalAt() *big.Int
}

func (e *Constant) EvalAt() *big.Int {
	return e.Value
}

type AirAdd Add[AirExpr]

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
