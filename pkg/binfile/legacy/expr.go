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

func (e *jsonTypedExpr) ToHir(colmap map[uint]uint, schema *hir.Schema) hir.Expr {
	if e.Expr.Column != nil {
		return e.Expr.Column.ToHir(colmap, schema)
	} else if e.Expr.Const != nil {
		return e.Expr.Const.ToHir(schema)
	} else if e.Expr.Funcall != nil {
		return e.Expr.Funcall.ToHir(colmap, schema)
	} else if e.Expr.List != nil {
		// Parse the arguments
		return jsonListToHir(e.Expr.List, colmap, schema)
	}

	panic("Unknown JSON expression encountered")
}

// ToHir converts a big integer represented as a sequence of unsigned 32bit
// words into HIR constant expression.
func (e *jsonExprConst) ToHir(schema *hir.Schema) hir.Expr {
	return hir.NewConst(e.ToField())
}

func (e *jsonExprConst) ToField() fr.Element {
	var num fr.Element
	//
	val := e.ToBigInt()
	// Construct Field Value
	num.SetBigInt(val)
	//
	return num
}

func (e *jsonExprConst) ToBigInt() *big.Int {
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
	// Done
	return val
}

func (e *jsonExprColumn) ToHir(colmap map[uint]uint, schema *hir.Schema) hir.Expr {
	// Determine binfile column index
	cid := asColumn(e.Handle)
	// Map to schema column index
	return hir.NewColumnAccess(colmap[cid], e.Shift)
}

func (e *jsonExprFuncall) ToHir(colmap map[uint]uint, schema *hir.Schema) hir.Expr {
	// Parse the arguments
	args := make([]hir.Expr, len(e.Args))
	for i := 0; i < len(e.Args); i++ {
		args[i] = e.Args[i].ToHir(colmap, schema)
	}
	// Construct appropriate expression
	switch e.Func {
	case "Normalize":
		if len(args) == 1 {
			return hir.Normalise(args[0])
		} else {
			panic("incorrect arguments for Normalize")
		}
	case "VectorAdd", "Add":
		return hir.Sum(args...)
	case "VectorMul", "Mul":
		return hir.Product(args...)
	case "VectorSub", "Sub":
		return hir.Subtract(args...)
	case "Exp":
		if len(args) != 2 {
			panic(fmt.Sprintf("incorrect number of arguments for Exp (%d)", len(args)))
		}

		c, ok := args[1].Term.(*hir.Constant)

		if !ok {
			panic(fmt.Sprintf("constant power expected for Exp, got %s", args[1].Lisp(schema)))
		} else if !c.Value.IsUint64() {
			panic("constant power too large for Exp")
		}

		var k big.Int
		// Convert power to uint64
		c.Value.BigInt(&k)
		// Done
		return hir.Exponent(args[0], k.Uint64())
	case "IfZero":
		if len(args) == 2 {
			return hir.If(args[0], args[1], hir.ZERO)
		} else if len(args) == 3 {
			return hir.If(args[0], args[1], args[2])
		} else {
			panic(fmt.Sprintf("incorrect number of arguments for IfZero (%d)", len(args)))
		}
	case "IfNotZero":
		if len(args) == 2 {
			return hir.If(args[0], hir.ZERO, args[1])
		} else if len(args) == 3 {
			return hir.If(args[0], args[2], args[1])
		} else {
			panic(fmt.Sprintf("incorrect number of arguments for IfNotZero (%d)", len(args)))
		}
	}
	// Catch anything we've missed
	panic(fmt.Sprintf("HANDLE %s\n", e.Func))
}

func jsonListToHir(Args []jsonTypedExpr, colmap map[uint]uint, schema *hir.Schema) hir.Expr {
	args := make([]hir.Expr, len(Args))
	for i := 0; i < len(Args); i++ {
		args[i] = Args[i].ToHir(colmap, schema)
	}

	return hir.ListOf(args...)
}

func jsonExprsToHirUnit(Args []jsonTypedExpr, colmap map[uint]uint, schema *hir.Schema) []hir.UnitExpr {
	args := make([]hir.UnitExpr, len(Args))
	for i := 0; i < len(Args); i++ {
		args[i] = hir.NewUnitExpr(Args[i].ToHir(colmap, schema))
	}

	return args
}
