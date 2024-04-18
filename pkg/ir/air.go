package ir

import (
	"github.com/Consensys/go-corset/pkg/trace"
	"math/big"
)

// ============================================================================
// Table
// ============================================================================

type AirTable = trace.Table[AirConstraint]

// For now, all constraints are vanishing constraints.
type AirConstraint = *trace.VanishingConstraint[AirExpr]

// ============================================================================
// Expressions
// ============================================================================

// An Expression in the Arithmetic Intermediate Representation (AIR).
// Any expression in this form can be lowered into a polynomial.
type AirExpr interface {
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, trace.Trace) *big.Int
}

type AirAdd Add[AirExpr]
type AirSub Sub[AirExpr]
type AirMul Mul[AirExpr]
type AirConstant struct { Val *big.Int }
type AirColumnAccess struct { Col string; Amt int}

// MirConstant implements Constant interface
func (e *AirConstant) Value() *big.Int { return e.Val }
// MirColumnAccess implements ColumnAccess interface
func (e *AirColumnAccess) Column() string { return e.Col }
func (e *AirColumnAccess) Shift() int { return e.Amt }

// ============================================================================
// Evaluation
// ============================================================================

func (e *AirColumnAccess) EvalAt(k int, tbl trace.Trace) *big.Int {
	val, _ := tbl.GetByName(e.Column(), k + e.Shift())
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

func (e *AirConstant) EvalAt(k int, tbl trace.Trace) *big.Int {
	var clone big.Int
	// Clone original value
	return clone.Set(e.Val)
}

func (e *AirAdd) EvalAt(k int, tbl trace.Trace) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Add(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

func (e *AirSub) EvalAt(k int, tbl trace.Trace) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Sub(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

func (e *AirMul) EvalAt(k int, tbl trace.Trace) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Mul(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

// Evaluate all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalAirExprsAt(k int, tbl trace.Trace, exprs []AirExpr, fn func(*big.Int, *big.Int)) *big.Int {
	// Evaluate first argument
	val := exprs[0].EvalAt(k, tbl)
	if val == nil { return nil }
	// Continue evaluating the rest
	for i := 1; i < len(exprs); i++ {
		ith := exprs[i].EvalAt(k, tbl)
		if ith == nil { return ith }
		fn(val, ith)
	}
	// Done
	return val
}

// ============================================================================
// Parser
// ============================================================================

// Parse a string representing an AIR expression formatted using
// S-expressions.
func ParseSExpToAir(s string) (AirExpr, error) {
	parser := NewIrParser[AirExpr]()
	// Configure parser
	AddSymbolTranslator(&parser, SExpConstantToAir)
	AddSymbolTranslator(&parser, SExpColumnToAir)
	AddListTranslator(&parser, "+", SExpAddToAir)
	AddListTranslator(&parser, "-", SExpSubToAir)
	AddListTranslator(&parser, "*", SExpMulToAir)
	AddListTranslator(&parser, "shift", SExpShiftToAir)
	// Parse string
	return Parse(parser, s)
}

func SExpConstantToAir(symbol string) (AirExpr, error) {
	c,err := StringToConstant(symbol)
	if err != nil { return nil,err }
	return &AirConstant{c},nil
}
func SExpColumnToAir(symbol string) (AirExpr, error) {
	c,n,err := StringToColumnAccess(symbol)
	if err != nil { return nil,err }
	return &AirColumnAccess{c,n},nil
}
func SExpAddToAir(args []AirExpr) (AirExpr, error)   { return &AirAdd{args}, nil }
func SExpSubToAir(args []AirExpr) (AirExpr, error)   { return &AirSub{args}, nil }
func SExpMulToAir(args []AirExpr) (AirExpr, error)   { return &AirMul{args}, nil }

func SExpShiftToAir(args []AirExpr) (AirExpr, error) {
	c,n,err := SliceToShiftAccess(args)
	if err != nil { return nil,err }
	return &AirColumnAccess{c,n},nil
}
