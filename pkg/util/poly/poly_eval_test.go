// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package poly

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

type Poly = *ArrayPoly[string]

func Test_PolyEval_01(t *testing.T) {
	points := [][]uint{{123, 0}, {123, 1}}
	check(t, "123", points)
}

func Test_PolyEval_02(t *testing.T) {
	points := [][]uint{{0, 0}, {1, 1}}
	check(t, "a", points)
}

func Test_PolyEval_03(t *testing.T) {
	points := [][]uint{{1, 0}, {2, 1}, {3, 2}}
	check(t, "(+ a 1)", points)
}

func Test_PolyEval_04(t *testing.T) {
	points := [][]uint{{0, 1}, {1, 2}, {2, 3}}
	check(t, "(- a 1)", points)
}

func Test_PolyEval_05(t *testing.T) {
	points := [][]uint{{2, 1}, {4, 2}, {6, 3}}
	check(t, "(* a 2)", points)
}

// Check the evaluation of a polynomial at evaluation given points.
func check(t *testing.T, input string, points [][]uint) {
	// Parse the polynomial, producing one or more errors.
	if p, errs := parse(input); len(errs) != 0 {
		t.Error(errs)
	} else {
		// Evaluate the polynomial at the given points, recalling that the first
		// point is always the outcome.
		for _, pnt := range points {
			env := make(map[string]big.Int)
			env["a"] = *big.NewInt(int64(pnt[1]))
			actual := Eval(p, env)
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
func parse(input string) (Poly, []source.SyntaxError) {
	srcfile := source.NewSourceFile("test", []byte(input))
	// Parse input as S-expression
	term, srcmap, err := sexp.Parse(srcfile)
	if err != nil {
		return nil, []source.SyntaxError{*err}
	}
	// Now, convert S-expression into polynomial
	parser := NewParser[string, Monomial[string], Poly](srcmap, termConstructor)
	//
	return parser.Parse(term)
}

// Default construct for terms.
func termConstructor(symbol string) (Monomial[string], error) {
	// Check for constant
	if (symbol[0] >= '0' && symbol[0] <= '9') || symbol[0] == '-' {
		return constantConstructor(symbol)
	}
	// Construct variable
	one := big.NewInt(1)
	//
	return NewMonomial(*one, symbol), nil
}

// Constructor for constant literals.
func constantConstructor(symbol string) (Monomial[string], error) {
	var (
		num  big.Int
		term Monomial[string]
	)
	//
	if _, ok := num.SetString(symbol, 10); !ok {
		return term, errors.New("invalid constant")
	}
	//
	return NewMonomial[string](num), nil
}
