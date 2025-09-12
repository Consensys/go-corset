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
package gadgets

import (
	"fmt"
	"math"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Normalise constructs an expression representing the normalised value of e.
// That is, an expression which is 0 when e is 0, and 1 when e is non-zero.
// This is done by introducing a computed column to hold the (pseudo)
// multiplicative inverse of e.
func Normalise[F field.Element[F]](e air.Term[F], module *air.ModuleBuilder[F]) air.Term[F] {
	// Construct pseudo multiplicative inverse of e.
	ie := applyPseudoInverseGadget(e, module)
	// Return e * e⁻¹.
	return ir.Product(e, ie)
}

// applyPseudoInverseGadget constructs an expression representing the
// (pseudo) multiplicative inverse of another expression.  Since this cannot be computed
// directly using arithmetic constraints, it is done by adding a new computed
// column which holds the multiplicative inverse.  Constraints are also added to
// ensure it really holds the inverted value.
func applyPseudoInverseGadget[F field.Element[F]](e air.Term[F], module *air.ModuleBuilder[F]) air.Term[F] {
	var (
		// Construct inverse computation
		ie = &pseudoInverse[F]{Expr: e}
		// Determine computed column name
		name = ie.Lisp(true, module).String(false)
		// Look up column
		index, ok = module.HasRegister(name)
		// Default padding (for now)
		padding = ir.PaddingFor(ie, module)
	)
	// Add new column (if it does not already exist)
	if !ok {
		// Indicate column has "field element width".
		var bitwidth uint = math.MaxUint
		// Add computed register.
		index = module.NewRegister(sc.NewComputedRegister(name, bitwidth, padding))
		target := sc.NewRegisterRef(module.Id(), index)
		// Add inverse assignment
		module.AddAssignment(assignment.NewPseudoInverse(target, e))
		// Construct proof of 1/e
		inv_e := ir.NewRegisterAccess[F, air.Term[F]](index, 0)
		// Construct e/e
		e_inv_e := ir.Product(e, inv_e)
		// Construct 1 == e/e
		one_e_e := ir.Subtract(ir.Const64[F, air.Term[F]](1), e_inv_e)
		// Construct (e != 0) ==> (1 == e/e)
		e_implies_one_e_e := ir.Product(e, one_e_e)
		l_name := fmt.Sprintf("%s <=", name)
		module.AddConstraint(air.NewVanishingConstraint(l_name, module.Id(), util.None[int](), e_implies_one_e_e))
	}
	// Done
	return ir.NewRegisterAccess[F, air.Term[F]](index, 0)
}

// pseudoInverse represents a computation which computes the multiplicative
// inverse of a given expression.  This is only needed now for the padding
// computation.
type pseudoInverse[F field.Element[F]] struct {
	Expr air.Term[F]
}

// EvalAt computes the multiplicative inverse of a given expression at a given
// row in the table.
func (e *pseudoInverse[F]) EvalAt(k int, tr trace.Module[F], sc schema.Module[F]) (F, error) {
	panic("unreachable")
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e *pseudoInverse[F]) Bounds() util.Bounds { return e.Expr.Bounds() }

// RequiredRegisters returns the set of registers on which this term depends.
// That is, registers whose values may be accessed when evaluating this term on
// a given trace.
func (e *pseudoInverse[F]) RequiredRegisters() *set.SortedSet[uint] {
	return e.Expr.RequiredRegisters()
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (e *pseudoInverse[F]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return e.Expr.RequiredCells(row, mid)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *pseudoInverse[F]) Lisp(global bool, mapping sc.RegisterMap) sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("inv"),
		e.Expr.Lisp(global, mapping),
	})
}

// Substitute implementation for Substitutable interface.
func (e *pseudoInverse[F]) Substitute(mapping map[string]F) {
	panic("unreachable")
}

// ValueRange implementation for Term interface.
func (e *pseudoInverse[F]) ValueRange(mapping schema.RegisterMap) util_math.Interval {
	// This could be managed by having a mechanism for representing infinity
	// (e.g. nil). For now, this is never actually used, so we can just ignore
	// it.
	panic("unreachable")
}
