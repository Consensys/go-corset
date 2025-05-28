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

// UniformSchema represents the simplest kind of schema which contains only
// modules of the same kind (e.g. all MIR modules).
type UniformSchema[M Module] struct {
	modules []M
}

// Sanity check
var _ Schema[Constraint] = UniformSchema[Module]{}

// NewUniformSchema constructs a new schema comprising the given modules.
func NewUniformSchema[M Module](modules []M) UniformSchema[M] {
	return UniformSchema[M]{modules}
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p UniformSchema[M]) Assertions() iter.Iterator[Constraint] {
	return iter.NewArrayIterator[Constraint](nil)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p UniformSchema[M]) Consistent() error {
	// TODO: implement safety checks
	return nil
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p UniformSchema[M]) Constraints() iter.Iterator[Constraint] {
	return constraintsOf(p.modules)
}

// Expand a given trace according to this schema by determining appropriate
// values for all computed columns within the schema.
func (p UniformSchema[M]) Expand(trace.Trace) (trace.Trace, []error) {
	panic("todo")
}

// Access a given module in this schema.
func (p UniformSchema[M]) Module(module uint) Module {
	return p.modules[module]
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p UniformSchema[M]) Modules() iter.Iterator[Module] {
	arrayIter := iter.NewArrayIterator(p.modules)
	return iter.NewCastIterator[M, Module](arrayIter)
}

// RawModules returns access to the underlying array of modules in this schema.
func (p UniformSchema[M]) RawModules() []M {
	return p.modules
}

// Returns the number of modules in this schema.
func (p UniformSchema[M]) Width() uint {
	return uint(len(p.modules))
}

// Extract an iterator over all the constraints in a given array using a
// projecting iterator.
func constraintsOf[M Module](modules []M) iter.Iterator[Constraint] {
	arrIter := iter.NewArrayIterator(modules)
	//
	return iter.NewFlattenIterator(arrIter, func(m M) iter.Iterator[Constraint] {
		return m.Constraints()
	})
}
