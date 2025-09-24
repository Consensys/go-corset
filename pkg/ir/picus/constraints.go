package picus

import (
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constraint is a recursive Boolean formula over arithmetic predicates.
type Constraint[F field.Element[F]] interface {
	isConstraint()
	Lisp() sexp.SExp
}

type PicusConstraint[F field.Element[F]] struct {
	Formula Formula[F]
}

func (*PicusConstraint[F]) isConstraint() {}

func (v *PicusConstraint[F]) Lisp() sexp.SExp {
	return sexp.NewList(
		[]sexp.SExp{
			sexp.NewSymbol("assert"),
			v.Formula.Lisp(),
		})
}

func NewPicusConstraint[F field.Element[F]](formula Formula[F]) *PicusConstraint[F] {
	return &PicusConstraint[F]{
		Formula: formula,
	}
}

type ConstraintBlock[F field.Element[F]] []Constraint[F]

func (cb *ConstraintBlock[F]) Lisp() sexp.SExp {
	elements := make([]sexp.SExp, len(*cb))
	for i, constraint := range *cb {
		elements[i] = constraint.Lisp()
	}
	return sexp.NewList(elements)
}

type IfElse[F field.Element[F]] struct {
	cond        Formula[F]
	trueBranch  ConstraintBlock[F]
	falseBranch ConstraintBlock[F]
}

func (*IfElse[F]) isConstraint() {}

func (ite *IfElse[F]) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		ite.cond.Lisp(),
		ite.trueBranch.Lisp(),
		ite.falseBranch.Lisp(),
	})
}

func NewIfElse[F field.Element[F]](cond Formula[F], thenConstraints []Constraint[F], elseConstraints []Constraint[F]) *IfElse[F] {
	return &IfElse[F]{
		cond:        cond,
		trueBranch:  ConstraintBlock[F](thenConstraints),
		falseBranch: ConstraintBlock[F](elseConstraints),
	}
}
