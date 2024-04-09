package ir

import (
	"errors"
	"fmt"
	"github.com/Consensys/go-corset/pkg/sexp"
	"github.com/Consensys/go-corset/pkg/trace"
	"math/big"
	"unicode"
)

// AirExpr An Expression in the Arithmetic Intermediate Representation (AIR).
// Any expression in this form can be lowered into a polynomial.
type AirExpr interface {
	// EvalAt Evaluate this expression in a given tabular context.
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

type AirAdd Add[AirExpr]
type AirSub Sub[AirExpr]
type AirMul Mul[AirExpr]
type AirConstant = Constant
type AirColumnAccess = ColumnAccess

// ============================================================================
// Evaluation
// ============================================================================

func (e *AirColumnAccess) EvalAt(k int, tbl trace.Table) *big.Int {
	val, _ := tbl.GetByName(e.Column, k+e.Shift)
	// We can ignore err as val is always nil when err != nil.
	// Furthermore, as stated in the documentation for this
	// method, we return nil upon error.
	if val == nil {
		// Indicates an out-of-bounds access of some kind.
		return val
	} else {
		var clone big.Int
		// Clone original value
		return clone.Set(val)
	}
}

func (e *AirConstant) EvalAt(k int, tbl trace.Table) *big.Int {
	var clone big.Int
	// Clone original value
	return clone.Set(e.Value)
}

func (e *AirAdd) EvalAt(k int, tbl trace.Table) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Add(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

func (e *AirSub) EvalAt(k int, tbl trace.Table) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Sub(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

func (e *AirMul) EvalAt(k int, tbl trace.Table) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Mul(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

// EvalAirExprsAt Evaluate all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalAirExprsAt(k int, tbl trace.Table, exprs []AirExpr, fn func(*big.Int, *big.Int)) *big.Int {
	// Evaluate first argument
	val := exprs[0].EvalAt(k, tbl)
	if val == nil {
		return nil
	}
	// Continue evaluating the rest
	for i := 1; i < len(exprs); i++ {
		ith := exprs[i].EvalAt(k, tbl)
		if ith == nil {
			return ith
		}
		fn(val, ith)
	}
	// Done
	return val
}

// ============================================================================
// Parser
// ============================================================================

// ParseSExpToAir Parse a string representing an AIR expression formatted using
// S-expressions.
func ParseSExpToAir(s string) (AirExpr, error) {
	// Parse string into S-expression form
	e, err := sexp.Parse(s)
	if err != nil {
		return nil, err
	}
	// Process S-expression into AIR expression
	return SExpToAir(e)
}

// SExpToAir Translate an S-Expression into an AIR expression.  Observe that
// this can still fail in the event that the given S-Expression does
// not describe a well-formed AIR expression.
func SExpToAir(s sexp.SExp) (AirExpr, error) {
	switch e := s.(type) {
	case *sexp.List:
		return SExpListToAir(e.Elements)
	case *sexp.Symbol:
		return SExpSymbolToAir(e.Value)
	default:
		panic("invalid S-Expression")
	}
}

// SExpListToAir Translate a list of S-Expressions into a unary, binary or n-ary AIR
// expression of some kind.
func SExpListToAir(elements []sexp.SExp) (AirExpr, error) {
	var err error
	// Sanity check this list makes sense
	if len(elements) == 0 || !elements[0].IsSymbol() {
		return nil, errors.New("Invalid sexp.List")
	}
	// Extract operator name
	name := (elements[0].(*sexp.Symbol)).Value
	// Translate arguments
	args := make([]AirExpr, len(elements)-1)
	for i, s := range elements[1:] {
		args[i], err = SExpToAir(s)
		if err != nil {
			return nil, err
		}
	}
	// Construct expression by name
	switch name {
	case "+":
		return &AirAdd{args}, nil
	case "-":
		return &AirSub{args}, nil
	case "*":
		return &AirMul{args}, nil
	case "shift":
		if len(args) == 2 {
			// Extract parameters
			c, ok1 := args[0].(*AirColumnAccess)
			n, ok2 := args[1].(*AirConstant)
			// Sanit check this make sense
			if ok1 && ok2 && n.Value.IsInt64() {
				n := int(n.Value.Int64())
				return &AirColumnAccess{c.Column, c.Shift + n}, nil
			} else if !ok1 {
				msg := fmt.Sprintf("Shift column malformed: {%s}", args[0])
				return nil, errors.New(msg)
			} else {
				msg := fmt.Sprintf("Shift amount malformed: {%s}", n)
				return nil, errors.New(msg)
			}
		}
	}
	// Default fall back
	return nil, errors.New("unknown symbol encountered")
}

func SExpSymbolToAir(symbol string) (AirExpr, error) {
	// Attempt to parse as a number
	num := new(big.Int)
	num, ok := num.SetString(symbol, 10)
	if ok {
		return &AirConstant{num}, nil
	}
	// Not a number!
	if isIdentifier(symbol) {
		return &AirColumnAccess{symbol, 0}, nil
	}
	// Problem
	msg := fmt.Sprintf("Invalid symbol: {%s}", symbol)
	return nil, errors.New(msg)
}

// Check whether a given identifier is made up from characters, digits
// or "_" and does not start with a digit.
func isIdentifier(s string) bool {
	for i, c := range s {
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
