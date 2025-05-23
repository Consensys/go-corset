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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Evaluable captures something which can be evaluated on a given table row to
// produce an evaluation point.  For example, expressions in the
// Mid-Level or Arithmetic-Level IR can all be evaluated at rows of a
// table.
type Evaluable interface {
	util.Boundable
	Contextual
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, trace.Module) (fr.Element, error)
	// Branches returns the number of unique evaluation paths through the given
	// constraint.
	Branches() uint
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
	// TestAt evaluates this expression in a given tabular context and checks it
	// against zero. Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be undefined for
	// several reasons: firstly, if it accesses a row which does not exist (e.g.
	// at index -1); secondly, if it accesses a column which does not exist.
	TestAt(int, trace.Module) (bool, uint, error)
	// Branches returns the number of unique evaluation paths through the given
	// constraint.
	Branches() uint
}

// Constraint represents an element which can "accept" a trace, or either reject
// with an error (or eventually perhaps report a warning).
type Constraint interface {
	Lispifiable
	// Accepts determines whether a given constraint accepts a given trace or
	// not.  If not, a failure is produced.  Otherwise, a bitset indicating
	// branch coverage is returned.
	Accepts(trace.Trace) (bit.Set, Failure)
	// Determine the well-definedness bounds for this constraint in both the
	// negative (left) or positive (right) directions.  For example, consider an
	// expression such as "(shift X -1)".  This is technically undefined for the
	// first row of any trace and, by association, any constraint evaluating
	// this expression on that first row is also undefined (and hence must pass)
	Bounds(module uint) util.Bounds
	// Return the total number of logical branches this constraint can take
	// during evaluation.
	Branches() uint
	// Contexts returns the evaluation contexts (i.e. enclosing module + length
	// multiplier) for this constraint.  Most constraints have only a single
	// evaluation context, though some (e.g. lookups) have more.  Note that all
	// constraints have at least one context (which we can call the "primary"
	// context).
	Contexts() []trace.Context
	// Name returns a unique name and case number for a given constraint.  This
	// is useful purely for identifying constraints in reports, etc.  The case
	// number is used to differentiate different low-level constraints which are
	// extracted from the same high-level constraint.
	Name() (string, uint)
}
