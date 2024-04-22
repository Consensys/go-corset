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

type HirTable = trace.Table[HirConstraint]

// For now, all constraints are vanishing constraints.
type HirConstraint = *HirVanishingConstraint

// Lower (or refine) an HIR table into an MIR table.  That means
// lowering all the columns and constraints, whilst adding additional
// columns / constraints as necessary to preserve the original
// semantics.
func LowerToMir(hir HirTable, mir MirTable) {
	for _,col := range hir.Columns() {
		mir.AddColumn(col)
	}
	for _,c := range hir.Constraints() {
		// FIXME: this is broken because its currently
		// assuming that an AirConstraint is always a
		// VanishingConstraint.  Eventually this will not be
		// true.
		mir_exprs := c.Expr.LowerTo()
		// Add individual constraints arising
		for _,mir_expr := range mir_exprs {
			mir.AddConstraint(&trace.VanishingConstraint[MirExpr]{Handle: c.Handle,Expr: mir_expr})
		}
	}
}

// ============================================================================
// Constraints
// ============================================================================

type HirVanishingConstraint = trace.VanishingConstraint[HirExpr]

// ============================================================================
// Expressions
// ============================================================================

// An expression in the High-Level Intermediate Representation (HIR).
type HirExpr interface {
	// Lower this expression into the Mid-Level Intermediate
	// Representation.  Observe that a single expression at this
	// level can expand into *multiple* expressions at the MIR
	// level.
	LowerTo() []MirExpr
	// Evaluate this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, trace.Trace) *fr.Element
}

type HirAdd Add[HirExpr]
type HirSub Sub[HirExpr]
type HirMul Mul[HirExpr]
type HirConstant struct { Val *fr.Element }
type HirIfZero IfZero[HirExpr]
type HirList List[HirExpr]
type HirNormalise Normalise[HirExpr]
type HirColumnAccess struct { Col string; Amt int}

// HirConstant implements Constant interface
func (e *HirConstant) Value() *fr.Element { return e.Val }
// HirColumnAccess implements ColumnAccess interface
func (e *HirColumnAccess) Column() string { return e.Col }
func (e *HirColumnAccess) Shift() int { return e.Amt }

// ============================================================================
// Lowering
// ============================================================================

func (e *HirAdd) LowerTo() []MirExpr {
	return LowerWithNaryConstructor(e.arguments,func(nargs []MirExpr) MirExpr {
		return &MirAdd{nargs}
	})
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *HirConstant) LowerTo() []MirExpr {
	c := MirConstant{e.Val}
	return []MirExpr{&c}
}

func (e *HirColumnAccess) LowerTo() []MirExpr {
	return []MirExpr{&MirColumnAccess{e.Column(),e.Shift()}}
}

func (e *HirMul) LowerTo() []MirExpr {
	return LowerWithNaryConstructor(e.arguments,func(nargs []MirExpr) MirExpr {
		return &MirMul{nargs}
	})
}

func (e *HirNormalise) LowerTo() []MirExpr {
	mir_es := e.expr.LowerTo()
	for i,mir_e := range mir_es {
		mir_es[i] = &MirNormalise{mir_e}
	}
	return mir_es
}

// A list is lowered by eliminating it altogether.
func (e *HirList) LowerTo() []MirExpr {
	var res []MirExpr
	for i := 0; i < len(e.elements); i++ {
		// Lower ith argument
		iths := e.elements[i].LowerTo()
		// Append all as one
		res = append(res, iths...)
	}
	return res
}

