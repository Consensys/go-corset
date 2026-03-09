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

// Mul represents an expresion which computes the product of one or more terms.
type Mul[S symbol.Symbol[S]] struct {
	bitwidth uint
	Exprs    []Expr[S]
}

// NewMul constructs an expression representing the product of one or more
// values.
func NewMul[S symbol.Symbol[S]](exprs ...Expr[S]) Expr[S] {
	if len(exprs) == 0 {
		panic("one or more subexpressions required")
	}
	//
	return &Mul[S]{Exprs: exprs, bitwidth: math.MaxUint}
}

// BitWidth implementation for Expr interface
func (p *Mul[S]) BitWidth() uint {
	if p.bitwidth == math.MaxUint {
		panic("untyped expression")
	}
	//
	return p.bitwidth
}

// SetBitWidth sets the (positive) bitwidth.
func (p *Mul[S]) SetBitWidth(bitwidth uint) {
	p.bitwidth = bitwidth
}

// ExternUses implementation for the Expr interface.
func (p *Mul[S]) ExternUses() set.AnySortedSet[S] {
	return externUses(p.Exprs...)
}

// LocalUses implementation for the Expr interface.
func (p *Mul[S]) LocalUses() bit.Set {
	return localUses(p.Exprs...)
}

func (p *Mul[S]) String(mapping variable.Map[S]) string {
	return String[S](p, mapping)
}
