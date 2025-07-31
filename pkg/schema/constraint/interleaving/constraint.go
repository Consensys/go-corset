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
package interleaving

import (
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

// Constraint declares a constraint that one expression represents the
// interleaving of one or more expressions.  For example, suppose X=[1,2] and
// Y=[3,4].  Then Z=[1,3,2,4] is the interleaving of X and Y.
type Constraint[E ir.Evaluable] struct {
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

// NewConstraint creates a new Interleave
func NewConstraint[E ir.Evaluable](handle string, targetContext schema.ModuleId,
	sourceContext schema.ModuleId, target E, sources []E) Constraint[E] {
	//
	return Constraint[E]{handle, targetContext, sourceContext, target, sources}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Constraint[E]) Consistent(schema schema.AnySchema) []error {
	// TODO: check column access, and widths, etc.
	return nil
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
	return []schema.ModuleId{p.TargetContext, p.SourceContext}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p Constraint[E]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Accepts checks whether a Interleave holds between the source and
// target columns.
func (p Constraint[E]) Accepts(tr trace.Trace[bls12_377.Element], sc schema.AnySchema) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		srcTrMod = tr.Module(p.SourceContext)
		tgtTrMod = tr.Module(p.TargetContext)
		srcScMod = sc.Module(p.SourceContext)
		tgtScMod = sc.Module(p.TargetContext)
		// Determine height of enclosing module for source columns
		tgtHeight = tr.Module(p.TargetContext).Height()
		//
		n = len(p.Sources)
	)
	//
	for row := range int(tgtHeight) {
		// Evaluate target on target row
		t, t_err := p.Target.EvalAt(row, tgtTrMod, tgtScMod)
		// Evaluate next source on kth row
		s, s_err := p.Sources[row%n].EvalAt(row/n, srcTrMod, srcScMod)
		// Checks
		if t_err != nil {
			return coverage, &constraint.InternalFailure{
				Handle:  p.Handle,
				Context: p.TargetContext,
				Row:     uint(row),
				Term:    p.Target,
				Error:   t_err.Error(),
			}
		} else if s_err != nil {
			return coverage, &constraint.InternalFailure{
				Handle:  p.Handle,
				Context: p.SourceContext,
				Row:     uint(row / n),
				Term:    p.Sources[row%n],
				Error:   s_err.Error(),
			}
		} else if t.Cmp(&s) != 0 {
			// Evaluation failure
			return coverage, &Failure{
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
func (p Constraint[E]) Lisp(schema schema.AnySchema) sexp.SExp {
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

// Substitute any matchined labelled constants within this constraint
func (p Constraint[E]) Substitute(mapping map[string]fr.Element) {
	for _, s := range p.Sources {
		s.Substitute(mapping)
	}
	//
	p.Target.Substitute(mapping)
}
