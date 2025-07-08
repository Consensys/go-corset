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
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Parser is responsible for parsing S-expressions into polynomials.
type Parser[S comparable, T Term[S, T], P Polynomial[S, T, P]] struct {
	// Maps S-Expressions to their spans in the original source file.  This is
	// used for reporting syntax errors.
	srcmap *source.Map[sexp.SExp]
	// Function for constructing constant terms from strings
	constructor func(string) (T, error)
}

// NewParser constructs a new parser for a given source map.
func NewParser[S comparable, T Term[S, T], P Polynomial[S, T, P]](srcmap *source.Map[sexp.SExp],
	constructor func(string) (T, error)) *Parser[S, T, P] {
	return &Parser[S, T, P]{srcmap, constructor}
}

// Parse a given S-expression into a polynomial, or produce one or more syntax errors.
//
// nolint
func (p *Parser[S, T, P]) Parse(sexp sexp.SExp) (P, []source.SyntaxError) {
	return p.parsePoly(sexp)
}

func (p *Parser[S, T, P]) parsePoly(expr sexp.SExp) (P, []source.SyntaxError) {
	var poly P
	//
	switch e := expr.(type) {
	case *sexp.Symbol:
		return p.parseSymbol(e)
	case *sexp.List:
		return p.parseList(e)
	default:
		return poly, p.srcmap.SyntaxErrors(expr, "unknown term")
	}
}

func (p *Parser[S, T, P]) parseSymbol(symbol *sexp.Symbol) (P, []source.SyntaxError) {
	var poly P
	//
	term, err := p.constructor(symbol.Value)
	// Check for errors
	if err == nil {
		// Initial polynomial from term
		poly = poly.Set(term)
		// Done
		return poly, nil
	}
	// Syntax error
	return poly, p.srcmap.SyntaxErrors(symbol, err.Error())
}

func (p *Parser[S, T, P]) parseList(list *sexp.List) (P, []source.SyntaxError) {
	var poly P
	//
	if list.Len() <= 1 {
		return poly, p.srcmap.SyntaxErrors(list, "malformed expression")
	} else if list.Get(0).AsSymbol() == nil {
		return poly, p.srcmap.SyntaxErrors(list.Get(0), "expected operator")
	}
	//
	operator := list.Get(0).AsSymbol().Value
	//
	switch operator {
	case "+":
		return p.foldList(list.Elements[1:], func(l, r P) P {
			return l.Add(r)
		})
	case "-":
		return p.foldList(list.Elements[1:], func(l, r P) P {
			return l.Sub(r)
		})
	case "*":
		return p.foldList(list.Elements[1:], func(l, r P) P {
			return l.Mul(r)
		})
	default:
		// problem
		return poly, p.srcmap.SyntaxErrors(list.Get(0), "unknown operator")
	}
}

// Type of operators to be used with fold.
type foldOp[S comparable, T Term[S, T], P Polynomial[S, T, P]] func(P, P) P

func (p *Parser[S, T, P]) foldList(elements []sexp.SExp, op foldOp[S, T, P]) (P, []source.SyntaxError) {
	var res P
	// Fold over each element
	for i := 0; i < len(elements); i++ {
		if poly, errs := p.parsePoly(elements[i]); len(errs) > 0 {
			return res, errs
		} else if i == 0 {
			res = poly
		} else {
			res = op(res, poly)
		}
	}
	//
	return res, nil
}
