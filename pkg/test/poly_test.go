package test

import (
	"testing"

	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/util/poly"
)

type Poly = poly.Polynomial[string, *poly.ArrayTerm[string]]

func Test_Poly_01(t *testing.T) {
	points := [][]uint{{123, 0}, {123, 1}}
	check(t, "123", points)
}

func Test_Poly_02(t *testing.T) {
	points := [][]uint{{0, 0}, {1, 1}}
	check(t, "a", points)
}

// Check the evaluation of a polynomial at evaluation given points.
func check(t *testing.T, input string, points [][]uint) {
	// Parse the polynomial, producing one or more errors.
	if _, errs := parse(input); len(errs) != 0 {
		t.Error(errs)
	} else {
		panic("got here")
	}
}

// Parse a given input string into a polynomial.
func parse(input string) (Poly, []sexp.SyntaxError) {
	srcfile := sexp.NewSourceFile("test", []byte(input))
	// Parse input as S-expression
	term, srcmap, err := srcfile.Parse()
	if err != nil {
		return nil, []sexp.SyntaxError{*err}
	}
	// Now, convert S-expression into polynomial
	parser := poly.NewParser(srcmap, termConstructor)
	return parser.Parse(term)
}

// Default construct for terms.
func termConstructor(symbol string) (*poly.ArrayTerm[string], error) {
	panic("got here")
}