func (e *HirIfZero) LowerTo() []MirExpr {
	var res []MirExpr
	// Lower required condition
	c := e.condition
	// Lower optional true branch
	t := e.trueBranch
	// Lower optional false branch
	f := e.falseBranch
	// Add constraints arising from true branch
	if t != nil {
		// (1 - NORM(x)) * y for true branch
		ts := LowerWithBinaryConstructor(c, t, func(x MirExpr, y MirExpr) MirExpr {
			one := new(fr.Element)
			one.SetOne()
			norm_x := &MirNormalise{x}
			one_minus_norm_x := &MirSub{[]MirExpr{&MirConstant{one},norm_x}}
			return &MirMul{[]MirExpr{one_minus_norm_x,y}}
		})
		res = append(res, ts...)
	}
	// Add constraints arising from false branch
	if f != nil {
		// x * y for false branch
		fs := LowerWithBinaryConstructor(c, f, func(x MirExpr, y MirExpr) MirExpr {
			return &MirMul{[]MirExpr{x,y}}
		})
		res = append(res, fs...)
	}
	// Done
	return res
}

func (e *HirSub) LowerTo() []MirExpr {
	return LowerWithNaryConstructor(e.arguments,func(nargs []MirExpr) MirExpr {
		return &MirSub{nargs}
	})
}

// ============================================================================
// Helpers
// ============================================================================

type BinaryConstructor func(MirExpr, MirExpr) MirExpr
type NaryConstructor func([]MirExpr) MirExpr

// A generic mechanism for lowering down to a binary expression.
func LowerWithBinaryConstructor(lhs HirExpr, rhs HirExpr, create BinaryConstructor) []MirExpr {
	var res []MirExpr
	// Lower all three expressions
	is := lhs.LowerTo()
	js := rhs.LowerTo()
	// Now construct
	for i := 0; i < len(is); i++ {
		for j := 0; j < len(js); j++ {
			// Construct binary expression
			expr := create(is[i], js[j])
			// Append to the end
			res = append(res, expr)
		}
	}
	return res
}

// Perform the cross-product expansion of an nary HIR expression.
// This is necessary because each argument of that expression will
// itself turn into one or more MIR expressions.  For example,
// consider lowering the following HIR expression:
//
// > (if X Y Z) + 10
//
// Here, (if X Y Z) will lower into two MIR expressions: (1-NORM(X))*Y
// and X*Z.  Thus, we need to generate two MIR expressions for our
// example:
//
// > ((1 - NORM(X)) * Y) + 10
// > (X * Y) + 10
//
// Finally, consider an expression such as the following:
//
// > (if X Y Z) + (if A B C)
//
// This will expand into *four* MIR expressions (i.e. the cross
// product of the left and right ifs).
func LowerWithNaryConstructor(args []HirExpr, constructor NaryConstructor) []MirExpr {
	// Accumulator is initially empty
	acc := make([]MirExpr,len(args))
	// Start from the first argument
	return LowerWithNaryConstructorHelper(0,acc,args,constructor)
}

// This manages progress through the cross-product expansion.
// Specifically, "i" determines how much of args has been lowered thus
// far, whilst "acc" represents the current array being generated.
func LowerWithNaryConstructorHelper(i int, acc []MirExpr, args []HirExpr, constructor NaryConstructor) []MirExpr {
	if i == len(acc) {
		// Base Case
		nacc := make([]MirExpr, len(acc))
		// Clone the slice because it is used as a temporary
		// working storage during the expansion.
		copy(nacc,acc)
		// Apply the constructor to produce the appropriate
		// MirExpr.
		return []MirExpr{constructor(nacc)}
	} else {
		// Recursive Case
		nargs := []MirExpr{}
		for _,ith := range args[i].LowerTo() {
			acc[i] = ith
			iths := LowerWithNaryConstructorHelper(i+1,acc,args,constructor)
			nargs = append(nargs,iths...)
		}
		return nargs
	}
}

// ============================================================================
// Evaluation
// ============================================================================

func (e *HirColumnAccess) EvalAt(k int, tbl trace.Trace) *fr.Element {
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

func (e *HirConstant) EvalAt(k int, tbl trace.Trace) *fr.Element {
	var clone fr.Element
	// Clone original value
	return clone.Set(e.Val)
}

func (e *HirAdd) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r*fr.Element) { l.Add(l,r) }
	return EvalHirExprsAt(k,tbl,e.arguments,fn)
}

