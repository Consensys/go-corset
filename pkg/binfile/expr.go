package binfile

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/hir"
)

// type jsonHandle struct {
// 	H  string `json:"h"`
// 	ID int    `json:"id"`
// }

// jsonColumnRef corresponds to a column reference.
type jsonColumnRef = string

// jsonTypedExpr corresponds to an optionally typed expression.
type jsonTypedExpr struct {
	Expr jsonExpr `json:"_e"`
}

// jsonExpr is an enumeration of expression forms.  Exactly one of these fields
// must be non-nil.
type jsonExpr struct {
	Funcall *jsonExprFuncall
	Const   *jsonExprConst
	Column  *jsonExprColumn
	List    []jsonTypedExpr
}

// jsonExprFuncall corresponds to an (intrinsic) function call with zero or more
// arguments.
type jsonExprFuncall struct {
	Func string          `json:"func"`
	Args []jsonTypedExpr `json:"args"`
}

// jsonExprConst corresponds to an (unbound) integer constant in the expression
// tree.
type jsonExprConst struct {
	BigInt []any
}

type jsonExprColumn struct {
	Handle    jsonColumnRef `json:"handle"`
	Shift     int           `json:"shift"`
	MustProve bool          `json:"must_prove"`
}

// =============================================================================
// Translation
// =============================================================================

// ToMir converts a typed expression extracted from a JSON file into an
// expression in the Mid-Level Intermediate Representation.  This
// should not generate an error provided the original JSON was
// well-formed.

func (e *jsonTypedExpr) ToHir(schema *hir.Schema) hir.Expr {
	if e.Expr.Column != nil {
		return e.Expr.Column.ToHir(schema)
	} else if e.Expr.Const != nil {
		return e.Expr.Const.ToHir(schema)
	} else if e.Expr.Funcall != nil {
		return e.Expr.Funcall.ToHir(schema)
	} else if e.Expr.List != nil {
		// Parse the arguments
		return jsonListToHir(e.Expr.List, schema)
	}

	panic("Unknown JSON expression encountered")
}

// ToHir converts a big integer represented as a sequence of unsigned 32bit
// words into HIR constant expression.
func (e *jsonExprConst) ToHir(schema *hir.Schema) hir.Expr {
	sign := int(e.BigInt[0].(float64))
	words := e.BigInt[1].([]any)
	// Begin
	val := big.NewInt(0)
	base := big.NewInt(1)
	// Construct 2^32 = 4294967296
	var two32, n = big.NewInt(2), big.NewInt(32)

	two32.Exp(two32, n, nil)
	// Iterate the words
	for _, w := range words {
		word := big.NewInt(int64(w.(float64)))
		word = word.Mul(word, base)
		val = val.Add(val, word)
		base = base.Mul(base, two32)
	}
	// Apply Sign
	if sign == 1 || sign == 0 {
		// do nothing
	} else if sign == -1 {
		val = val.Neg(val)
	} else {
		panic(fmt.Sprintf("Unknown BigInt sign: %d", sign))
	}
	// Construct Field Value
	num := new(fr.Element)
	num.SetBigInt(val)

	// Done!
	return &hir.Constant{Val: num}
}

func (e *jsonExprColumn) ToHir(schema *hir.Schema) hir.Expr {
	cref := asColumnRef(e.Handle)
	_, cid := cref.resolve(schema)

	return &hir.ColumnAccess{Column: cid, Shift: e.Shift}
}

func (e *jsonExprFuncall) ToHir(schema *hir.Schema) hir.Expr {
	// Parse the arguments
	args := make([]hir.Expr, len(e.Args))
	for i := 0; i < len(e.Args); i++ {
		args[i] = e.Args[i].ToHir(schema)
	}
	// Construct appropriate expression
	switch e.Func {
	case "Normalize":
		if len(args) == 1 {
			return &hir.Normalise{Arg: args[0]}
		} else {
			panic("incorrect arguments for Normalize")
		}
	case "VectorAdd", "Add":
		return &hir.Add{Args: args}
	case "VectorMul", "Mul":
		return &hir.Mul{Args: args}
	case "VectorSub", "Sub":
		return &hir.Sub{Args: args}
	case "IfZero":
		if len(args) == 2 {
			return &hir.IfZero{Condition: args[0], TrueBranch: args[1], FalseBranch: nil}
		} else if len(args) == 3 {
			return &hir.IfZero{Condition: args[0], TrueBranch: args[1], FalseBranch: args[2]}
		} else {
			panic(fmt.Sprintf("incorrect number of arguments for IfZero (%d)", len(args)))
		}
	case "IfNotZero":
		if len(args) == 2 {
			return &hir.IfZero{Condition: args[0], TrueBranch: nil, FalseBranch: args[1]}
		} else if len(args) == 3 {
			return &hir.IfZero{Condition: args[0], TrueBranch: args[2], FalseBranch: args[1]}
		} else {
			panic(fmt.Sprintf("incorrect number of arguments for IfNotZero (%d)", len(args)))
		}
	}
	// Catch anything we've missed
	panic(fmt.Sprintf("HANDLE %s\n", e.Func))
}

func jsonListToHir(Args []jsonTypedExpr, schema *hir.Schema) hir.Expr {
	args := make([]hir.Expr, len(Args))
	for i := 0; i < len(Args); i++ {
		args[i] = Args[i].ToHir(schema)
	}

	return &hir.List{Args: args}
}

func jsonExprsToHirUnit(Args []jsonTypedExpr, schema *hir.Schema) []hir.UnitExpr {
	args := make([]hir.UnitExpr, len(Args))
	for i := 0; i < len(Args); i++ {
		args[i] = hir.NewUnitExpr(Args[i].ToHir(schema))
	}

	return args
}
