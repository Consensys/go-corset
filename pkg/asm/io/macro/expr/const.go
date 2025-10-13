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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// Const represents a constant value within an expresion.
type Const struct {
	Label    string
	Constant big.Int
	Base     uint
}

// Eval implementation for the Expr interface.
func (p *Const) Eval([]big.Int) big.Int {
	return p.Constant
}

// Polynomial implementation for the Expr interface.
func (p *Const) Polynomial() agnostic.StaticPolynomial {
	var (
		monomial = poly.NewMonomial[io.RegisterId](p.Constant)
		result   agnostic.StaticPolynomial
	)
	//
	return result.Set(monomial)
}

// RegistersRead implementation for the Expr interface.
func (p *Const) RegistersRead() bit.Set {
	var empty bit.Set
	return empty
}

func (p *Const) String(mapping schema.RegisterMap) string {
	return String(p, mapping)
}

// ValueRange implementation for the Expr interface.
func (p *Const) ValueRange(mapping schema.RegisterMap) math.Interval {
	// Return as interval
	return math.NewInterval(p.Constant, p.Constant)
}
