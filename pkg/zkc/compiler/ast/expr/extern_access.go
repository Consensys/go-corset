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
	"math"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// ExternAccess represents a reference to an external declaration, such as a
// named constant or memory.
type ExternAccess[I symbol.Symbol[I]] struct {
	bitwidth uint
	Name     I
	Args     []Expr[I]
}

// NewExternAccess constructs an expression representing a non-local access,
// such as for a named constant or memory.
func NewExternAccess[I symbol.Symbol[I]](name I, args ...Expr[I]) Expr[I] {
	return &ExternAccess[I]{Name: name, Args: args, bitwidth: math.MaxUint}
}

// BitWidth implementation for Expr interface
func (p *ExternAccess[I]) BitWidth() uint {
	if p.bitwidth == math.MaxUint {
		panic("untyped expression")
	}
	//
	return p.bitwidth
}

// SetBitWidth sets the (positive) bitwidth.
func (p *ExternAccess[I]) SetBitWidth(bitwidth uint) {
	p.bitwidth = bitwidth
}

// ExternUses implementation for the Expr interface.
func (p *ExternAccess[I]) ExternUses() set.AnySortedSet[I] {
	var uses = externUses(p.Args...)
	//
	uses.Insert(p.Name)
	//
	return uses
}

// LocalUses implementation for the Expr interface.
func (p *ExternAccess[I]) LocalUses() bit.Set {
	return localUses(p.Args...)
}

func (p *ExternAccess[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