func (e *HirMul) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r*fr.Element) { l.Mul(l,r) }
	return EvalHirExprsAt(k,tbl,e.arguments,fn)
}

func (e *HirIfZero) EvalAt(k int, tbl trace.Trace) *fr.Element {
	// Evaluate condition
	cond := e.condition.EvalAt(k,tbl)
	// Check whether zero or not
	if cond.IsZero() && e.trueBranch != nil {
		return e.trueBranch.EvalAt(k,tbl)
	} else if !cond.IsZero() && e.falseBranch != nil {
		return e.falseBranch.EvalAt(k,tbl)
	} else {
		// If either true / false branch undefined.
		return nil
	}
}

func (e *HirNormalise) EvalAt(k int, tbl trace.Trace) *fr.Element {
	// Check whether argument evaluates to zero or not.
	val := e.expr.EvalAt(k,tbl)
	// Normalise value (if necessary)
	if !val.IsZero() { val.SetOne() }
	// Done
	return val
}

func (e *HirSub) EvalAt(k int, tbl trace.Trace) *fr.Element {
	fn := func(l *fr.Element, r*fr.Element) { l.Sub(l,r) }
	return EvalHirExprsAt(k,tbl,e.arguments,fn)
}


// Evaluate all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalHirExprsAt(k int, tbl trace.Trace, exprs []HirExpr, fn func(*fr.Element,*fr.Element)) *fr.Element {
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

// Parse a string representing an HIR expression formatted using
// S-expressions.
func ParseSExpToHir(s string) (HirExpr,error) {
	parser := NewIrParser[HirExpr]()
	// Configure parser
	AddSymbolTranslator(&parser, SExpConstantToHir)
	AddSymbolTranslator(&parser, SExpColumnToHir)
	AddRecursiveListTranslator(&parser, "+", SExpAddToHir)
	AddRecursiveListTranslator(&parser, "-", SExpSubToHir)
	AddRecursiveListTranslator(&parser, "*", SExpMulToHir)
	AddBinaryListTranslator(&parser, "shift", SExpShiftToHir)
	AddRecursiveListTranslator(&parser, "~", SExpNormToHir)
	AddRecursiveListTranslator(&parser, "if", SExpIfToHir)
	// Parse string
	return Parse(parser,s)
}

func SExpConstantToHir(symbol string) (HirExpr,error) {
	num := new(fr.Element)
	// Attempt to parse
	c,err := num.SetString(symbol)
	// Check for errors
	if err != nil { return nil,err }
	// Done
	return &HirConstant{c},nil
}
func SExpColumnToHir(col string) (HirExpr,error) {
	return &HirColumnAccess{col,0},nil
}
func SExpAddToHir(args []HirExpr)(HirExpr,error) { return &HirAdd{args},nil }
func SExpSubToHir(args []HirExpr)(HirExpr,error) { return &HirSub{args},nil }
func SExpMulToHir(args []HirExpr)(HirExpr,error) { return &HirMul{args},nil }

func SExpIfToHir(args []HirExpr)(HirExpr,error) {
	if len(args) == 2 {
		return &HirIfZero{args[0],args[1],nil},nil
	} else if len(args) == 3 {
		return &HirIfZero{args[0],args[1],args[2]},nil
	} else {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}",len(args))
		return nil, errors.New(msg)
	}
}

func SExpShiftToHir(col string, amt string) (HirExpr,error) {
	n,err1 := strconv.Atoi(amt)
	if err1 != nil { return nil,err1 }
	return &HirColumnAccess{col,n},nil
}

func SExpNormToHir(args []HirExpr) (HirExpr,error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}",len(args))
		return nil, errors.New(msg)
	} else {
		return &HirNormalise{args[0]}, nil
	}
}
