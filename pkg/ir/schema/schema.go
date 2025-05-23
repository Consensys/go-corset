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

type Schema[M any, C any] struct {
	modules     []M
	constraints []C
}

func (p *Schema[M, C]) AddModule(name string) uint {
	panic("todo")
}

// Access a given module in this schema.
func (p *Schema[M, C]) Module(uint) M {
	panic("todo")
}

// Returns the number of modules in this schema.
func (p *Schema[M, C]) Width() uint {
	panic("todo")
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p *Schema[M, C]) Constraints() iter.Iterator[C] {
	panic("todo")
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *Schema[M, C]) Assertions() iter.Iterator[C] {
	panic("todo")
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *Schema[M, C]) Modules() iter.Iterator[M] {
	panic("todo")
}

// Add a new constraint into this schema.
func (p *Schema[M, C]) Add(constraint C) {
	p.constraints = append(p.constraints, constraint)
}
