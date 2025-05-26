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
type AnySchema = Schema[Module, Constraint]

// Any converts a concrete schema into a generic view of the schema.
func Any[M Module, C Constraint](schema Schema[M, C]) AnySchema {
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
type Expander[M any, C any] func(Schema[M, C], trace.Trace) trace.Trace

// ============================================================================

type Schema[M any, C any] interface {
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
	Module(module uint) M
	// Modules returns an iterator over the declared set of modules within this
	// schema.
	Modules() iter.Iterator[M]
	// Returns the number of modules in this schema.
	Width() uint
}

// Empty constructs an empty schema.
func Empty[M any, C any](expander Expander[M, C]) Schema[M, C] {
	panic("got here")
}
