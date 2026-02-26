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
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Const represents a constant value within an expresion.
type Const struct {
	Label    string
	Constant big.Int
	Base     uint
}

// NewConstant constructs an expression representing a constant value, along with a
// base (which is used for pretty printing, etc).
func NewConstant(constant big.Int, base uint) Expr {
	return &Const{Constant: constant, Base: base}
}

// Equals implementation for the Expr interface.
func (p *Const) Equals(e Expr) bool {
	if e, ok := e.(*Const); ok {
		return p.Constant.Cmp(&e.Constant) == 0
	}
	//
	return false
}

// Uses implementation for the Expr interface.
func (p *Const) Uses() bit.Set {
	var empty bit.Set
	return empty
}

func (p *Const) String(mapping variable.Map) string {
	return String(p, mapping)
}

// ValueRange implementation for the Expr interface.
func (p *Const) ValueRange(env variable.Map) math.Interval {
	// Return as interval
	return math.NewInterval(p.Constant, p.Constant)
}
