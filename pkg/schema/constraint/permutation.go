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
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// PermutationFailure provides structural information about a failing permutation constraint.
type PermutationFailure struct {
	Msg string
}

// Message provides a suitable error message
func (p *PermutationFailure) Message() string {
	return p.Msg
}

func (p *PermutationFailure) String() string {
	return p.Msg
}

// PermutationConstraint declares a constraint that one (or more) columns are a permutation
// of another.
type PermutationConstraint struct {
	Handle string
	// Evaluation Context for this constraint which must match that of the
	// source and target expressions.
	Context tr.Context
	// Targets returns the indices of the columns composing the "left" table of the
	// permutation.
	Targets []uint
	// Sources returns the indices of the columns composing the "right" table of the
	// permutation.
	Sources []uint
}

// NewPermutationConstraint creates a new permutation
func NewPermutationConstraint(handle string, context tr.Context, targets []uint,
	sources []uint) *PermutationConstraint {
	if len(targets) != len(sources) {
		panic("differeng number of target / source permutation columns")
	}

	return &PermutationConstraint{handle, context, targets, sources}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p *PermutationConstraint) Name() (string, uint) {
	return p.Handle, 0
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p *PermutationConstraint) Contexts() []tr.Context {
	return []tr.Context{p.Context}
}

// Branches returns the total number of logical branches this constraint can
// take during evaluation.
func (p *PermutationConstraint) Branches() uint {
	return 1
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *PermutationConstraint) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Accepts checks whether a permutation holds between the source and
// target columns.
func (p *PermutationConstraint) Accepts(trace tr.Trace) (bit.Set, sc.Failure) {
	// Coverage currently always empty for permutation constraints.
	var coverage bit.Set
	// Slice out data
	src := sliceColumns(p.Sources, trace)
	dst := sliceColumns(p.Targets, trace)
	// Sanity check whether column exists
	if util.ArePermutationOf(dst, src) {
		// Success
		return coverage, nil
	}
	// Prepare suitable error message
	src_names := tr.QualifiedColumnNamesToCommaSeparatedString(p.Sources, trace)
	dst_names := tr.QualifiedColumnNamesToCommaSeparatedString(p.Targets, trace)
	//
	msg := fmt.Sprintf("Target columns (%s) not permutation of source columns (%s)",
		dst_names, src_names)
	// Done
	return coverage, &PermutationFailure{msg}
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *PermutationConstraint) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	sources := sexp.EmptyList()

	for _, tid := range p.Targets {
		target := schema.Columns().Nth(tid)
		targets.Append(sexp.NewSymbol(target.QualifiedName(schema)))
	}

	for _, sid := range p.Sources {
		source := schema.Columns().Nth(sid)
		sources.Append(sexp.NewSymbol(source.QualifiedName(schema)))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("permutation"),
		targets,
		sources,
	})
}

func sliceColumns(columns []uint, tr tr.Trace) []field.FrArray {
	// Allocate return array
	cols := make([]field.FrArray, len(columns))
	// Slice out the data
	for i, n := range columns {
		nth := tr.Column(n)
		// Copy over
		cols[i] = nth.Data()
	}
	// Done
	return cols
}
