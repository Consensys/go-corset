package ir

import (
	"fmt"
	"strconv"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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
	EvalAt(int, trace.Trace) *fr.Element

	// Produce an string representing this as an S-Expression.
	String() string
}

type AirAdd Add[AirExpr]
type AirSub Sub[AirExpr]
type AirMul Mul[AirExpr]
type AirConstant struct { Val *fr.Element }
type AirColumnAccess struct { Col string; Amt int}
// A computation-only expression which computes the multiplicative
// inverse of a given expression.
type AirInverse struct { expr AirExpr }

// MirConstant implements Constant interface
func (e *AirConstant) Value() *fr.Element { return e.Val }
// MirColumnAccess implements ColumnAccess interface
func (e *AirColumnAccess) Column() string { return e.Col }
func (e *AirColumnAccess) Shift() int { return e.Amt }

// ============================================================================
// Evaluation
// ============================================================================

func (e *AirColumnAccess) EvalAt(k int, tbl trace.Trace) *fr.Element {
	val, _ := tbl.GetByName(e.Column(), k + e.Shift())
	// We can ignore err as val is always nil when err != nil.
	// Furthermore, as stated in the documentation for this
	// method, we return nil upon error.
	if val == nil {
		// Indicates an out-of-bounds access of some kind.
		return val
	} else {
		var clone fr.Element
		// Clone original value
		return clone.Set(val)
	}
}

func (e *AirConstant) EvalAt(k int, tbl trace.Trace) *fr.Element {
	var clone fr.Element
	// Clone original value
	return clone.Set(e.Val)
}

func (e *AirAdd) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Add(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

func (e *AirSub) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Sub(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

func (e *AirMul) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Mul(l, r) }
	return EvalAirExprsAt(k, tbl, e.arguments, fn)
}

func (e *AirInverse) EvalAt(k int, tbl trace.Trace) *fr.Element {
	inv := new(fr.Element)
	val := e.expr.EvalAt(k, tbl)
	// Go syntax huh?
	return inv.Inverse(val)
}

// Evaluate all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalAirExprsAt(k int, tbl trace.Trace, exprs []AirExpr, fn func(*fr.Element, *fr.Element)) *fr.Element {
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
// Stringification
// ============================================================================

func (e *AirColumnAccess) String() string {
	if e.Shift() == 0 {
		return e.Column()
	} else {
		return fmt.Sprintf("(shift %s %d)",e.Column(),e.Shift())
	}
}

func (e *AirConstant) String() string {
	return e.Value().String()
}

func (e *AirAdd) String() string {
	return AirNaryString("+",e.arguments)
}

func (e *AirSub) String() string {
	return AirNaryString("-",e.arguments)
}

func (e *AirMul) String() string {
	return AirNaryString("*",e.arguments)
}

func (e *AirInverse) String() string {
	return fmt.Sprintf("(inv %s)",e.expr)
}

func AirNaryString(operator string, exprs []AirExpr) string {
	// This should be generalised and moved into common?
	rs := ""
	for _,e := range exprs {
		es := e.String()
		rs = fmt.Sprintf("%s %s",rs,es)
	}
	return fmt.Sprintf("(%s%s)",operator,rs)
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
	AddRecursiveListTranslator(&parser, "+", SExpAddToAir)
	AddRecursiveListTranslator(&parser, "-", SExpSubToAir)
	AddRecursiveListTranslator(&parser, "*", SExpMulToAir)
	AddBinaryListTranslator(&parser, "shift", SExpShiftToAir)
	// Parse string
	return Parse(parser, s)
}

func SExpConstantToAir(symbol string) (AirExpr, error) {
	num := new(fr.Element)
	// Attempt to parse
	c,err := num.SetString(symbol)
	// Check for errors
	if err != nil { return nil,err }
	// Done
	return &AirConstant{c},nil
}
func SExpColumnToAir(col string) (AirExpr, error) { return &AirColumnAccess{col,0},nil }
func SExpAddToAir(args []AirExpr) (AirExpr, error)   { return &AirAdd{args}, nil }
func SExpSubToAir(args []AirExpr) (AirExpr, error)   { return &AirSub{args}, nil }
func SExpMulToAir(args []AirExpr) (AirExpr, error)   { return &AirMul{args}, nil }

func SExpShiftToAir(col string, amt string) (AirExpr, error) {
	n,err1 := strconv.Atoi(amt)
	if err1 != nil { return nil,err1 }
	return &AirColumnAccess{col,n},nil
}
