package polynomial

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
		panic("todo")
	default:
		panic("todo")
	}
}

func (p *Parser[S, T]) parseSymbol(symbol *sexp.Symbol) (*ArrayPoly[S, T], []sexp.SyntaxError) {
	term, err := p.constructor(symbol.Value)
	// Check for errors
	if err == nil {
		return NewArrayPoly(term), nil
	}
	// Syntax error
	p.srcmap.SyntaxError(symbol, err)
}

func (p *Parser[S, T]) syntaxErrors(term sexp.SExp) []sexp.SyntaxError {
	span := p.srcmap.Get(term)
	err := sexp.SyntaxError{}
}
