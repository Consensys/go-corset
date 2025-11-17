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
package mir

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// FunctionCall represents a special kind of lookup constraint which triggers
// trace propagation.
type FunctionCall[F any] struct {
	Handle         string
	Callee, Caller sc.ModuleId
	Returns        []Term[F]
	Arguments      []Term[F]
	Selector       util.Option[LogicalTerm[F]]
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p FunctionCall[F]) Consistent(schema sc.AnySchema[F]) []error {
	var (
		errors []error
		nargs  = uint(len(p.Arguments))
		nrets  = uint(len(p.Returns))
		n      = nargs + nrets
		mod    = schema.Module(p.Callee)
	)
	//
	if mod.Width() < n {
		errors = append(errors,
			fmt.Errorf("incorrect number of arguments / returns (%d vs %d)", n, mod.Width()))
	} else {
		//
		for i := range n {
			var (
				id  = sc.NewRegisterId(i)
				reg = mod.Register(id)
			)
			if i < nargs && !reg.IsInput() {
				errors = append(errors,
					fmt.Errorf("inconsistent number of arguments (%d vs %d)", nargs, i))
				//
				break
			} else if i >= nargs && !reg.IsOutput() {
				errors = append(errors,
					fmt.Errorf("inconsistent number of returns (%d vs %d)", nrets, i-nargs))
				//
				break
			}
		}
	}
	//
	return errors
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p FunctionCall[F]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p FunctionCall[F]) Contexts() []sc.ModuleId {
	return []sc.ModuleId{
		p.Callee, p.Caller,
	}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p FunctionCall[F]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	//
	if module == p.Caller {
		// Include bounds for arguments
		for _, e := range p.Arguments {
			eth := e.Bounds()
			bound.Union(&eth)
		}
		// Include bounds for returns
		for _, e := range p.Returns {
			eth := e.Bounds()
			bound.Union(&eth)
		}
		// Bound selector (if applicable)
		if p.Selector.HasValue() {
			eth := p.Selector.Unwrap().Bounds()
			bound.Union(&eth)
		}
	}
	//
	return bound
}

// Accepts checks whether a lookup constraint into the target columns holds for
// all rows of the source columns.
//
//nolint:revive
func (p FunctionCall[F]) Accepts(_ trace.Trace[F], _ sc.AnySchema[F],
) (bit.Set, sc.Failure) {
	// Currently this is unreachable.
	panic("unreachable")
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p FunctionCall[F]) Lisp(mapping sc.AnySchema[F]) sexp.SExp {
	var (
		module = mapping.Module(p.Caller)
		callee = mapping.Module(p.Callee)
		args   = sexp.EmptyList()
		rets   = sexp.EmptyList()
	)
	//
	// Iterate arguments
	for _, ith := range p.Arguments {
		args.Append(ith.Lisp(true, module))
	}
	// Iterate returns
	for _, ith := range p.Returns {
		rets.Append(ith.Lisp(true, module))
	}
	// Done
	list := sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("call"),
		rets,
		sexp.NewSymbol(callee.Name()),
		args,
	})
	//
	if p.Selector.HasValue() {
		list.Append(p.Selector.Unwrap().Lisp(true, module))
	}
	//
	return list
}

// Substitute any matchined labelled constants within this constraint
func (p FunctionCall[F]) Substitute(mapping map[string]F) {
	//nolint
	for _, ith := range p.Returns {
		ith.Substitute(mapping)
	}

	for _, ith := range p.Arguments {
		ith.Substitute(mapping)
	}
	// Substitute through selector (if applicable)
	if p.Selector.HasValue() {
		p.Selector.Unwrap().Substitute(mapping)
	}
}
