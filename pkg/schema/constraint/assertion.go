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
package constraint

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// AssertionFailure provides structural information about a failing vanishing constraint.
type AssertionFailure[F any] struct {
	// Handle of the failing constraint
	Handle string
	//
	Context schema.ModuleId
	// Constraint expression
	Constraint ir.Testable[F]
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *AssertionFailure[F]) Message() string {
	// Construct useful error message
	return fmt.Sprintf("assertion \"%s\" does not hold (row %d)", p.Handle, p.Row)
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *AssertionFailure[F]) RequiredCells(tr trace.Trace[F]) *set.AnySortedSet[trace.CellRef] {
	return p.Constraint.RequiredCells(int(p.Row), p.Context)
}

func (p *AssertionFailure[F]) String() string {
	return p.Message()
}

// Assertion is similar to a vanishing constraint but is used only for
// debugging / testing / verification.  Unlike vanishing constraints, property
// assertions do not represent something that the prover can enforce.  Rather,
// they represent properties which are expected to hold for every valid trace.
// That is, they should be implied by the actual constraints.  Thus, whilst the
// prover cannot enforce such properties, external tools (such as for formal
// verification) can attempt to ensure they do indeed always hold.
type Assertion[F field.Element[F], T ir.Testable[F]] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// Enclosing module for this assertion.  This restricts the asserted
	// property to access only columns from within this module.
	Context schema.ModuleId
	// The actual assertion itself, namely an expression which
	// should hold (i.e. vanish) for every row of a trace.
	// Observe that this can be any function which is computable
	// on a given trace --- we are not restricted to expressions
	// which can be arithmetised.
	Property T
}

// NewAssertion constructs a new property assertion!
func NewAssertion[F field.Element[F], T ir.Testable[F]](handle string, ctx schema.ModuleId, property T,
) Assertion[F, T] {
	//
	return Assertion[F, T]{handle, ctx, property}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Assertion[F, T]) Consistent(schema schema.AnySchema[F]) []error {
	return CheckConsistent(p.Context, schema, p.Property)
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p Assertion[F, T]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p Assertion[F, T]) Contexts() []schema.ModuleId {
	return []schema.ModuleId{p.Context}
}

// Bounds is not required for a property assertion since these are not real
// constraints.
func (p Assertion[F, T]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table. If so, return nil otherwise return an error.
//
//nolint:revive
func (p Assertion[F, T]) Accepts(tr trace.Trace[F], sc schema.AnySchema[F]) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		trModule = tr.Module(p.Context)
		scModule = sc.Module(p.Context)
		// Determine height of enclosing module
		height = tr.Module(p.Context).Height()
		// Determine well-definedness bounds for this constraint
		bounds = p.Property.Bounds()
	)
	// Sanity check enough rows
	if bounds.End < height {
		// Check all in-bounds values
		for k := bounds.Start; k < (height - bounds.End); k++ {
			// Check whether property holds (or was undefined)
			if ok, id, err := p.Property.TestAt(int(k), trModule, scModule); err != nil {
				// Evaluation failure
				return coverage, &InternalFailure[F]{Handle: p.Handle, Context: p.Context, Row: k, Error: err.Error()}
			} else if !ok {
				return coverage, &AssertionFailure[F]{p.Handle, p.Context, p.Property, k}
			} else {
				// Update coverage
				coverage.Insert(id)
			}
		}
	}
	// All good
	return coverage, nil
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p Assertion[F, T]) Lisp(schema schema.AnySchema[F]) sexp.SExp {
	var module = schema.Module(p.Context)
	// Construct the list
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("assert"),
		sexp.NewSymbol(p.Handle),
		p.Property.Lisp(false, module),
	})
}

// Substitute any matchined labelled constants within this constraint
func (p Assertion[F, T]) Substitute(mapping map[string]F) {
	p.Property.Substitute(mapping)
}
