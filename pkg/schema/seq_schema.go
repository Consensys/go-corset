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
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

type TableSchema[M Module] struct {
}

func (p *TableSchema[M]) AddModule(name string) uint {
	panic("todo")
}

// Access a given module in this schema.
func (p *TableSchema[M]) Module(uint) Module {
	panic("todo")
}

// Returns the number of modules in this schema.
func (p *TableSchema[M]) Width() uint {
	panic("todo")
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p *TableSchema[M]) Constraints() iter.Iterator[Constraint] {
	panic("todo")
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *TableSchema[M]) Assertions() iter.Iterator[Constraint] {
	panic("todo")
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *TableSchema[M]) Modules() iter.Iterator[Module] {
	panic("todo")
}

// ========================================================================

// ========================================================================

// fixed set of data
type StaticModule struct {
}
