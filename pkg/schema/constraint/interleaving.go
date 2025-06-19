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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// InterleavingFailure provides structural information about a failing lookup constraint.
type InterleavingFailure struct {
	// Handle of the failing constraint
	Handle string
	// Relevant context for target expressions.
	TargetContext schema.ModuleId
	// Target expression involved
	Target ir.Evaluable
	// Relevant context for source expressions.
	SourceContext schema.ModuleId
	// Source expression which were missing
	Source ir.Evaluable
	// Target row on which constraint
	Row uint
}

// Message provides a suitable error message
func (p *InterleavingFailure) Message() string {
	return fmt.Sprintf("interleaving \"%s\" failed (row %d)", p.Handle, p.Row)
}

func (p *InterleavingFailure) String() string {
	return p.Message()
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *InterleavingFailure) RequiredCells(tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	var res = set.NewAnySortedSet[trace.CellRef]()
	//
	res.InsertSorted(p.Source.RequiredCells(int(p.Row), p.SourceContext))
	res.InsertSorted(p.Target.RequiredCells(int(p.Row), p.TargetContext))
	//
	return res
}

// InterleavingConstraint declares a constraint that one expression represents the
// interleaving of one or more expressions.  For example, suppose X=[1,2] and
// Y=[3,4].  Then Z=[1,3,2,4] is the interleaving of X and Y.
type InterleavingConstraint[E ir.Evaluable] struct {
	Handle string
	// Context in which all target columns are evaluated.
	TargetContext schema.ModuleId
	// Context in which all source columns are evaluated.
	SourceContext schema.ModuleId
	// Target expression of interleaving.
	Target E
	// Source expressions of interleaving.
	Sources []E
}

// NewInterleavingConstraint creates a new Interleave
func NewInterleavingConstraint[E ir.Evaluable](handle string, targetContext schema.ModuleId,
	sourceContext schema.ModuleId, target E, sources []E) InterleavingConstraint[E] {
	//
	return InterleavingConstraint[E]{handle, targetContext, sourceContext, target, sources}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p InterleavingConstraint[E]) Consistent(schema schema.AnySchema) []error {
	// TODO: check column access, and widths, etc.
	return nil
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p InterleavingConstraint[E]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p InterleavingConstraint[E]) Contexts() []schema.ModuleId {
	return []schema.ModuleId{p.TargetContext, p.SourceContext}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p InterleavingConstraint[E]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Accepts checks whether a Interleave holds between the source and
// target columns.
func (p InterleavingConstraint[E]) Accepts(tr trace.Trace) (bit.Set, schema.Failure) {
	var (
		coverage  bit.Set
		srcModule = tr.Module(p.SourceContext)
		tgtModule = tr.Module(p.TargetContext)
		// Determine height of enclosing module for source columns
		tgtHeight = tr.Module(p.TargetContext).Height()
		//
		n = len(p.Sources)
	)
	//
	for row := range int(tgtHeight) {
		// Evaluate target on target row
		t, t_err := p.Target.EvalAt(row, tgtModule)
		// Evaluate next source on kth row
		s, s_err := p.Sources[row%n].EvalAt(row/n, srcModule)
		// Checks
		if t_err != nil {
			return coverage, &InternalFailure{
				p.Handle, p.TargetContext, uint(row), p.Target, t_err.Error(),
			}
		} else if s_err != nil {
			return coverage, &InternalFailure{
				p.Handle, p.SourceContext, uint(row / n), p.Sources[row%n], s_err.Error(),
			}
		} else if t.Cmp(&s) != 0 {
			// Evaluation failure
			return coverage, &InterleavingFailure{
				p.Handle,
				p.TargetContext,
				p.Target,
				p.SourceContext,
				p.Sources[row%n],
				uint(row),
			}
		}
	}
	// Success
	return coverage, nil
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p InterleavingConstraint[E]) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		sourceModule = schema.Module(p.SourceContext)
		targetModule = schema.Module(p.TargetContext)
		sources      = sexp.EmptyList()
	)
	// Iterate source expressions
	for i := range p.Sources {
		sources.Append(p.Sources[i].Lisp(sourceModule))
	}
	// Iterate target expression
	target := p.Target.Lisp(targetModule)
	// Done
	if p.Handle == "" {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("interleave"),
			target,
			sources,
		})
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("interleave"),
		sexp.NewSymbol(p.Handle),
		target,
		sources,
	})
}

// Subdivide implementation for the FieldAgnosticConstraint interface.
func (p InterleavingConstraint[E]) Subdivide(bandwidth uint, maxRegisterWidth uint) InterleavingConstraint[E] {
	return p
}

// Substitute any matchined labelled constants within this constraint
func (p InterleavingConstraint[E]) Substitute(mapping map[string]fr.Element) {
	for _, s := range p.Sources {
		s.Substitute(mapping)
	}
	//
	p.Target.Substitute(mapping)
}
