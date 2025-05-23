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

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// VanishingFailure provides structural information about a failing vanishing constraint.
type VanishingFailure struct {
	// Handle of the failing constraint
	Handle string
	// Constraint expression
	Constraint sc.Testable
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *VanishingFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("constraint \"%s\" does not hold (row %d)", p.Handle, p.Row)
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *VanishingFailure) RequiredCells(trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	return p.Constraint.RequiredCells(int(p.Row), trace)
}

func (p *VanishingFailure) String() string {
	return p.Message()
}

// VanishingConstraint specifies a constraint which should hold on every row of the
// table.  The only exception is when the constraint is undefined (e.g. because
// it references a non-existent table cell).  In such case, the constraint is
// ignored.  This is parameterised by the type of the constraint expression.
// Thus, we can reuse this definition across the various intermediate
// representations (e.g. Mid-Level IR, Arithmetic IR, etc).
type VanishingConstraint[T sc.Testable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	Handle string
	// A further differentiator to manage distinct low-level constraints arising
	// from high-level constraints.
	Case uint
	// Evaluation Context for this constraint which must match that of the
	// constrained expression itself.
	Context tr.Context
	// Indicates (when empty) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given.
	Domain util.Option[int]
	// The actual Constraint itself (e.g. an expression which
	// should evaluate to zero, etc)
	Constraint T
}

// NewVanishingConstraint constructs a new vanishing constraint!
func NewVanishingConstraint[T sc.Testable](handle string, casenum uint, context tr.Context,
	domain util.Option[int], constraint T) *VanishingConstraint[T] {
	return &VanishingConstraint[T]{handle, casenum, context, domain, constraint}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p *VanishingConstraint[E]) Name() (string, uint) {
	return p.Handle, p.Case
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p *VanishingConstraint[E]) Contexts() []tr.Context {
	return []tr.Context{p.Context}
}

// Branches returns the total number of logical branches this constraint can
// take during evaluation.
func (p *VanishingConstraint[E]) Branches() uint {
	return p.Constraint.Branches()
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p *VanishingConstraint[T]) Bounds(module uint) util.Bounds {
	if p.Context.Module() == module {
		return p.Constraint.Bounds()
	}
	//
	return util.EMPTY_BOUND
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p *VanishingConstraint[T]) Accepts(tr tr.Trace) (bit.Set, sc.Failure) {
	handle := determineHandle(p.Handle, p.Context, tr)
	//
	if p.Domain.IsEmpty() {
		// Global Constraint
		return HoldsGlobally(handle, p.Context, p.Constraint, tr)
	}
	// Extract domain
	domain := p.Domain.Unwrap()
	// Local constraint
	var start uint
	// Handle negative domains
	if domain < 0 {
		// Determine height of enclosing module
		height := tr.Height(p.Context)
		// Negative rows calculated from end of trace.
		start = height + uint(domain)
	} else {
		start = uint(domain)
	}
	//
	var coverage bit.Set
	// Check specific row
	err, id := HoldsLocally(start, handle, p.Constraint, tr)
	//
	coverage.Insert(id)
	//
	return coverage, err
}

// HoldsGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func HoldsGlobally[T sc.Testable](handle string, ctx tr.Context, constraint T, tr tr.Trace) (bit.Set, sc.Failure) {
	var coverage bit.Set
	// Determine height of enclosing module
	height := tr.Height(ctx)
	// Determine well-definedness bounds for this constraint
	bounds := constraint.Bounds()
	// Sanity check enough rows
	if bounds.End < height {
		// Check all in-bounds values
		for k := bounds.Start; k < (height - bounds.End); k++ {
			err, id := HoldsLocally(k, handle, constraint, tr)
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
func HoldsLocally[T sc.Testable](k uint, handle string, constraint T, tr tr.Trace) (sc.Failure, uint) {
	ok, id, err := constraint.TestAt(int(k), tr)
	// Check for errors
	if err != nil {
		return &sc.InternalFailure{Handle: handle, Row: k, Term: constraint, Error: err.Error()}, id
	} else if !ok {
		// Evaluation failure
		return &VanishingFailure{handle, constraint, k}, id
	}
	// Success
	return nil, id
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p *VanishingConstraint[T]) Lisp(schema sc.Schema) sexp.SExp {
	var (
		name       string
		multiplier uint
	)
	// Construct qualified name
	if p.Context.IsVoid() {
		name = p.Handle
		multiplier = 1
	} else {
		if module := schema.Modules().Nth(p.Context.Module()); module.Name != "" {
			name = fmt.Sprintf("%s:%s", module.Name, p.Handle)
		} else {
			name = p.Handle
		}
		//
		multiplier = p.Context.LengthMultiplier()
	}
	// Add case
	name = fmt.Sprintf("%s#%d", name, p.Case)
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
			sexp.NewSymbol(name),
			sexp.NewSymbol(fmt.Sprintf("x%d", multiplier))}),
		p.Constraint.Lisp(schema),
	})
}

func determineHandle(handle string, ctx tr.Context, trace tr.Trace) string {
	return fmt.Sprintf("%s.%s", trace.Modules().Nth(ctx.ModuleId).Name(), handle)
}
