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
package expr

import (
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// NonLocalAccess represents a reference to a non-local variable, such as a
// named constant or memory.
type NonLocalAccess[I symbol.Symbol[I]] struct {
	Name I
}

// NewNonLocalAccess constructs an expression representing a non-local access,
// such as for a named constant or memory.
func NewNonLocalAccess[I symbol.Symbol[I]](name I) Expr[I] {
	return &NonLocalAccess[I]{name}
}

// NonLocalUses implementation for the Expr interface.
func (p *NonLocalAccess[I]) NonLocalUses() set.AnySortedSet[I] {
	return *set.NewAnySortedSet(p.Name)
}

// LocalUses implementation for the Expr interface.
func (p *NonLocalAccess[I]) LocalUses() bit.Set {
	var empty bit.Set
	return empty
}

func (p *NonLocalAccess[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}

// ValueRange implementation for the Expr interface.
func (p *NonLocalAccess[I]) ValueRange(env variable.Map) math.Interval {
	panic("todo")
}
