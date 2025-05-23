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

// Any converts a concrete schema into a generic view of the schema.
func Any[C Constraint](schema Schema[C]) AnySchema {
	return schema.(Schema[Constraint])
}

// AnySchema captures a generic view of a schema, which is useful in situations
// where exactly details about the schema are not important.
type AnySchema = Schema[Constraint]

// ============================================================================

// Expander functions are responsible for "filling" traces according to a given
// schema.  More specifically, the determine values for all computed columns.
type Expander[M any, C any] func(Schema[C], trace.Trace) trace.Trace

// ============================================================================

// Schema provides a fundamental interface which attempts to capture the essence
// of an arithmetisation.  For simplicity, a schema consists entirely of one or
// more modules, where each module comprises some number of registers,
// constraints and assignments.  Registers can be loosely thought of as columns
// in the final trace, whilst constraints are properties which should hold for
// any acceptable trace.  Finally, assignments represent arbitrary computations
// which "assign" values to registers during "trace expansion".
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
	// Access a given module in this schema.
	Module(module uint) Module
	// Modules returns an iterator over the declared set of modules within this
	// schema.
	Modules() iter.Iterator[Module]
	// Returns the number of modules in this schema.
	Width() uint
}
