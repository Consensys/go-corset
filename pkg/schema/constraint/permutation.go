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

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
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
	Context trace.Context
	// Targets returns the indices of the columns composing the "left" table of the
	// permutation.
	Targets []uint
	// Sources returns the indices of the columns composing the "right" table of the
	// permutation.
	Sources []uint
}

// NewPermutationConstraint creates a new permutation
func NewPermutationConstraint(handle string, context trace.Context, targets []uint,
	sources []uint) PermutationConstraint {
	if len(targets) != len(sources) {
		panic("differeng number of target / source permutation columns")
	}

	return PermutationConstraint{handle, context, targets, sources}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p PermutationConstraint) Consistent(schema schema.AnySchema) []error {
	// TODO: check column access, and widths, etc.
	return nil
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p PermutationConstraint) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p PermutationConstraint) Contexts() []trace.Context {
	return []trace.Context{p.Context}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p PermutationConstraint) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Accepts checks whether a permutation holds between the source and
// target columns.
func (p PermutationConstraint) Accepts(tr trace.Trace) (bit.Set, schema.Failure) {
	var (
		// Coverage currently always empty for permutation constraints.
		coverage bit.Set
		// Determine enclosing module
		module trace.Module = tr.Module(p.Context.ModuleId)
	)
	// Slice out data
	src := sliceColumns(p.Sources, module)
	dst := sliceColumns(p.Targets, module)
	// Sanity check whether column exists
	if util.ArePermutationOf(dst, src) {
		// Success
		return coverage, nil
	}
	// Prepare suitable error message
	src_names := trace.QualifiedColumnNamesToCommaSeparatedString(p.Sources, module)
	dst_names := trace.QualifiedColumnNamesToCommaSeparatedString(p.Targets, module)
	//
	msg := fmt.Sprintf("Target columns (%s) not permutation of source columns (%s)",
		dst_names, src_names)
	// Done
	return coverage, &PermutationFailure{msg}
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p PermutationConstraint) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		module  = schema.Module(p.Context.ModuleId)
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)

	for _, tid := range p.Targets {
		target := module.Register(tid)
		targets.Append(sexp.NewSymbol(target.QualifiedName(module)))
	}

	for _, sid := range p.Sources {
		source := module.Register(sid)
		sources.Append(sexp.NewSymbol(source.QualifiedName(module)))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("permutation"),
		targets,
		sources,
	})
}

func sliceColumns(columns []uint, tr trace.Module) []field.FrArray {
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
