package binfile

import (
	"github.com/consensys/go-corset/pkg/mir"
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

type JsonExprColumn = any // for now

// =============================================================================
// Translation
// =============================================================================

// Convert a typed expression extracted from a JSON file into an
// expression in the Mid-Level Intermediate Representation.  This
// should not generate an error provided the original JSON was
// well-formed.
func (e JsonTypedExpr) ToMir() mir.Expr {
	if e.Expr.Funcall != nil {
		return e.Expr.Funcall.ToMir()
	} else if e.Expr.Const != nil {
		return e.Expr.Const.ToMir()
	} else {
		panic("Unknown JSON expression form encountered")
	}
}

func (e *JsonExprFuncall) ToMir() mir.Expr {
	switch e.Func {
	case "VectorSub":
		panic("VectorSub")
	}
	panic("Rest")
}

func (e *JsonExprConst) ToMir() mir.Expr {
	// one := new(fr.Element)
	// one.SetOne()
	// c := new(ir.MirConstant)
	// c.Val = one
	// return c
	panic("TO DO")
}
