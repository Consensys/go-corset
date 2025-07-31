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
package vanishing

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	bls12_377 "github.com/consensys/go-corset/pkg/util/field/bls12-377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constraint specifies a constraint which should hold on every row of the
// table.  The only exception is when the constraint is undefined (e.g. because
// it references a non-existent table cell).  In such case, the constraint is
// ignored.  This is parameterised by the type of the constraint expression.
// Thus, we can reuse this definition across the various intermediate
// representations (e.g. Mid-Level IR, Arithmetic IR, etc).
type Constraint[T ir.Testable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// Evaluation Context for this constraint which must match that of the
	// constrained expression itself.
	Context schema.ModuleId
	// Indicates (when empty) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given.
	Domain util.Option[int]
	// The actual Constraint itself (e.g. an expression which
	// should evaluate to zero, etc)
	Constraint T
}

// NewConstraint constructs a new vanishing constraint!
func NewConstraint[T ir.Testable](handle string, context schema.ModuleId,
	domain util.Option[int], constraint T) Constraint[T] {
	return Constraint[T]{handle, context, domain, constraint}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Constraint[E]) Consistent(schema schema.AnySchema) []error {
	return constraint.CheckConsistent(p.Context, schema, p.Constraint)
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p Constraint[E]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p Constraint[E]) Contexts() []schema.ModuleId {
	return []schema.ModuleId{p.Context}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p Constraint[T]) Bounds(module uint) util.Bounds {
	if p.Context == module {
		return p.Constraint.Bounds()
	}
	//
	return util.EMPTY_BOUND
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p Constraint[T]) Accepts(tr trace.Trace[bls12_377.Element], sc schema.AnySchema) (bit.Set, schema.Failure) {
	var (
		// Handle is used for error reporting.
		handle = constraint.DetermineHandle(p.Handle, p.Context, tr)
		// Determine enclosing module
		trModule = tr.Module(p.Context)
		scModule = sc.Module(p.Context)
	)
	//
	if p.Domain.IsEmpty() {
		// Global Constraint
		return HoldsGlobally(handle, p.Context, p.Constraint, trModule, scModule)
	}
	// Extract domain
	domain := p.Domain.Unwrap()
	// Local constraint
	var start uint
	// Handle negative domains
	if domain < 0 {
		// Determine height of enclosing module
		height := tr.Module(p.Context).Height()
		// Negative rows calculated from end of trace.
		start = height + uint(domain)
	} else {
		start = uint(domain)
	}
	//
	var coverage bit.Set
	// Check specific row
	err, id := HoldsLocally(start, handle, p.Constraint, p.Context, trModule, scModule)
	//
	coverage.Insert(id)
	//
	return coverage, err
}

// HoldsGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func HoldsGlobally[T ir.Testable](handle string, ctx schema.ModuleId, constraint T,
	trMod trace.Module, scMod schema.Module) (bit.Set, schema.Failure) {
	//
	var (
		coverage bit.Set
		// Determine height of enclosing module
		height = trMod.Height()
		// Determine well-definedness bounds for this constraint
		bounds = constraint.Bounds()
	)
	// Sanity check enough rows
	if bounds.End < height {
		// Check all in-bounds values
		for k := bounds.Start; k < (height - bounds.End); k++ {
			err, id := HoldsLocally(k, handle, constraint, ctx, trMod, scMod)
			if err != nil {
				return coverage, err
			}
			// Update coverage
			coverage.Insert(id)
		}
	}
	// Success
	return coverage, nil
}

// HoldsLocally checks whether a given constraint holds (e.g. vanishes) on a
// specific row of a trace. If not, report an appropriate error.
func HoldsLocally[T ir.Testable](k uint, handle string, term T, ctx schema.ModuleId,
	trMod trace.Module, scMod schema.Module) (schema.Failure, uint) {
	//
	ok, id, err := term.TestAt(int(k), trMod, scMod)
	// Check for errors
	if err != nil {
		return &constraint.InternalFailure{
			Handle:  handle,
			Context: ctx,
			Row:     k,
			Term:    term,
			Error:   err.Error()}, id
	} else if !ok {
		// Evaluation failure
		return &Failure{handle, term, ctx, k}, id
	}
	// Success
	return nil, id
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p Constraint[T]) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		module = schema.Module(p.Context)
		name   string
	)
	// Construct qualified name
	if module.Name() != "" {
		name = fmt.Sprintf("%s:%s", module.Name(), p.Handle)
	} else {
		name = p.Handle
	}
	// Handle attributes
	if p.Domain.HasValue() {
		switch p.Domain.Unwrap() {
		case 0:
			name = fmt.Sprintf("%s:first", name)
		case -1:
			name = fmt.Sprintf("%s:last", name)
		default:
			domain := p.Domain.Unwrap()
			panic(fmt.Sprintf("domain value %d not supported for local constraint", domain))
		}
	}
	// Construct the list
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("vanish"),
		sexp.NewList([]sexp.SExp{
			sexp.NewSymbol(name)}),
		p.Constraint.Lisp(module),
	})
}

// Substitute any matchined labelled constants within this constraint
func (p Constraint[T]) Substitute(mapping map[string]fr.Element) {
	p.Constraint.Substitute(mapping)
}
