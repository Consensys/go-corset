package ir

import (
	"errors"
	"fmt"
	"math/big"
	"unicode"
	"github.com/Consensys/go-corset/pkg/sexp"
)

// An Expression in the Arithmetic Intermediate Representation (AIR).
// Any expression in this form can be lowered into a polynomial.
type AirExpr interface {
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt() *big.Int
}

// ============================================================================
// Definitions
// ============================================================================

type AirAdd Add[AirExpr]
type AirSub Sub[AirExpr]
type AirMul Mul[AirExpr]
type AirConstant = Constant
type AirColumnAccess = ColumnAccess

// ============================================================================
// Evaluation
// ============================================================================

func (e *AirColumnAccess) EvalAt() *big.Int {
	panic("got here") // todo
}

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
		return nil,errors.New("Invalid sexp.List")
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
		panic("Unknown symbol")
	}
}

func SExpSymbolToAir(symbol string) (AirExpr,error) {
	// Attempt to parse as a number
	num := new(big.Int)
	num,ok := num.SetString(symbol,10)
	if ok { return &AirConstant{num},nil }
	// Not a number!
	if isIdentifier(symbol) {
		return &AirColumnAccess{symbol,0},nil
	}
	// Problem
	msg := fmt.Sprintf("Invalid symbol: {%s}",symbol)
	return nil,errors.New(msg)
}

// Check whether a given identifier is made up from characters, digits
// or "_" and does not start with a digit.
func isIdentifier(s string) bool {
	for i,c := range s {
		if unicode.IsLetter(c) || c == '_' {
			// OK
		} else if i != 0 && unicode.IsNumber(c) {
			// Also OK
		} else {
			// Otherwise, not OK.
			return false
		}
	}
	return true
}
