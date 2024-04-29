package binfile

import (
	"fmt"
	"math/big"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Corresponds to an optionally typed expression.
type JsonTypedExpr struct {
	Expr JsonExpr `json:"_e"`
}

// An enumeration of expression forms.  Exactly one of these fields
// must be non-nil.
type JsonExpr struct {
	Funcall *JsonExprFuncall
	Const   *JsonExprConst
	Column  *JsonExprColumn
}

// Corresponds to an (intrinsic) function call with zero or more
// arguments.
type JsonExprFuncall struct {
	Func string          `json:"func"`
	Args []JsonTypedExpr `json:"args"`
}

// Corresponds to an (unbound) integer constant in the expression
// tree.
type JsonExprConst struct {
	BigInt []any
}

type JsonExprColumn struct {
	Handle Handle `json:"handle"`
	Shift int `json:"shift"`
	MustProve bool `json:"must_prove"`
}

// =============================================================================
// Translation
// =============================================================================

// Convert a typed expression extracted from a JSON file into an
// expression in the Mid-Level Intermediate Representation.  This
// should not generate an error provided the original JSON was
// well-formed.
func (e JsonTypedExpr) ToHir() hir.Expr {
	if e.Expr.Funcall != nil {
		return e.Expr.Funcall.ToHir()
	} else if e.Expr.Const != nil {
		return e.Expr.Const.ToHir()
	} else if e.Expr.Column != nil {
		return e.Expr.Column.ToHir()
	} else {
		panic("Unknown JSON expression encountered")
	}
}

func (e *JsonExprFuncall) ToHir() hir.Expr {
	// Parse the arguments
	args := make([]hir.Expr,len(e.Args))
	for i := 0; i <len(e.Args); i++ {
		args[i] = e.Args[i].ToHir()
	}
	// Construct appropriate expression
	switch e.Func {
	case "VectorAdd","Add":
		return &hir.Add{Args: args}
	case "VectorMul","Mul":
		return &hir.Mul{Args: args}
	case "VectorSub","Sub":
		return &hir.Sub{Args: args}
	case "IfZero":
		if len(args) == 2 {
			return &hir.IfZero{Condition: args[0], TrueBranch: args[1], FalseBranch: nil}
		} else if len(args) == 3 {
			return &hir.IfZero{Condition: args[0], TrueBranch: args[1], FalseBranch: args[2]}
		} else { panic("incorrect arguments for IfZero") }
	case "IfNotZero":
		if len(args) == 2 {
			return &hir.IfZero{Condition: args[0], TrueBranch: nil, FalseBranch: args[1]}
		} else { panic("incorrect arguments for IfZero") }
	}
	// Catch anything we've missed
	panic(fmt.Sprintf("HANDLE %s\n",e.Func))
}

func (e *JsonExprColumn) ToHir() hir.Expr {
	return &hir.ColumnAccess{Column: e.Handle.H, Shift: e.Shift}
}

// Convert a big integer represented as a sequence of unsigned 32bit
// words into HIR constant expression.
func (e *JsonExprConst) ToHir() hir.Expr {
	sign := int(e.BigInt[0].(float64))
	words := e.BigInt[1].([]any)
	// Begin
	val := big.NewInt(0)
	base := big.NewInt(1)
	// Construct 2^32 = 4294967296
	var two_32, n = big.NewInt(2), big.NewInt(32)
	two_32.Exp(two_32, n, nil)
	// Iterate the words
	for _,w := range words {
		word := big.NewInt(int64(w.(float64)))
		word = word.Mul(word,base)
		val = val.Add(val,word)
		base = base.Mul(base,two_32)
	}
	// Apply Sign
	if sign == 1 || sign == 0 {
		// do nothing
	} else if sign == -1 {
		val = val.Neg(val)
	} else { panic(fmt.Sprintf("Unknown BigInt sign: %d",sign)) }
	// Construct Field Value
	num := new(fr.Element)
	num.SetBigInt(val)
	// Done!
	return &hir.Constant{Val: num}
}
