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
package schema

import (
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Contextual captures something which requires an evaluation context (i.e. a
// single enclosing module) in order to make sense.  For example, expressions
// require a single context.  This interface is separated from Evaluable (and
// Testable) because HIR expressions do not implement Evaluable.
type Contextual interface {
	Lispifiable
	// Context returns the evaluation context (i.e. enclosing module + length
	// multiplier) for this constraint.  Every expression must have a single
	// evaluation context.  This function therefore attempts to determine what
	// that is, or return false to signal an error. There are several failure
	// modes which need to be considered.  Firstly, if the expression has no
	// enclosing module (e.g. because it is a constant expression) then it will
	// return 'math.MaxUint` to signal this.  Secondly, if the expression has
	// multiple (i.e. conflicting) enclosing modules then it will return false
	// to signal this.  Likewise, the expression could have a single enclosing
	// module but multiple conflicting length multipliers, in which case it also
	// returns false.
	Context(Module) trace.Context

	// RequiredColumns returns the set of columns on which this term depends.
	// That is, columns whose values may be accessed when evaluating this term
	// on a given trace.
	RequiredColumns() *set.SortedSet[uint]
	// RequiredCells returns the set of trace cells on which evaluation of this
	// constraint element depends.
	RequiredCells(int, trace.Module) *set.AnySortedSet[trace.CellRef]
}

// Lispifiable captures a schema element which can be turned into a stand alone
// S-expression (e.g. for printing).
type Lispifiable interface {
	// Lisp converts this schema element into a simple S-Expression, for example
	// so it can be printed.
	Lisp(Module) sexp.SExp
}
