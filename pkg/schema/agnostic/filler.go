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
package agnostic

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
	"github.com/consensys/go-corset/pkg/util/word"
)

var _ Computation = &PolyFil{}

// PolyFil is a simple construct used for filling the various temporary
// registers (e.g. for holding carry values) that arise as part of register
// splitting.
type PolyFil struct {
	rshift uint
	poly   RelativePolynomial
}

// NewPolyFil constructs a new poly fil computation.
func NewPolyFil(rshift uint, poly RelativePolynomial) *PolyFil {
	return &PolyFil{rshift, poly}
}

// ApplyShift implementation for Term interface.
func (p *PolyFil) ApplyShift(shift int) Computation {
	panic("unsupported operation")
}

// Bounds implementation for Boundable interface.
func (p *PolyFil) Bounds() util.Bounds {
	return BoundsForPolynomial(p.poly)
}

// EvalAt implementation for Evaluable interface.
func (p *PolyFil) EvalAt(k int, tr trace.Module[word.BigEndian], sc register.Map) (word.BigEndian, error) {
	val := EvalPolynomial(uint(k), p.poly, tr)
	//
	if p.rshift != 0 {
		return val.Rsh(p.rshift), nil
	}
	//
	return val, nil
}

// Lisp implementation for Lispifiable interface.
func (p *PolyFil) Lisp(global bool, mapping register.Map) sexp.SExp {
	body := poly.Lisp(p.poly, func(id register.RelativeId) string {
		var name = mapping.Register(id.Id()).Name
		//
		if id.Shift() == 0 {
			return name
		}
		//
		return fmt.Sprintf("%s[%d]", name, id.Shift())
	})
	//
	if p.rshift == 0 {
		return body
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol(">>"),
		sexp.NewSymbol(fmt.Sprintf("%d", p.rshift)),
		body,
	})
}

// RequiredRegisters implementation for Contextual interface.
func (p *PolyFil) RequiredRegisters() *set.SortedSet[uint] {
	var regs = set.NewSortedSet[uint]()
	//
	for i := range p.poly.Len() {
		for _, ident := range p.poly.Term(i).Vars() {
			regs.Insert(ident.Id().Unwrap())
		}
	}
	//
	return regs
}

// RequiredCells implementation for Contextual interface
func (p *PolyFil) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	panic("unsupported operation")
}

// ShiftRange implementation for Term interface.
func (p *PolyFil) ShiftRange() (int, int) {
	panic("unsupported operation")
}

// Simplify implementation for Term interface.
func (p *PolyFil) Simplify(casts bool) Computation {
	// By definition, cannot further simplify a polynomial
	return p
}

// Substitute implementation for Substitutable interface.
func (p *PolyFil) Substitute(mapping map[string]word.BigEndian) {
	panic("unsupported operation")
}

// ValueRange implementation for Term interface.
func (p *PolyFil) ValueRange(mapping register.Map) math.Interval {
	panic("unsupported operation")
}

// ============================================================================
// Helpers
// ============================================================================

// EvalPolynomial evaluates a given polynomial with a given environment (i.e. mapping of variables to values)
func EvalPolynomial[F field.Element[F]](row uint, poly RelativePolynomial, mod trace.Module[F]) F {
	var val F
	// Sum evaluated terms
	for i := uint(0); i < poly.Len(); i++ {
		ith := EvalMonomial(row, poly.Term(i), mod)
		//
		if i == 0 {
			val = ith
		} else {
			val = val.Add(ith)
		}
	}
	// Done
	return val
}

// EvalMonomial evaluates a given polynomial with a given environment (i.e. mapping of variables to values)
func EvalMonomial[F field.Element[F]](row uint, term RelativeMonomial, mod trace.Module[F]) F {
	var (
		acc   F
		coeff big.Int = term.Coefficient()
	)
	// Initialise accumulator
	acc = acc.SetBytes(coeff.Bytes())
	//
	for j := uint(0); j < term.Len(); j++ {
		var (
			jth = term.Nth(j)
			col = mod.Column(jth.Unwrap())
			v   = col.Get(int(row) + jth.Shift())
		)
		//
		acc = acc.Mul(v)
	}
	//
	return acc
}

// BoundsForPolynomial determines the largest positive / negative shift for any variable in a given polynomial.
func BoundsForPolynomial(p RelativePolynomial) util.Bounds {
	var bounds util.Bounds
	//
	for i := range p.Len() {
		ith := BoundsForMonomial(p.Term(i))
		bounds.Union(&ith)
	}
	//
	return bounds
}

// BoundsForMonomial determines the largest positive / negative shift for any variable in a given monomial
func BoundsForMonomial(p RelativeMonomial) util.Bounds {
	var bounds util.Bounds
	//
	for i := range p.Len() {
		ith := BoundsForRegister(p.Nth(i))
		bounds.Union(&ith)
	}
	//
	return bounds
}

// BoundsForRegister determines the largest positive / negative shift for any relative register access.
func BoundsForRegister(p register.RelativeId) util.Bounds {
	if p.Shift() >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.Shift()))
	}
	// Negative shift
	return util.NewBounds(uint(-p.Shift()), 0)
}
