package test

import (
	"fmt"
	"math/big"
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

func Test_Poly_03(t *testing.T) {
	points := [][]uint{{1, 0}, {2, 1}, {3, 2}}
	check(t, "(+ a 1)", points)
}

func Test_Poly_04(t *testing.T) {
	points := [][]uint{{0, 1}, {1, 2}, {2, 3}}
	check(t, "(- a 1)", points)
}

func Test_Poly_05(t *testing.T) {
	points := [][]uint{{2, 1}, {4, 2}, {6, 3}}
	check(t, "(* a 2)", points)
}

// Check the evaluation of a polynomial at evaluation given points.
func check(t *testing.T, input string, points [][]uint) {
	// Parse the polynomial, producing one or more errors.
	if p, errs := parse(input); len(errs) != 0 {
		t.Error(errs)
	} else {
		fmt.Printf("POLY=%s\n", p)
		// Evaluate the polynomial at the given points, recalling that the first
		// point is always the outcome.
		for _, pnt := range points {
			env := make(map[string]big.Int)
			env["a"] = *big.NewInt(int64(pnt[1]))
			actual := poly.Eval(p, env)
			expected := big.NewInt(int64(pnt[0]))
			// Evaluate and check
			if actual.Cmp(expected) != 0 {
				err := fmt.Sprintf("incorrect evaluation (was %s, expected %s)", actual.String(), expected.String())
				t.Error(err)
			}
		}
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
	parser := poly.NewParser[string](srcmap, termConstructor)
	//
	return parser.Parse(term)
}

// Default construct for terms.
func termConstructor(symbol string) (string, error) {
	// In theory, we could do some sanity check of the symbol to ensure it meets
	// certain requirements (e.g. does not include arbitrary symbols, etc).
	return symbol, nil
}
