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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Normalise constructs an expression representing the normalised value of e.
// That is, an expression which is 0 when e is 0, and 1 when e is non-zero.
// This is done by introducing a computed column to hold the (pseudo)
// mutliplicative inverse of e.
func Normalise(e air.Expr, schema *air.Schema) air.Expr {
	// Construct pseudo multiplicative inverse of e.
	ie := applyPseudoInverseGadget(e, schema)
	// Return e * e⁻¹.
	return e.Mul(ie)
}

// applyPseudoInverseGadget constructs an expression representing the
// (pseudo) multiplicative inverse of another expression.  Since this cannot be computed
// directly using arithmetic constraints, it is done by adding a new computed
// column which holds the multiplicative inverse.  Constraints are also added to
// ensure it really holds the inverted value.
func applyPseudoInverseGadget(e air.Expr, schema *air.Schema) air.Expr {
	// Determine enclosing module.
	ctx := e.Context(schema)
	// Sanity check
	if ctx.IsVoid() || ctx.IsConflicted() {
		panic("conflicting (or void) context")
	}
	// Construct inverse computation
	ie := &psuedoInverse{Expr: e}
	// Determine computed column name
	name := ie.Lisp(schema).String(false)
	// Look up column
	index, ok := sc.ColumnIndexOf(schema, ctx.Module(), name)
	// Add new column (if it does not already exist)
	if !ok {
		// Add computed column, noting that inverses often require the full
		// 256bits available.
		index = schema.AddAssignment(assignment.NewComputedColumn(ctx, name, &sc.FieldType{}, ie))
		// Construct 1/e
		inv_e := air.NewColumnAccess(index, 0)
		// Construct e/e
		e_inv_e := e.Mul(inv_e)
		// Construct 1 == e/e
		one_e_e := air.NewConst64(1).Equate(e_inv_e)
		// Construct (e != 0) ==> (1 == e/e)
		e_implies_one_e_e := e.Mul(one_e_e) // Ensure (e != 0) ==> (1 == e/e)
		l_name := fmt.Sprintf("%s <=", name)
		schema.AddVanishingConstraint(l_name, 0, ctx, util.None[int](), e_implies_one_e_e)
	}
	// Done
	return air.NewColumnAccess(index, 0)
}

// psuedoInverse represents a computation which computes the multiplicative
// inverse of a given AIR expression.
type psuedoInverse struct {
	Expr air.Expr
}

// EvalAt computes the multiplicative inverse of a given expression at a given
// row in the table.
func (e *psuedoInverse) EvalAt(k int, tbl tr.Trace) (fr.Element, error) {
	var inv fr.Element
	// Convert expression into something which can be evaluated, then evaluate
	// it.
	val, err := e.Expr.EvalAt(k, tbl)
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

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (e *psuedoInverse) Context(schema sc.Schema) tr.Context {
	return e.Expr.Context(schema)
}

// Branches returns the total number of logical branches this constraint can
// take during evaluation.
func (e *psuedoInverse) Branches() uint {
	return e.Expr.Branches()
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (e *psuedoInverse) RequiredColumns() *set.SortedSet[uint] {
	return e.Expr.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (e *psuedoInverse) RequiredCells(row int, trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	return e.Expr.RequiredCells(row, trace)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *psuedoInverse) Lisp(schema sc.Schema) sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("inv"),
		e.Expr.Lisp(schema),
	})
}
