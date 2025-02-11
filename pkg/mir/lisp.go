package mir

import (
	"fmt"
	"reflect"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

func lispOfTerm(e Term, schema sc.Schema) sexp.SExp {
	switch e := e.(type) {
	case *Add:
		return nary2Lisp(schema, "+", e.Args)
	case *Constant:
		return sexp.NewSymbol(e.Value.String())
	case *ColumnAccess:
		return lispOfColumnAccess(e, schema)
	case *Exp:
		return lispOfExp(e, schema)
	case *Mul:
		return nary2Lisp(schema, "*", e.Args)
	case *Norm:
		return lispOfNormalise(e, schema)
	case *Sub:
		return nary2Lisp(schema, "-", e.Args)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func lispOfColumnAccess(e *ColumnAccess, schema sc.Schema) sexp.SExp {
	var name string
	// Generate name, whilst allowing for schema to be nil.
	if schema != nil {
		name = schema.Columns().Nth(e.Column).QualifiedName(schema)
	} else {
		name = fmt.Sprintf("#%d", e.Column)
	}
	//
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

func nary2Lisp(schema sc.Schema, op string, exprs []Term) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = lispOfTerm(e, schema)
	}
	// Done
	return sexp.NewList(arr)
}

func lispOfNormalise(e *Norm, schema sc.Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("~"), arg})
}

func lispOfExp(e *Exp, schema sc.Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", e.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}
