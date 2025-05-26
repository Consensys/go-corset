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
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

type MutableSchema[M any, C any] struct {
	modules     []M
	constraints []C
	// Expander is responsible for expanding a given trace according to this
	// schema.  Specifically, given a well-formed set of input columns, this
	// means computing values for all computed columns in the schema.
	expander Expander[M, C]
}

// Add a new constraint into this schema.
func (p *MutableSchema[M, C]) Add(constraint C) {
	p.constraints = append(p.constraints, constraint)
}

func (p *MutableSchema[M, C]) AddModule(name string) uint {
	panic("todo")
}

// Expand a given trace according to this schema by computing the values for all
// computed columns.  Observe that this can result in modules having different
// heights after the expansion, for a variety of reasons.  For example, spillage
// and/or defensive padding maybe applied.  Likewise, function instances may be
// fleshed out with their full trace, etc.
func (p *MutableSchema[M, C]) Expand(tr trace.Trace) trace.Trace {
	//return p.expander(*p, tr)
	panic("todo")
}

// Access a given module in this schema.
func (p *MutableSchema[M, C]) Module(module uint) M {
	return p.modules[module]
}

// Returns the number of modules in this schema.
func (p *MutableSchema[M, C]) Width() uint {
	return uint(len(p.modules))
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p *MutableSchema[M, C]) Consistent() error {
	panic("todo")
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p *MutableSchema[M, C]) Constraints() iter.Iterator[C] {
	return iter.NewArrayIterator(p.constraints)
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *MutableSchema[M, C]) Assertions() iter.Iterator[C] {
	panic("todo")
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *MutableSchema[M, C]) Modules() iter.Iterator[M] {
	return iter.NewArrayIterator(p.modules)
}
