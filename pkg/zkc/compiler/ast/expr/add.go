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

// Add represents an expresion which adds one or more terms together.
type Add[I symbol.Symbol[I]] struct {
	bitwidth uint
	Exprs    []Expr[I]
}

// NewAdd constructs an expression representing the sum of one or more values.
func NewAdd[I symbol.Symbol[I]](exprs ...Expr[I]) Expr[I] {
	if len(exprs) == 0 {
		panic("one or more subexpressions required")
	}
	//
	return &Add[I]{Exprs: exprs, bitwidth: math.MaxUint}
}

// BitWidth implementation for Expr interface
func (p *Add[I]) BitWidth() uint {
	if p.bitwidth == math.MaxUint {
		panic("untyped expression")
	}

	return p.bitwidth
}

// SetBitWidth sets the (positive) bitwidth.
func (p *Add[I]) SetBitWidth(bitwidth uint) {
	p.bitwidth = bitwidth
}

// NonLocalUses implementation for the Expr interface.
func (p *Add[I]) NonLocalUses() set.AnySortedSet[I] {
	return nonLocalUses(p.Exprs...)
}

// LocalUses implementation for the Expr interface.
func (p *Add[I]) LocalUses() bit.Set {
	return localUses(p.Exprs...)
}

func (p *Add[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
