package pcl

import (
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// A Picus constraint
type Constraint[F field.Element[F]] interface {
	isConstraint()
	Lisp() sexp.SExp // PCL is similar to smtlib which is LISP-like
}

// Represents a Picus `assert` statement which asserts boolean combinations polynomial (in)-equalities
type Assert[F field.Element[F]] struct {
	Formula Formula[F]
}

// Implements the `Constraint` interface
func (*Assert[F]) isConstraint() {}

// Creates a constraint `(assert (= x 1))`
func (v *Assert[F]) Lisp() sexp.SExp {
	return sexp.NewList(
		[]sexp.SExp{
			sexp.NewSymbol("assert"),
			v.Formula.Lisp(),
		})
}

func NewPicusConstraint[F field.Element[F]](formula Formula[F]) *Assert[F] {
	return &Assert[F]{
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

// Defines a PCL if-then-else statement
type IfElse[F field.Element[F]] struct {
	cond        Formula[F]
	trueBranch  ConstraintBlock[F]
	falseBranch ConstraintBlock[F]
}

// Implements constraint interface
func (*IfElse[F]) isConstraint() {}

func (ite *IfElse[F]) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		ite.cond.Lisp(),
		ite.trueBranch.Lisp(),
		ite.falseBranch.Lisp(),
	})
}

// Constructs an ite
func NewIfElse[F field.Element[F]](cond Formula[F], thenConstraints []Constraint[F], elseConstraints []Constraint[F]) *IfElse[F] {
	return &IfElse[F]{
		cond:        cond,
		trueBranch:  ConstraintBlock[F](thenConstraints),
		falseBranch: ConstraintBlock[F](elseConstraints),
	}
}
