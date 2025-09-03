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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// PseudoInverse represents a computation which computes the multiplicative
// inverse of a given expression.
type PseudoInverse[F field.Element[F]] struct {
	Expr air.Term[F]
}

// EvalAt computes the multiplicative inverse of a given expression at a given
// row in the table.
func (e *PseudoInverse[F]) EvalAt(k int, tr trace.Module[F], sc schema.Module[F]) (F, error) {
	// Convert expression into something which can be evaluated, then evaluate
	// it.
	val, err := e.Expr.EvalAt(k, tr, sc)
	// Go syntax huh?
	inv := val.Inverse()
	// Done
	return inv, err
}

// AsConstant determines whether or not this is a constant expression.  If
// so, the constant is returned; otherwise, nil is returned.  NOTE: this
// does not perform any form of simplification to determine this.
func (e *PseudoInverse[F]) AsConstant() *fr.Element { return nil }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e *PseudoInverse[F]) Bounds() util.Bounds { return e.Expr.Bounds() }

// RequiredRegisters returns the set of registers on which this term depends.
// That is, registers whose values may be accessed when evaluating this term on
// a given trace.
func (e *PseudoInverse[F]) RequiredRegisters() *set.SortedSet[uint] {
	return e.Expr.RequiredRegisters()
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (e *PseudoInverse[F]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return e.Expr.RequiredCells(row, mid)
}

// IsDefined implementation for Evaluable interface.
func (e *PseudoInverse[F]) IsDefined() bool {
	// NOTE: this is technically safe given the limited way that IsDefined is
	// used for lookup selectors.
	return true
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *PseudoInverse[F]) Lisp(global bool, mapping sc.RegisterMap) sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("inv"),
		e.Expr.Lisp(global, mapping),
	})
}

// Substitute implementation for Substitutable interface.
func (e *PseudoInverse[F]) Substitute(mapping map[string]F) {
	panic("unreachable")
}

// ValueRange implementation for Term interface.
func (e *PseudoInverse[F]) ValueRange(mapping schema.RegisterMap) util_math.Interval {
	// This could be managed by having a mechanism for representing infinity
	// (e.g. nil). For now, this is never actually used, so we can just ignore
	// it.
	panic("unreachable")
}
