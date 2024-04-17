package ir

import (
	"errors"
	"fmt"
	"math/big"
	"github.com/Consensys/go-corset/pkg/trace"
)

// ============================================================================
// Table
// ============================================================================

type MirTable = trace.Table[MirConstraint]

// For now, all constraints are vanishing constraints.
type MirConstraint = trace.VanishingConstraint[MirExpr]

// Lower (or refine) an MIR table into an AIR table.  That means
// lowering all the columns and constraints, whilst adding additional
// columns / constraints as necessary to preserve the original
// semantics.
func LowerTo(from MirTable, to AirTable) {
	panic("GOT HERE")
}

// ============================================================================
// Expressions
// ============================================================================

// An MirExpression in the Mid-Level Intermediate Representation (MIR).
type MirExpr interface {
	// Lower this MirExpression into the Arithmetic Intermediate
	// Representation.  Essentially, this means eliminating
	// normalising expressions by introducing new columns into the
	// given table (with appropriate constraints).
	LowerTo(AirTable) AirExpr
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, trace.Trace) *big.Int
}

type MirAdd Add[MirExpr]
type MirSub Sub[MirExpr]
type MirMul Mul[MirExpr]
type MirConstant struct { Val *big.Int }
type MirNormalise Normalise[MirExpr]
type MirColumnAccess struct { Col string; Amt int}

// MirConstant implements Constant interface
func (e *MirConstant) Value() *big.Int { return e.Val }
// MirColumnAccess implements ColumnAccess interface
func (e *MirColumnAccess) Column() string { return e.Col }
func (e *MirColumnAccess) Shift() int { return e.Amt }

// ============================================================================
// Lowering
// ============================================================================

func (e *MirAdd) LowerTo(tbl AirTable) AirExpr {
	return &AirAdd{LowerMirExprs(e.arguments,tbl)}
}

func (e *MirSub) LowerTo(tbl AirTable) AirExpr {
	return &AirSub{LowerMirExprs(e.arguments,tbl)}
}

func (e *MirMul) LowerTo(tbl AirTable) AirExpr {
	return &AirMul{LowerMirExprs(e.arguments,tbl)}
}

func (e *MirNormalise) LowerTo(tbl AirTable) AirExpr {
	panic("Implement MirNormalise.LowerTo()!")
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirColumnAccess) LowerTo(tbl AirTable) AirExpr {
	return &AirColumnAccess{e.Column(),e.Shift()}
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirConstant) LowerTo(tbl AirTable) AirExpr {
	return e
}

// Lower a set of zero or more MIR expressions.
func LowerMirExprs(exprs []MirExpr,tbl AirTable) []AirExpr {
	n := len(exprs)
	nexprs := make([]AirExpr, n)
	for i := 0; i < n; i++ {
		nexprs[i] = exprs[i].LowerTo(tbl)
	}
	return nexprs
}

// ============================================================================
// Constraints
// ============================================================================

type MirVanishingConstraint = trace.VanishingConstraint[MirExpr]

// ============================================================================
// Evaluation
// ============================================================================

func (e *MirColumnAccess) EvalAt(k int, tbl trace.Trace) *big.Int {
	val, _ := tbl.GetByName(e.Column(), k+e.Shift())
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

func (e *MirConstant) EvalAt(k int, tbl trace.Trace) *big.Int {
	var clone big.Int
	// Clone original value
	return clone.Set(e.Val)
}

func (e *MirAdd) EvalAt(k int, tbl trace.Trace) *big.Int {
	fn := func(l *big.Int, r*big.Int) { l.Add(l,r) }
	return EvalMirExprsAt(k,tbl,e.arguments,fn)
}

func (e *MirMul) EvalAt(k int, tbl trace.Trace) *big.Int {
	fn := func(l *big.Int, r*big.Int) { l.Mul(l,r) }
	return EvalMirExprsAt(k,tbl,e.arguments,fn)
}

func (e *MirNormalise) EvalAt(k int, tbl trace.Trace) *big.Int {
	// Check whether argument evaluates to zero or not.
	if e.expr.EvalAt(k,tbl).BitLen() == 0 {
		return big.NewInt(0)
	} else {
		return big.NewInt(1)
	}
}

func (e *MirSub) EvalAt(k int, tbl trace.Trace) *big.Int {
	fn := func(l *big.Int, r*big.Int) { l.Sub(l,r) }
	return EvalMirExprsAt(k,tbl,e.arguments,fn)
}


// Evaluate all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalMirExprsAt(k int, tbl trace.Trace, exprs []MirExpr, fn func(*big.Int,*big.Int)) *big.Int {
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

// ============================================================================
// Parser
// ============================================================================

// Parse a string representing an MIR expression formatted using
// S-expressions.
func ParseSExpToMir(s string) (MirExpr,error) {
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
	return Parse(parser,s)
}

func SExpConstantToMir(symbol string) (MirExpr,error) {
	c,err := StringToConstant(symbol)
	if err != nil { return nil,err }
	return &MirConstant{c},nil
}
func SExpColumnToMir(symbol string) (MirExpr,error) {
	c,n,err := StringToColumnAccess(symbol)
	if err != nil { return nil,err }
	return &MirColumnAccess{c,n},nil
}
func SExpAddToMir(args []MirExpr)(MirExpr,error) { return &MirAdd{args},nil }
func SExpSubToMir(args []MirExpr)(MirExpr,error) { return &MirSub{args},nil }
func SExpMulToMir(args []MirExpr)(MirExpr,error) { return &MirMul{args},nil }
func SExpShiftToMir(args []MirExpr) (MirExpr,error) {
	c,n,err := SliceToShiftAccess(args)
	if err != nil { return nil,err }
	return &MirColumnAccess{c,n},nil
}

func SExpNormToMir(args []MirExpr) (MirExpr,error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of shift arguments: {%d}",len(args))
		return nil, errors.New(msg)
	} else {
		return &MirNormalise{args[0]}, nil
	}
}
