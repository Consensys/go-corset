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
package assignment

import (
	"fmt"
	"math/big"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// CarryAssign is used for computing the value of carry lines introduced during
// register splitting.  The intuition is that we have an expression which is
// evaluated to a given value, and the right shifted by a given amount.  The
// result is assigned to the carry register.
type CarryAssign[F field.Element[F]] struct {
	// Target column for this shift assignment
	Target sc.RegisterRef
	Shift  uint
	Source sc.Polynomial
}

// NewCarryAssign constructs a new carry assignment.
func NewCarryAssign[F field.Element[F]](target sc.RegisterRef, shift uint, source sc.Polynomial) *CarryAssign[F] {
	//
	return &CarryAssign[F]{target, shift, source}
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *CarryAssign[F]) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute computes the values of columns defined by this assignment. This
// requires copying the data in the source columns, and sorting that data
// according to the permutation criteria.
func (p *CarryAssign[F]) Compute(trace tr.Trace[F], schema sc.AnySchema[F],
) ([]array.MutArray[F], error) {
	var (
		trModule = trace.Module(p.Target.Module())
		scModule = schema.Module(p.Target.Module())
		width    = scModule.Register(p.Target.Register()).Width
	)
	// Determine multiplied height
	height := trModule.Height()
	//
	data := trace.Builder().NewArray(height, width)
	// Calculate in a forwards direction
	for i := range height {
		var element F
		// Evaluate polynomial
		val := evalPolynomial(i, p.Source, trModule)
		// Right shift result
		val.Rsh(&val, p.Shift)
		// Assign result
		data.Set(i, element.SetBytes(val.Bytes()))
	}
	// Done
	return []array.MutArray[F]{data}, nil
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *CarryAssign[F]) Consistent(_ sc.AnySchema[F]) []error {
	return nil
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *CarryAssign[F]) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *CarryAssign[F]) RegistersRead() []sc.RegisterRef {
	var regs []sc.RegisterRef
	//
	for _, rid := range agnostic.RegistersRead(p.Source) {
		regs = append(regs, sc.NewRegisterRef(p.Target.Module(), rid))
	}
	//
	return regs
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *CarryAssign[F]) RegistersWritten() []sc.RegisterRef {
	return []sc.RegisterRef{p.Target}
}

// Substitute any matchined labelled constants within this assignment
func (p *CarryAssign[F]) Substitute(map[string]F) {
	// Nothing to do here.
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *CarryAssign[F]) Lisp(schema sc.AnySchema[F]) sexp.SExp {
	var (
		sources = sexp.EmptyList()
	)
	// Determine target details
	module := schema.Module(p.Target.Module())
	ith := module.Register(p.Target.Register())
	name := sexp.NewSymbol(ith.QualifiedName(module))
	datatype := sexp.NewSymbol(fmt.Sprintf("u%d", ith.Width))
	target := sexp.NewList([]sexp.SExp{name, datatype})

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("carryassign"),
		target,
		sources,
	})
}

// ============================================================================
// Eval
// ============================================================================

// Eval evaluates a given polynomial with a given environment (i.e. mapping of variables to values)
func evalPolynomial[F field.Element[F]](row uint, poly agnostic.Polynomial, mod tr.Module[F]) big.Int {
	var val big.Int
	// Sum evaluated terms
	for i := uint(0); i < poly.Len(); i++ {
		ith := evalTerm(row, poly.Term(i), mod)
		//
		if i == 0 {
			val = ith
		} else {
			val.Add(&val, &ith)
		}
	}
	// Done
	return val
}

func evalTerm[F field.Element[F]](row uint, term agnostic.Monomial, mod tr.Module[F]) big.Int {
	var (
		acc   big.Int
		coeff big.Int = term.Coefficient()
	)
	// Initialise accumulator
	acc.Set(&coeff)
	//
	for j := uint(0); j < term.Len(); j++ {
		var (
			jth = mod.Column(term.Nth(j).Unwrap())
			v   = jth.Get(int(row))
			w   big.Int
		)
		//
		w.SetBytes(v.Bytes())
		acc.Mul(&acc, &w)
	}
	//
	return acc
}
