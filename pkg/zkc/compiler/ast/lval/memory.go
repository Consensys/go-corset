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
package lval

import (
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// MemAccess represents a memory write within an assignment.
type MemAccess[I symbol.Symbol[I]] struct {
	// name of the memory being written
	Name I
	// index of location being written
	Index expr.Expr[I]
}

// NewMemAccess constructs an expression representing a register access.
func NewMemAccess[I symbol.Symbol[I]](name I, index expr.Expr[I]) LVal[I] {
	return &MemAccess[I]{name, index}
}

// ExternUses implementation for the LVal interface.
func (p *MemAccess[I]) ExternUses() set.AnySortedSet[I] {
	return p.Index.ExternUses()
}

// LocalUses implementation for the LVal interface.
func (p *MemAccess[I]) LocalUses() bit.Set {
	return p.Index.LocalUses()
}

// LocalDefs implementation for the LVal interface.
func (p *MemAccess[I]) LocalDefs() bit.Set {
	return bit.Set{}
}

func (p *MemAccess[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
