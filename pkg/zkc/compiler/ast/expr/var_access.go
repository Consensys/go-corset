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

// VarAccess represents a register access within an expresion.
type VarAccess struct {
	Variable variable.Id
}

// NewVarAccess constructs an expression representing a register access.
func NewVarAccess(variable variable.Id) Expr {
	return &VarAccess{Variable: variable}
}

// Equals implementation for the Expr interface.
func (p *VarAccess) Equals(e Expr) bool {
	if e, ok := e.(*VarAccess); ok {
		return p.Variable == e.Variable
	}
	//
	return false
}

// Uses implementation for the Expr interface.
func (p *VarAccess) Uses() bit.Set {
	var read bit.Set
	read.Insert(p.Variable)
	//
	return read
}

func (p *VarAccess) String(mapping variable.Map) string {
	return String(p, mapping)
}

// ValueRange implementation for the Expr interface.
func (p *VarAccess) ValueRange(env variable.Map) math.Interval {
	var (
		bound    = big.NewInt(2)
		bitwidth = env.Variable(p.Variable).BitWidth()
	)
	// compute 2^bitwidth
	bound.Exp(bound, big.NewInt(int64(bitwidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, &biONE)
	// Done
	return math.NewInterval(biZERO, *bound)
}
