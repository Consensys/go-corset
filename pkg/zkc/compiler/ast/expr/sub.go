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

// Sub represents an expresion which subtracts zero or more terms from a given term.
type Sub[I symbol.Symbol[I]] struct {
	poswidth uint
	negwidth uint
	Exprs    []Expr[I]
}

// NewSub constructs an expression representing the subtraction of one or more
// values.
func NewSub[I symbol.Symbol[I]](exprs ...Expr[I]) Expr[I] {
	if len(exprs) == 0 {
		panic("one or more subexpressions required")
	}
	//
	return &Sub[I]{Exprs: exprs, poswidth: math.MaxUint, negwidth: math.MaxUint}
}

// BitWidth implementation for Expr interface
func (p *Sub[I]) BitWidth() uint {
	if p.poswidth == math.MaxUint {
		panic("untyped expression")
	}
	//
	return p.poswidth
}

// NegWidth returns the negative bitwidth for this expression.
func (p *Sub[I]) NegWidth() uint {
	if p.negwidth == math.MaxUint {
		panic("untyped expression")
	}
	//
	return p.negwidth
}

// SetBitWidths sets the negative and positive bitwidths.
func (p *Sub[I]) SetBitWidths(negwidth, poswidth uint) {
	p.poswidth = poswidth
	p.negwidth = negwidth
}

// ExternUses implementation for the Expr interface.
func (p *Sub[I]) ExternUses() set.AnySortedSet[I] {
	return externUses(p.Exprs...)
}

// LocalUses implementation for the Expr interface.
func (p *Sub[I]) LocalUses() bit.Set {
	return localUses(p.Exprs...)
}

func (p *Sub[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
