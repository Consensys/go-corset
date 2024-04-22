package ir

import (
	"errors"
	"fmt"
	"strconv"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ============================================================================
// Table
// ============================================================================

type MirTable = trace.Table[MirConstraint]

// For now, all constraints are vanishing constraints.
type MirConstraint = *trace.VanishingConstraint[MirExpr]

// Lower (or refine) an MIR table into an AIR table.  That means
// lowering all the columns and constraints, whilst adding additional
// columns / constraints as necessary to preserve the original
// semantics.
func LowerToAir(mir MirTable, air AirTable) {
	for _,col := range mir.Columns() {
		air.AddColumn(col)
	}
	for _,c := range mir.Constraints() {
		// FIXME: this is broken because its currently
		// assuming that an AirConstraint is always a
		// VanishingConstraint.  Eventually this will not be
		// true.
		air_expr := c.Expr.LowerTo(air)
		air.AddConstraint(&trace.VanishingConstraint[AirExpr]{Handle: c.Handle,Expr: air_expr})
	}
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
	EvalAt(int, trace.Trace) *fr.Element
}

type MirAdd Add[MirExpr]
type MirSub Sub[MirExpr]
type MirMul Mul[MirExpr]
type MirConstant struct { Val *fr.Element }
type MirNormalise Normalise[MirExpr]
type MirColumnAccess struct { Col string; Amt int}

// MirConstant implements Constant interface
func (e *MirConstant) Value() *fr.Element { return e.Val }
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

func (p *MirNormalise) LowerTo(tbl AirTable) AirExpr {
	// Lower the expression being normalised
	e := p.expr.LowerTo(tbl)
	// Invert expression
	ie := &AirInverse{e}
	// Determine computed column name
	name := ie.String()
	// Add new column (if it does not already exist)
	if !tbl.HasColumn(name) {
		// Add computed column
		tbl.AddColumn(trace.NewComputedColumn(name,ie))
	}
	one := fr.NewElement(1)
	// Construct 1/e
	inv_e := &AirColumnAccess{name,0}
	// Construct e/e
	e_inv_e := &AirMul{[]AirExpr{e,inv_e}}
	// Construct 1 == e/e
	one_e_e := &AirSub{[]AirExpr{&AirConstant{&one},e_inv_e}}
	// Construct (e != 0) ==> (1 == e/e)
	e_implies_one_e_e := &AirMul{[]AirExpr{e,one_e_e}}
	// Construct (1/e != 0) ==> (1 == e/e)
	inv_e_implies_one_e_e := &AirMul{[]AirExpr{inv_e,one_e_e}}
	// Ensure (e != 0) ==> (1 == e/e)
	l_name := fmt.Sprintf("[%s <=]",ie.String())
	tbl.AddConstraint(&trace.VanishingConstraint[AirExpr]{Handle: l_name, Expr: e_implies_one_e_e})
	// Ensure (e/e != 0) ==> (1 == e/e)
	r_name := fmt.Sprintf("[%s =>]",ie.String())
	tbl.AddConstraint(&trace.VanishingConstraint[AirExpr]{Handle: r_name, Expr: inv_e_implies_one_e_e})
	// Done
	return &AirMul{[]AirExpr{e,&AirColumnAccess{name,0}}}
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirColumnAccess) LowerTo(tbl AirTable) AirExpr {
	return &AirColumnAccess{e.Column(),e.Shift()}
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirConstant) LowerTo(tbl AirTable) AirExpr {
	return &AirConstant{e.Value()}
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

func (e *MirColumnAccess) EvalAt(k int, tbl trace.Trace) *fr.Element {
	val, _ := tbl.GetByName(e.Column(), k+e.Shift())
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

func (e *MirConstant) EvalAt(k int, tbl trace.Trace) *fr.Element {
	var clone fr.Element
	// Clone original value
	return clone.Set(e.Val)
}

func (e *MirAdd) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r*fr.Element) { l.Add(l,r) }
	return EvalMirExprsAt(k,tbl,e.arguments,fn)
}

func (e *MirMul) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r*fr.Element) { l.Mul(l,r) }
	return EvalMirExprsAt(k,tbl,e.arguments,fn)
}

func (e *MirNormalise) EvalAt(k int, tbl trace.Trace) *fr.Element {
	// Check whether argument evaluates to zero or not.
	val := e.expr.EvalAt(k,tbl)
	// Normalise value (if necessary)
	if !val.IsZero() { val.SetOne() }
	// Done
	return val
}

func (e *MirSub) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r*fr.Element) { l.Sub(l,r) }
	return EvalMirExprsAt(k,tbl,e.arguments,fn)
}


// Evaluate all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalMirExprsAt(k int, tbl trace.Trace, exprs []MirExpr, fn func(*fr.Element,*fr.Element)) *fr.Element {
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
	AddRecursiveListTranslator(&parser, "+", SExpAddToMir)
	AddRecursiveListTranslator(&parser, "-", SExpSubToMir)
	AddRecursiveListTranslator(&parser, "*", SExpMulToMir)
	AddBinaryListTranslator(&parser, "shift", SExpShiftToMir)
	AddRecursiveListTranslator(&parser, "~", SExpNormToMir)
	// Parse string
	return Parse(parser,s)
}

func SExpConstantToMir(symbol string) (MirExpr,error) {
	num := new(fr.Element)
	// Attempt to parse
	c,err := num.SetString(symbol)
	// Check for errors
	if err != nil { return nil,err }
	// Done
	return &MirConstant{c},nil
}
func SExpColumnToMir(col string) (MirExpr,error) {
	return &MirColumnAccess{col,0},nil
}
func SExpAddToMir(args []MirExpr)(MirExpr,error) { return &MirAdd{args},nil }
func SExpSubToMir(args []MirExpr)(MirExpr,error) { return &MirSub{args},nil }
func SExpMulToMir(args []MirExpr)(MirExpr,error) { return &MirMul{args},nil }
func SExpShiftToMir(col string, amt string) (MirExpr,error) {
	n,err1 := strconv.Atoi(amt)
	if err1 != nil { return nil,err1 }
	return &MirColumnAccess{col,n},nil
}

func SExpNormToMir(args []MirExpr) (MirExpr,error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of shift arguments: {%d}",len(args))
		return nil, errors.New(msg)
	} else {
		return &MirNormalise{args[0]}, nil
	}
}
