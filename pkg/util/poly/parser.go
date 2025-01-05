package poly

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/sexp"
)

// Parser is responsible for parsing S-expressions into polynomials.
type Parser[S comparable] struct {
	// Maps S-Expressions to their spans in the original source file.  This is
	// used for reporting syntax errors.
	srcmap *sexp.SourceMap[sexp.SExp]
	// Function for constructing variable identifiers from symbols
	constructor func(string) (S, error)
}

// NewParser constructs a new parser for a given source map.
func NewParser[S comparable](srcmap *sexp.SourceMap[sexp.SExp],
	constructor func(string) (S, error)) *Parser[S] {
	return &Parser[S]{srcmap, constructor}
}

// Parse a given S-expression into a polynomial, or produce one or more syntax errors.
func (p *Parser[S]) Parse(sexp sexp.SExp) (*ArrayPoly[S], []sexp.SyntaxError) {
	return p.parsePoly(sexp)
}

func (p *Parser[S]) parsePoly(expr sexp.SExp) (*ArrayPoly[S], []sexp.SyntaxError) {
	switch e := expr.(type) {
	case *sexp.Symbol:
		return p.parseSymbol(e)
	case *sexp.List:
		return p.parseList(e)
	default:
		return nil, p.srcmap.SyntaxErrors(expr, "unknown term")
	}
}

func (p *Parser[S]) parseSymbol(symbol *sexp.Symbol) (*ArrayPoly[S], []sexp.SyntaxError) {
	value := symbol.Value
	// Check for constant
	if (value[0] >= '0' && value[0] <= '9') || value[0] == '-' {
		return p.parseConstant(symbol)
	}
	// Variable identifier
	var_id, err := p.constructor(symbol.Value)
	// Check for errors
	if err == nil {
		term := NewArrayTerm[S](big.NewInt(1), []S{var_id})
		return NewArrayPoly(term), nil
	}
	// Syntax error
	return nil, p.srcmap.SyntaxErrors(symbol, err.Error())
}

// Constructor for constant literals.
func (p *Parser[S]) parseConstant(symbol *sexp.Symbol) (*ArrayPoly[S], []sexp.SyntaxError) {
	var num big.Int
	//
	if _, ok := num.SetString(symbol.Value, 10); !ok {
		return nil, p.srcmap.SyntaxErrors(symbol, "invalid constant")
	}
	//
	term := NewArrayTerm[S](&num, nil)
	//
	return NewArrayPoly(term), nil
}

func (p *Parser[S]) parseList(list *sexp.List) (*ArrayPoly[S], []sexp.SyntaxError) {
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
		return p.foldList(list.Elements[1:], func(l *ArrayPoly[S], r *ArrayPoly[S]) {
			l.Add(r)
		})
	case "-":
		return p.foldList(list.Elements[1:], func(l *ArrayPoly[S], r *ArrayPoly[S]) {
			l.Sub(r)
		})
	case "*":
		return p.foldList(list.Elements[1:], func(l *ArrayPoly[S], r *ArrayPoly[S]) {
			l.Mul(r)
		})
	default:
		// problem
		return nil, p.srcmap.SyntaxErrors(list.Get(0), "unknown operator")
	}
}

// Type of operators to be used with fold.
type foldOp[S comparable] func(*ArrayPoly[S], *ArrayPoly[S])

func (p *Parser[S]) foldList(elements []sexp.SExp, op foldOp[S]) (*ArrayPoly[S], []sexp.SyntaxError) {
	var res *ArrayPoly[S]
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
