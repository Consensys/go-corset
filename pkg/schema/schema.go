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

// AnySchema captures a generic view of a schema, which is useful in situations
// where exactly details about the schema are not important.
type AnySchema = Schema[Constraint]

// Any converts a concrete schema into a generic view of the schema.
func Any[C Constraint](schema Schema[C]) AnySchema {
	// var (
	// 	modules     []Module
	// 	constraints []Constraint
	// )
	// //
	// for _, m := range schema.modules {
	// 	modules = append(modules, m)
	// }
	// for _, c := range schema.constraints {
	// 	constraints = append(constraints, c)
	// }
	// //
	// return AnySchema{modules, constraints, schema.expander}
	panic("got here")
}

// ============================================================================

// Expander functions are responsible for "filling" traces according to a given
// schema.  More specifically, the determine values for all computed columns.
type Expander[M any, C any] func(Schema[C], trace.Trace) trace.Trace

// ============================================================================

type Schema[C any] interface {
	// Assertions returns an iterator over the property assertions of this
	// schema.  These are properties which should hold true for any valid trace
	// (though, of course, may not hold true for an invalid trace).
	Assertions() iter.Iterator[C]
	// Consistent applies a number of internal consistency checks.  Whilst not
	// strictly necessary, these can highlight otherwise hidden problems as an aid
	// to debugging.
	Consistent() error
	// Constraints returns an iterator over all constraints defined in this
	// schema.
	Constraints() iter.Iterator[C]
	// Expand a given trace according to this schema by determining appropriate
	// values for all computed columns within the schema.
	Expand(trace.Trace) (trace.Trace, []error)
	// Access a given module in this schema.
	Module(module uint) Module
	// Modules returns an iterator over the declared set of modules within this
	// schema.
	Modules() iter.Iterator[Module]
	// Returns the number of modules in this schema.
	Width() uint
}

// ============================================================================

// MixedSchema represents a schema comprised of modules from different layers.
// In particular, we might have assembly and constraint (i.e. MIR) modules mixed
// together.
type MixedSchema[M1 Module, M2 Module] struct {
	left  []M1
	right []M2
}

var _ Schema[Constraint] = MixedSchema[Module, Module]{}

func NewMixedSchema[M1 Module, M2 Module](leftModules []M1, rightModules []M2) MixedSchema[M1, M2] {
	return MixedSchema[M1, M2]{leftModules, rightModules}
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p MixedSchema[M1, M2]) Assertions() iter.Iterator[Constraint] {
	panic("todo")
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p MixedSchema[M1, M2]) Consistent() error {
	// TODO: implement safety checks
	return nil
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p MixedSchema[M1, M2]) Constraints() iter.Iterator[Constraint] {
	panic("todo")
}

// Expand a given trace according to this schema by determining appropriate
// values for all computed columns within the schema.
func (p MixedSchema[M1, M2]) Expand(trace.Trace) (trace.Trace, []error) {
	panic("todo")
}

// Access a given module in this schema.
func (p MixedSchema[M1, M2]) Module(module uint) Module {
	panic("todo")
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p MixedSchema[M1, M2]) Modules() iter.Iterator[Module] {
	panic("todo")
}

// Returns the number of modules in this schema.
func (p MixedSchema[M1, M2]) Width() uint {
	panic("todo")
}
