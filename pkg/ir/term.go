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
package ir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Enclosed represents anything which, ultimately, must be associated with a
// module (for example a term).  Furthermore, it indicates the given item has
// been associated with the given module.
type Enclosed[T any] struct {
	Module schema.ModuleId
	Item   T
}

// Enclose a given item in a given module.
func Enclose[T any](mid schema.ModuleId, item T) Enclosed[T] {
	return Enclosed[T]{mid, item}
}

// Contextual captures something which requires an evaluation context (i.e. a
// single enclosing module) in order to make sense.  For example, expressions
// require a single context.  This interface is separated from Evaluable (and
// Testable) because HIR expressions do not implement Evaluable.
type Contextual interface {
	// RequiredRegisters returns the set of registers on which this term depends.
	// That is, registers whose values may be accessed when evaluating this term
	// on a given trace.
	RequiredRegisters() *set.SortedSet[uint]
	// RequiredCells returns the set of trace cells on which evaluation of this
	// constraint element depends.
	RequiredCells(int, trace.ModuleId) *set.AnySortedSet[trace.CellRef]
}

// Evaluable captures something which can be evaluated on a given table row to
// produce an evaluation point.  For example, expressions in the
// Mid-Level or Arithmetic-Level IR can all be evaluated at rows of a
// table.
type Evaluable interface {
	util.Boundable
	Contextual
	Substitutable
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a register which does not exist.
	EvalAt(int, trace.Module, schema.Module) (fr.Element, error)
	// Lisp converts this schema element into a simple S-Expression, for example
	// so it can be printed.
	Lisp(schema.RegisterMap) sexp.SExp
	// IsDefined checks whether a given evaluable expression is defined, or not.
	IsDefined() bool
	// ValueRange returns the interval of values that this term can evaluate to.
	// For terms accessing registers, this is determined by the declared width of
	// the register.
	ValueRange(mapping schema.RegisterMap) math.Interval
}

// Substitutable captures the notion of a term which may contain labelled
// constants that can be substituted.
type Substitutable interface {
	// Substitute any matchined labelled constants within this constraint
	Substitute(map[string]fr.Element)
}

// Shiftable captures something which can contain row shifted accesses, and
// where we want information or to manipulate those accesses.
type Shiftable[T any] interface {
	// ApplyShift applies a given shift to all variable accesses in a given term
	// by a given amount. This can be used to normalise shifting in certain
	// circumstances.
	ApplyShift(int) T

	// ShiftRange returns the minimum and maximum shift value used anywhere in
	// the given term.
	ShiftRange() (int, int)
}

// Term represents a component of an AIR expression.
type Term[T any] interface {
	Contextual
	Shiftable[T]
	Evaluable
	util.Boundable
	Substitutable

	// Simplify constant expressions down to single values.  For example, "(+ 1
	// 2)" would be collapsed down to "3".  This is then progagated throughout
	// an expression, so that e.g. "(+ X (+ 1 2))" becomes "(+ X 3)"", etc.
	// There is also an option to retain casts, or not.
	Simplify(casts bool) T
}

// Testable captures the notion of a constraint which can be tested on a given
// row of a given trace.  It is very similar to Evaluable, except that it only
// indicates success or failure.  The reason for using this interface over
// Evaluable is that, for historical reasons, constraints at the HIR cannot be
// Evaluable (i.e. because they return multiple values, rather than a single
// value).  However, constraints at the HIR level remain testable.
type Testable interface {
	util.Boundable
	Contextual
	Substitutable
	// TestAt evaluates this expression in a given tabular context and checks it
	// against zero. Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be undefined for
	// several reasons: firstly, if it accesses a row which does not exist (e.g.
	// at index -1); secondly, if it accesses a register which does not exist.
	TestAt(int, trace.Module, schema.Module) (bool, uint, error)
	// Lisp converts this schema element into a simple S-Expression, for example
	// so it can be printed.
	Lisp(schema.RegisterMap) sexp.SExp
}

// LogicalTerm represents a term which can be tested for truth or falsehood.
// For example, an equality comparing two arithmetic terms is a logical term.
type LogicalTerm[T any] interface {
	Contextual
	Shiftable[T]
	Testable

	// Simplify constant expressions down to single values.  For example, "(+ 1
	// 2)" would be collapsed down to "3".  This is then progagated throughout
	// an expression, so that e.g. "(+ X (+ 1 2))" becomes "(+ X 3)"", etc.
	// There is also an option to retain casts, or not.
	Simplify(casts bool) T
}
