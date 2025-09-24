package picus

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

//=======================
// Boolean Constraint AST
//=======================

type RelOp int

const (
	OpEq RelOp = iota // =
	OpNe              // !=
	OpLt              // <
	OpLe              // <=
	OpGt              // >
	OpGe              // >=
)

func (op RelOp) String() string {
	switch op {
	case OpEq:
		return "="
	case OpNe:
		return "!="
	case OpLt:
		return "<"
	case OpLe:
		return "<="
	case OpGt:
		return ">"
	case OpGe:
		return ">="
	default:
		panic(fmt.Sprintf("Unknown relational op: %d", op))
	}
}

// Interface defining a Picus formula. In PCL formulas appear in four places
// 1. (Assertions) - Circuit constraints (vanishing constraints here).
// 2. (Postconditions) - Formulas that Picus must prove are entailed by the module constraints
// 3. (Assumptions) - Formulas Picus assumes hold on the parameters of the module
// 3. (Conditional Guards)
type Formula[F field.Element[F]] interface {
	isFormula()
	Lisp() sexp.SExp
}

// Atomic predicate: (Left OP Right)
type Pred[F field.Element[F]] struct {
	Op    RelOp
	Left  Expr[F]
	Right Expr[F]
}

func (p *Pred[F]) Lisp() sexp.SExp {
	return sexp.NewList(
		[]sexp.SExp{
			sexp.NewSymbol(p.Op.String()),
			p.Left.Lisp(),
			p.Right.Lisp(),
		})
}

func (*Pred[F]) isFormula() {}

// Predicate constructor
func NewPred[F field.Element[F]](op RelOp, left Expr[F], right Expr[F]) *Pred[F] {
	return &Pred[F]{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

// Builds (< left right)
func NewLt[F field.Element[F]](left Expr[F], right Expr[F]) *Pred[F] {
	return NewPred[F](OpLt, left, right)
}

// Builds (> left right)
func NewGt[F field.Element[F]](left Expr[F], right Expr[F]) *Pred[F] {
	return NewPred[F](OpGt, left, right)
}

// Builds (<= left right)
func NewLeq[F field.Element[F]](left Expr[F], right Expr[F]) *Pred[F] {
	return NewPred[F](OpLe, left, right)
}

// Builds (>= left right)
func NewGeq[F field.Element[F]](left Expr[F], right Expr[F]) *Pred[F] {
	return NewPred[F](OpGe, left, right)
}

// Builds (= left right)
func NewEq[F field.Element[F]](left Expr[F], right Expr[F]) *Pred[F] {
	return NewPred[F](OpEq, left, right)
}

// Builds (!= left right)
func NewNeq[F field.Element[F]](left Expr[F], right Expr[F]) *Pred[F] {
	return NewPred[F](OpNe, left, right)
}

// FoldBinopPred constructs a left-associative Boolean formula by folding xs with op,
// i.e., (((x0 op x1) op x2) ...). It panics if fewer than two formulas are provided.
func FoldBinopPred[F field.Element[F]](op BinopConnective, xs []Formula[F]) Formula[F] {
	if len(xs) < 2 {
		panic(fmt.Sprintf("FoldAnd: expects at least two elements. Found %d", len(xs)))
	}
	acc := xs[0]
	for i := 1; i < len(xs); i++ {
		acc = &BinopConnectivePred[F]{Op: op, Left: acc, Right: xs[i]}
	}
	return acc
}

// Helper function to build an n-ary conjunction of formulas
func FoldAnd[F field.Element[F]](xs []Formula[F]) Formula[F] {
	return FoldBinopPred(OpAnd, xs)
}

func FoldOr[F field.Element[F]](xs []Formula[F]) Formula[F] {
	return FoldBinopPred(OpOr, xs)
}

type BinopConnective int

const (
	OpAnd BinopConnective = iota
	OpOr
	OpIff
	OpImplies
)

type BinopConnectivePred[F field.Element[F]] struct {
	Op    BinopConnective
	Left  Formula[F]
	Right Formula[F]
}

func (*BinopConnectivePred[F]) isFormula() {}

func (op BinopConnective) String() string {
	switch op {
	case OpAnd:
		return andSymbol
	case OpOr:
		return orSymbol
	case OpIff:
		return iffSymbol
	case OpImplies:
		return impliesSymbol
	default:
		panic(fmt.Sprintf("Unknown binop connective: %d", op))
	}
}

func (b *BinopConnectivePred[F]) Lisp() sexp.SExp {
	return sexp.NewList(
		[]sexp.SExp{
			sexp.NewSymbol(b.Op.String()),
			b.Left.Lisp(),
			b.Right.Lisp(),
		})
}

type Not[F field.Element[F]] struct{ X Formula[F] }

func (*Not[F]) isFormula() {}
func (n *Not[F]) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{sexp.NewSymbol(notSymbol), n.X.Lisp()})
}

func NewNot[F field.Element[F]](formula Formula[F]) *Not[F] {
	return &Not[F]{
		X: formula,
	}
}
