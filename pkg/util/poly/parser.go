package poly

import (
	"github.com/consensys/go-corset/pkg/sexp"
)

// Parser is responsible for parsing S-expressions into polynomials.
type Parser[S comparable, T Term[S]] struct {
	// Maps S-Expressions to their spans in the original source file.  This is
	// used for reporting syntax errors.
	srcmap *sexp.SourceMap[sexp.SExp]
	// Function for constructing constant terms from strings
	constructor func(string) (T, error)
}

// NewParser constructs a new parser for a given source map.
func NewParser[S comparable, T Term[S]](srcmap *sexp.SourceMap[sexp.SExp],
	constructor func(string) (T, error)) *Parser[S, T] {
	return &Parser[S, T]{srcmap, constructor}
}

// Parse a given S-expression into a polynomial, or produce one or more syntax errors.
func (p *Parser[S, T]) Parse(sexp sexp.SExp) (*ArrayPoly[S, T], []sexp.SyntaxError) {
	return p.parsePoly(sexp)
}

func (p *Parser[S, T]) parsePoly(expr sexp.SExp) (*ArrayPoly[S, T], []sexp.SyntaxError) {
	switch e := expr.(type) {
	case *sexp.Symbol:
		return p.parseSymbol(e)
	case *sexp.List:
		return p.parseList(e)
	default:
		return nil, p.srcmap.SyntaxErrors(expr, "unknown term")
	}
}

func (p *Parser[S, T]) parseSymbol(symbol *sexp.Symbol) (*ArrayPoly[S, T], []sexp.SyntaxError) {
	term, err := p.constructor(symbol.Value)
	// Check for errors
	if err == nil {
		return NewArrayPoly(term), nil
	}
	// Syntax error
	return nil, p.srcmap.SyntaxErrors(symbol, err.Error())
}

func (p *Parser[S, T]) parseList(list *sexp.List) (*ArrayPoly[S, T], []sexp.SyntaxError) {
	if list.Len() <= 1 {
		return nil, p.srcmap.SyntaxErrors(list, "malformed expression")
	} else if list.Get(0).AsSymbol() == nil {
		return nil, p.srcmap.SyntaxErrors(list.Get(0), "expected operator")
	}
	//
	operator := list.Get(0).AsSymbol().Value
	//
	switch operator {
	case "+":
		return p.foldList(list.Elements[1:], func(l *ArrayPoly[S, T], r *ArrayPoly[S, T]) {
			l.Add(r)
		})
	case "-":
		return p.foldList(list.Elements[1:], func(l *ArrayPoly[S, T], r *ArrayPoly[S, T]) {
			l.Sub(r)
		})
	case "*":
		return p.foldList(list.Elements[1:], func(l *ArrayPoly[S, T], r *ArrayPoly[S, T]) {
			l.Mul(r)
		})
	default:
		// problem
		return nil, p.srcmap.SyntaxErrors(list.Get(0), "unknown operator")
	}
}

// Type of operators to be used with fold.
type foldOp[S comparable, T Term[S]] func(*ArrayPoly[S, T], *ArrayPoly[S, T])

func (p *Parser[S, T]) foldList(elements []sexp.SExp, op foldOp[S, T]) (*ArrayPoly[S, T], []sexp.SyntaxError) {
	var res *ArrayPoly[S, T]
	// Fold over each element
	for i := 0; i < len(elements); i++ {
		if poly, errs := p.parsePoly(elements[i]); len(errs) > 0 {
			return nil, errs
		} else if i == 0 {
			res = poly
		} else {
			op(res, poly)
		}
	}
	//
	return res, nil
}
