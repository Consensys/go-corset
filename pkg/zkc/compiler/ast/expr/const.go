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
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Const represents a constant value within an expresion.
type Const[S symbol.Symbol[S]] struct {
	Label    string
	Constant big.Int
	Base     uint
}

// NewConstant constructs an expression representing a constant value, along with a
// base (which is used for pretty printing, etc).
func NewConstant[S symbol.Symbol[S]](constant big.Int, base uint) Expr[S] {
	return &Const[S]{Constant: constant, Base: base}
}

// ExternUses implementation for the Expr interface.
func (p *Const[S]) ExternUses() set.AnySortedSet[S] {
	return nil
}

// LocalUses implementation for the Expr interface.
func (p *Const[S]) LocalUses() bit.Set {
	var empty bit.Set
	return empty
}

func (p *Const[S]) String(mapping variable.Map[S]) string {
	return String[S](p, mapping)
}

// SetType implementation for Expr interface
func (p *Const[S]) SetType(t data.Type[S]) {

}

// Type implementation for Expr interface
func (p *Const[S]) Type() data.Type[S] {
	bitwidth := uint(p.Constant.BitLen())
	return data.NewUnsignedInt[S](bitwidth, true)
}
