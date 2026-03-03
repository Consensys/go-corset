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

// LocalAccess represents a register access within an expresion.
type LocalAccess[I symbol.Symbol[I]] struct {
	bitwidth uint
	Variable variable.Id
}

// NewLocalAccess constructs an expression representing a register access.
func NewLocalAccess[I symbol.Symbol[I]](variable variable.Id) Expr[I] {
	return &LocalAccess[I]{Variable: variable, bitwidth: math.MaxUint}
}

// BitWidth implementation for Expr interface
func (p *LocalAccess[I]) BitWidth() uint {
	if p.bitwidth == math.MaxUint {
		panic("untyped expression")
	}
	//
	return p.bitwidth
}

// SetBitWidth sets the (positive) bitwidth.
func (p *LocalAccess[I]) SetBitWidth(bitwidth uint) {
	p.bitwidth = bitwidth
}

// NonLocalUses implementation for the Expr interface.
func (p *LocalAccess[I]) NonLocalUses() set.AnySortedSet[I] {
	return nil
}

// LocalUses implementation for the Expr interface.
func (p *LocalAccess[I]) LocalUses() bit.Set {
	var read bit.Set
	read.Insert(p.Variable)
	//
	return read
}

func (p *LocalAccess[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
