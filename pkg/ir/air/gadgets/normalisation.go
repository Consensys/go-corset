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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Normalise constructs an expression representing the normalised value of e.
// That is, an expression which is 0 when e is 0, and 1 when e is non-zero.
// This is done by introducing a computed column to hold the (pseudo)
// mutliplicative inverse of e.
func Normalise(e air.Term, module *air.ModuleBuilder) air.Term {
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
func applyPseudoInverseGadget(e air.Term, module *air.ModuleBuilder) air.Term {
	// Construct inverse computation
	ie := &psuedoInverse{Expr: e}
	// Determine computed column name
	name := ie.Lisp(module).String(false)
	// Look up column
	index, ok := module.HasRegister(name)
	// Add new column (if it does not already exist)
	if !ok {
		// FIXME: this hard-coded constant will need to be changed at some point
		// to properly support field agnosticity.  Currently, this simply
		// signals that the column has no bitwidth constraint.
		var bitwidth uint = math.MaxUint
		// Add computed register.
		index = module.NewRegister(sc.NewComputedRegister(name, bitwidth))
		// Add assignment
		module.AddAssignment(assignment.NewComputedRegister(sc.NewRegisterRef(module.Id(), index), ie))
		// Construct proof of 1/e
		inv_e := ir.NewRegisterAccess[air.Term](index, 0)
		// Construct e/e
		e_inv_e := ir.Product(e, inv_e)
		// Construct 1 == e/e
		one_e_e := ir.Subtract(ir.Const64[air.Term](1), e_inv_e)
		// Construct (e != 0) ==> (1 == e/e)
		e_implies_one_e_e := ir.Product(e, one_e_e)
		l_name := fmt.Sprintf("%s <=", name)
		module.AddConstraint(air.NewVanishingConstraint(l_name, module.Id(), util.None[int](), e_implies_one_e_e))
	}
	// Done
	return ir.NewRegisterAccess[air.Term](index, 0)
}

// psuedoInverse represents a computation which computes the multiplicative
// inverse of a given expression.
type psuedoInverse struct {
	Expr air.Term
}

// EvalAt computes the multiplicative inverse of a given expression at a given
// row in the table.
func (e *psuedoInverse) EvalAt(k int, tr trace.Module, sc schema.Module) (fr.Element, error) {
	var inv fr.Element
	// Convert expression into something which can be evaluated, then evaluate
	// it.
	val, err := e.Expr.EvalAt(k, tr, sc)
	// Go syntax huh?
	inv.Inverse(&val)
	// Done
	return inv, err
}

// AsConstant determines whether or not this is a constant expression.  If
// so, the constant is returned; otherwise, nil is returned.  NOTE: this
// does not perform any form of simplification to determine this.
func (e *psuedoInverse) AsConstant() *fr.Element { return nil }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e *psuedoInverse) Bounds() util.Bounds { return e.Expr.Bounds() }

// RequiredRegisters returns the set of registers on which this term depends.
// That is, registers whose values may be accessed when evaluating this term on
// a given trace.
func (e *psuedoInverse) RequiredRegisters() *set.SortedSet[uint] {
	return e.Expr.RequiredRegisters()
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (e *psuedoInverse) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return e.Expr.RequiredCells(row, mid)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *psuedoInverse) Lisp(module sc.Module) sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("inv"),
		e.Expr.Lisp(module),
	})
}

// Substitute implementation for Substitutable interface.
func (e *psuedoInverse) Substitute(mapping map[string]fr.Element) {
	panic("unreachable")
}
