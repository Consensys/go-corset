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
package macro

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Expr represents an arbitrary expression used within an instruction.
type Expr interface {
	// RegistersRead returns the set of registers read by this expression
	RegistersRead() bit.Set
}

// AddExpr represents an expresion which adds one or more terms together.
type AddExpr struct {
	Exprs []Expr
}

// RegistersRead implementation for the Expr interface.
func (p *AddExpr) RegistersRead() bit.Set {
	var reads bit.Set
	//
	for _, e := range p.Exprs {
		reads.Union(e.RegistersRead())
	}
	//
	return reads
}

// ConstantExpr represents a constant value within an expresion.
type ConstantExpr struct {
	Constant big.Int
}

// RegistersRead implementation for the Expr interface.
func (p *ConstantExpr) RegistersRead() bit.Set {
	var empty bit.Set
	return empty
}

// RegisterAccessExpr represents a register access within an expresion.
type RegisterAccessExpr struct {
	Register io.RegisterId
}

// RegistersRead implementation for the Expr interface.
func (p *RegisterAccessExpr) RegistersRead() bit.Set {
	var read bit.Set
	read.Insert(p.Register.Unwrap())
	//
	return read
}
