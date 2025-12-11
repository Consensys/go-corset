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
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
)

// RegAccess represents a register access within an expresion.
type RegAccess struct {
	Register io.RegisterId
}

// IsAtomic implementation for AtomicExpr interface
func (p *RegAccess) IsAtomic() {

}

// Eval implementation for the Expr interface.
func (p *RegAccess) Eval(env []big.Int) big.Int {
	return env[p.Register.Unwrap()]
}

// Polynomial implementation for the Expr interface.
func (p *RegAccess) Polynomial() agnostic.StaticPolynomial {
	var (
		monomial = poly.NewMonomial(biONE, p.Register)
		result   agnostic.StaticPolynomial
	)
	//
	return result.Set(monomial)
}

// RegistersRead implementation for the Expr interface.
func (p *RegAccess) RegistersRead() bit.Set {
	var read bit.Set
	read.Insert(p.Register.Unwrap())
	//
	return read
}

func (p *RegAccess) String(mapping register.Map) string {
	return String(p, mapping)
}

// ValueRange implementation for the Expr interface.
func (p *RegAccess) ValueRange(mapping register.Map) math.Interval {
	var (
		bound    = big.NewInt(2)
		bitwidth = mapping.Register(p.Register).Width
	)
	// compute 2^bitwidth
	bound.Exp(bound, big.NewInt(int64(bitwidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, &biONE)
	// Done
	return math.NewInterval(biZERO, *bound)
}
