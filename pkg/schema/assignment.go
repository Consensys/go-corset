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
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Assignment represents an arbitrary computation which determines the values
// for one (or more) computed registers.  For any computed register, there
// should only ever be one assignment.  Likewise, every computed register should
// have an associated assignment.  A good example of an assignment is computed
// the multiplicative inverse of a column in order to implement a non-zero
// check.
type Assignment interface {
	util.Boundable
	// ComputeColumns computes the values of columns defined by this assignment.
	// In order for this computation to makes sense, all columns on which this
	// assignment depends must exist (e.g. are either inputs or have been
	// computed already).  Computed columns do not exist in the original trace,
	// but are added during trace expansion to form the final trace.
	Compute(tr.Trace, Schema[Constraint]) ([]tr.ArrayColumn, error)
	// Consistent applies a number of internal consistency checks.  Whilst not
	// strictly necessary, these can highlight otherwise hidden problems as an aid
	// to debugging.
	Consistent(Schema[Constraint]) []error
	// Returns the set of columns that this assignment depends upon.  That can
	// include both input columns, as well as other computed columns.
	Dependencies() []uint
	// Module returns the enclosing register for all columns computed by this
	// assignment.
	Module() uint
	// Identifier registers assigned by this assignment.
	Registers() []uint
	// Lisp converts this schema element into a simple S-Expression, for example
	// so it can be printed.
	Lisp(AnySchema) sexp.SExp
}
