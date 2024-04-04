package ir

import (
	"errors"
	"math/big"
	"github.com/Consensys/go-corset/pkg/sexp"
)

// An Expression in the Arithmetic Intermediate Representation (AIR).
// Any expression in this form can be lowered into a polynomial.
type AirExpr interface {
	// Evaluate this expression in the context of a given table.
	EvalAt() *big.Int
}

// ============================================================================
// Definitions
// ============================================================================

type AirAdd Add[AirExpr]
type AirSub Sub[AirExpr]
type AirMul Mul[AirExpr]
type AirConstant = Constant

// ============================================================================
// Evaluation
// ============================================================================

func (e *AirConstant) EvalAt() *big.Int {
	return e.Value
}

func (e *AirAdd) EvalAt() *big.Int {
	// Evaluate first argument
	val := e.arguments[0].EvalAt()
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		val.Add(val, e.arguments[i].EvalAt())
	}
	// Done
	return val
}

func (e *AirSub) EvalAt() *big.Int {
	// Evaluate first argument
	val := e.arguments[0].EvalAt()
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		val.Sub(val, e.arguments[i].EvalAt())
	}
	// Done
	return val
}

func (e *AirMul) EvalAt() *big.Int {
	// Evaluate first argument
	val := e.arguments[0].EvalAt()
	// Continue evaluating the rest
	for i := 1; i < len(e.arguments); i++ {
		val.Mul(val, e.arguments[i].EvalAt())
	}
	// Done
	return val
}

// ============================================================================
// Parser
// ============================================================================

// Parse a string representing an AIR expression formatted using
// S-expressions.
func ParseSExpToAir(s string) (AirExpr,error) {
	// Parse string into S-expression form
	e,err := sexp.Parse(s)
	if err != nil { return nil,err }
	// Process S-expression into AIR expression
	return SExpToAir(e)
}

// Translate an S-Expression into an AIR expression.  Observe that
// this can still fail in the event that the given S-Expression does
// not describe a well-formed AIR expression.
func SExpToAir(s sexp.SExp) (AirExpr,error) {
	switch e := s.(type) {
	case *sexp.List:
		return SExpListToAir(e.Elements)
	case *sexp.Symbol:
		return SExpSymbolToAir(e.Value)
	default:
		panic("invalid S-Expression")
	}
}

// Translate a list of S-Expressions into a unary, binary or n-ary AIR
// expression of some kind.
func SExpListToAir(elements []sexp.SExp) (AirExpr,error) {
	var err error
	// Sanity check this list makes sense
	if len(elements) == 0 || !elements[0].IsSymbol() {
		return nil,errors.New("invalid sexp.List")
	}
	// Extract operator name
	name := (elements[0].(*sexp.Symbol)).Value
	// Translate arguments
	args := make([]AirExpr,len(elements)-1)
	for i,s := range elements[1:] {
		args[i],err = SExpToAir(s)
		if err != nil { return nil,err }
	}
	// Construct expression by name
	switch name {
	case "+":
		return &AirAdd{args},nil
	case "-":
		return &AirSub{args},nil
	case "*":
		return &AirMul{args},nil
	default:
		panic("unknown symbol")
	}
}

func SExpSymbolToAir(symbol string) (AirExpr,error) {
	// Attempt to parse as a number
	num := new(big.Int)
	num,ok := num.SetString(symbol,10)
	if ok { return &AirConstant{num},nil }
	// Not a number!
	panic("Parsing SExp.Symbol")
}
