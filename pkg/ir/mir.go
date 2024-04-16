package ir

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/Consensys/go-corset/pkg/trace"
)

// MirExpr is a MirExpression in the Mid-Level Intermediate Representation (MIR).
type MirExpr interface {
	// LowerToAir lowers this MirExpression into the Arithmetic Intermediate
	// Representation.  Essentially, this means eliminating normalising
	// expressions by introducing new columns into the enclosing table (with
	// appropriate constraints).
	LowerToAir() AirExpr

	// EvalAt evaluates this expression in a given tabular context.
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
type MirColumnAccess = ColumnAccess

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
	panic("Implement MirNormalise.LowerToAir()!")
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirColumnAccess) LowerToAir() AirExpr {
	return e
}

// LowerToAir lowering a constant is straightforward as it is already in the correct form.
func (e *MirConstant) LowerToAir() AirExpr {
	return e
}

// LowerMirExprs lowers a set of zero or more MIR expressions.
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
	fn := func(l *big.Int, r *big.Int) { l.Add(l, r) }
	return EvalMirExprsAt(k, tbl, e.arguments, fn)
}

func (e *MirSub) EvalAt(k int, tbl trace.Table) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Sub(l, r) }
	return EvalMirExprsAt(k, tbl, e.arguments, fn)
}

func (e *MirMul) EvalAt(k int, tbl trace.Table) *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Mul(l, r) }
	return EvalMirExprsAt(k, tbl, e.arguments, fn)
}

func (e *MirNormalise) EvalAt(k int, tbl trace.Table) *big.Int {
	// Check whether argument evaluates to zero or not.
	if e.expr.EvalAt(k, tbl).BitLen() == 0 {
		return big.NewInt(0)
	} else {
		return big.NewInt(1)
	}
}

// EvalMirExprsAt evaluates all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalMirExprsAt(k int, tbl trace.Table, exprs []MirExpr, fn func(*big.Int, *big.Int)) *big.Int {
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

// Parse a string representing an MIR expression formatted using
// S-expressions.
func ParseSExpToMir(s string) (MirExpr, error) {
	parser := NewIrParser[MirExpr]()
	// Configure parser
	AddSymbolTranslator(&parser, SExpConstantToMir)
	AddSymbolTranslator(&parser, SExpColumnToMir)
	AddListTranslator(&parser, "+", SExpAddToMir)
	AddListTranslator(&parser, "-", SExpSubToMir)
	AddListTranslator(&parser, "*", SExpMulToMir)
	AddListTranslator(&parser, "shift", SExpShiftToMir)
	AddListTranslator(&parser, "norm", SExpNormToMir)
	// Parse string
	return Parse(parser, s)
}

func SExpConstantToMir(symbol string) (MirExpr, error) { return StringToConstant(symbol) }
func SExpColumnToMir(symbol string) (MirExpr, error)   { return StringToColumnAccess(symbol) }
func SExpAddToMir(args []MirExpr) (MirExpr, error)     { return &MirAdd{args}, nil }
func SExpSubToMir(args []MirExpr) (MirExpr, error)     { return &MirSub{args}, nil }
func SExpMulToMir(args []MirExpr) (MirExpr, error)     { return &MirMul{args}, nil }
func SExpShiftToMir(args []MirExpr) (MirExpr, error)   { return SliceToShiftAccess(args) }

func SExpNormToMir(args []MirExpr) (MirExpr, error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of shift arguments: {%d}", len(args))
		return nil, errors.New(msg)
	} else {
		return &MirNormalise{args[0]}, nil
	}
}
