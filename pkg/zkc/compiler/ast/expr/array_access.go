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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// ArrayAccess represents an index into an array of some kind.
type ArrayAccess[I symbol.Symbol[I]] struct {
	Source Expr[I]
	Index  Expr[I]
}

// NewArrayAccess constructs an expression representing the sum of one or more values.
func NewArrayAccess[I symbol.Symbol[I]](src, index Expr[I]) Expr[I] {
	//
	return &ArrayAccess[I]{src, index}
}

// BitWidth implementation for Expr interface
func (p *ArrayAccess[I]) BitWidth() uint {
	// This is unsupported because, by the time it is needed, it should aleady
	// have been compiled out.
	panic("unsupported operation")
}

// NonLocalUses implementation for the Expr interface.
func (p *ArrayAccess[I]) NonLocalUses() set.AnySortedSet[I] {
	panic("todo")
}

// LocalUses implementation for the Expr interface.
func (p *ArrayAccess[I]) LocalUses() bit.Set {
	var reads bit.Set
	//
	reads.Union(p.Source.LocalUses())
	reads.Union(p.Index.LocalUses())
	//
	return reads
}

func (p *ArrayAccess[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
