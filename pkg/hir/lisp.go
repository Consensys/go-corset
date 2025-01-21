package hir

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *ColumnAccess) Lisp(schema sc.Schema) sexp.SExp {
	name := schema.Columns().Nth(e.Column).QualifiedName(schema)
	access := sexp.NewSymbol(name)
	// Check whether shifted (or not)
	if e.Shift == 0 {
		// Not shifted
		return access
	}
	// Shifted
	shift := sexp.NewSymbol(fmt.Sprintf("%d", e.Shift))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("shift"), access, shift})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Constant) Lisp(schema sc.Schema) sexp.SExp {
	return sexp.NewSymbol(e.Val.String())
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *List) Lisp(schema sc.Schema) sexp.SExp {
	return nary2Lisp(schema, "begin", e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Add) Lisp(schema sc.Schema) sexp.SExp {
	return nary2Lisp(schema, "+", e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Sub) Lisp(schema sc.Schema) sexp.SExp {
	return nary2Lisp(schema, "-", e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Mul) Lisp(schema sc.Schema) sexp.SExp {
	return nary2Lisp(schema, "*", e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Normalise) Lisp(schema sc.Schema) sexp.SExp {
	arg := e.Arg.Lisp(schema)
	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("~"), arg})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Exp) Lisp(schema sc.Schema) sexp.SExp {
	arg := e.Arg.Lisp(schema)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", e.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *IfZero) Lisp(schema sc.Schema) sexp.SExp {
	// Translate Condition
	condition := e.Condition.Lisp(schema)
	// Dispatch on type
	if e.FalseBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("if"),
			condition,
			e.TrueBranch.Lisp(schema),
		})
	} else if e.TrueBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("ifnot"),
			condition,
			e.FalseBranch.Lisp(schema),
		})
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		condition,
		e.TrueBranch.Lisp(schema),
		e.FalseBranch.Lisp(schema),
	})
}

func nary2Lisp(schema sc.Schema, op string, exprs []Expr) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = e.Lisp(schema)
	}
	// Done
	return sexp.NewList(arr)
}
