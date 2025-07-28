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
package lookup

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Vector encapsulates all columns on one side of a lookup (i.e. it
// represents all source columns or all target columns).
type Vector[E ir.Evaluable] struct {
	// Module in which all terms are evaluated.
	Module schema.ModuleId
	// Selector for this vector (optional)
	Selector util.Option[E]
	// Terms making up this vector.
	Terms []E
}

// NewLookupVector constructs a new vector in a given context with an optional selector.
func NewLookupVector[E ir.Evaluable](mid schema.ModuleId, selector util.Option[E], terms ...E) Vector[E] {
	if selector.HasValue() {
		return FilteredLookupVector(mid, selector.Unwrap(), terms...)
	}
	//
	return UnfilteredLookupVector(mid, terms...)
}

// UnfilteredLookupVector constructs a new vector in a given context which has no selector.
func UnfilteredLookupVector[E ir.Evaluable](mid schema.ModuleId, terms ...E) Vector[E] {
	return Vector[E]{
		mid,
		util.None[E](),
		terms,
	}
}

// FilteredLookupVector constructs a new vector in a given context which has a selector.
func FilteredLookupVector[E ir.Evaluable](mid schema.ModuleId, selector E, terms ...E) Vector[E] {
	return Vector[E]{
		mid,
		util.Some(selector),
		terms,
	}
}

// Bounds determines the well-definedness bounds for all terms within this vector.
//
//nolint:revive
func (p *Vector[E]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	//
	if module == p.Module {
		// Include bounds for selector (if applicable)
		if p.HasSelector() {
			sel := p.Selector.Unwrap().Bounds()
			bound.Union(&sel)
		}
		// Include bounds for all terms
		for _, e := range p.Terms {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	}
	//
	return bound
}

// Context returns the conterxt in which all terms of this vector must be
// evaluated.
func (p *Vector[E]) Context() schema.ModuleId {
	return p.Module
}

// HasSelector determines whether or not this lookup vector has a selector or
// not.
func (p *Vector[E]) HasSelector() bool {
	return p.Selector.HasValue()
}

// Ith returns the ith term in this vector.
func (p *Vector[E]) Ith(index uint) E {
	return p.Terms[index]
}

// Len returns the number of items in this lookup vector.  Note this doesn't
// include the selector (since this is optional anyway).
func (p *Vector[E]) Len() uint {
	return uint(len(p.Terms))
}

// Lisp returns a textual representation of this vector.
func (p *Vector[E]) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		module = schema.Module(p.Module)
		terms  = sexp.EmptyList()
	)
	//
	if p.HasSelector() {
		terms.Append(p.Selector.Unwrap().Lisp(module))
	} else {
		terms.Append(sexp.NewSymbol("_"))
	}
	// Iterate source expressions
	for i := range p.Len() {
		terms.Append(p.Ith(i).Lisp(module))
	}
	// Done
	return terms
}

// Substitute any matchined labelled constants within this vector
func (p *Vector[E]) Substitute(mapping map[string]fr.Element) {
	for _, ith := range p.Terms {
		ith.Substitute(mapping)
	}
	// Substitute through selector (if applicable)
	if p.HasSelector() {
		p.Selector.Unwrap().Substitute(mapping)
	}
}
