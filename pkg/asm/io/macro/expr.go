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
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/poly"
)

var (
	biZERO big.Int = *big.NewInt(0)
	biONE  big.Int = *big.NewInt(1)
)

// Expr represents an arbitrary expression used within an instruction.
type Expr interface {
	// Polynomial returns this expression flatterned into a polynomial form.
	Polynomial() agnostic.Polynomial
	// RegistersRead returns the set of registers read by this expression
	RegistersRead() bit.Set
	// String returns a string representation of this expression in a given base.
	String(mapping schema.RegisterMap) string
	// ValueRange returns the interval of values that this term can evaluate to.
	// For terms accessing registers, this is determined by the declared width of
	// the register.
	ValueRange(mapping schema.RegisterMap) math.Interval
}

// AddExpr represents an expresion which adds one or more terms together.
type AddExpr struct {
	Exprs []Expr
}

// Polynomial implementation for the Expr interface.
func (p *AddExpr) Polynomial() agnostic.Polynomial {
	panic("todo")
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

// ValueRange implementation for the Expr interface.
func (p *AddExpr) ValueRange(mapping schema.RegisterMap) math.Interval {
	var values math.Interval
	//
	for i, e := range p.Exprs {
		if i == 0 {
			values = e.ValueRange(mapping)
		} else {
			values.Add(e.ValueRange(mapping))
		}
	}
	//
	return values
}

func (p *AddExpr) String(schema.RegisterMap) string {
	panic("todo")
}

// ConstantExpr represents a constant value within an expresion.
type ConstantExpr struct {
	Constant big.Int
}

// Polynomial implementation for the Expr interface.
func (p *ConstantExpr) Polynomial() agnostic.Polynomial {
	var (
		monomial = poly.NewMonomial[io.RegisterId](p.Constant)
		result   agnostic.Polynomial
	)
	//
	return result.Set(monomial)
}

// RegistersRead implementation for the Expr interface.
func (p *ConstantExpr) RegistersRead() bit.Set {
	var empty bit.Set
	return empty
}

func (p *ConstantExpr) String(schema.RegisterMap) string {
	return p.Constant.String()
}

// ValueRange implementation for the Expr interface.
func (p *ConstantExpr) ValueRange(mapping schema.RegisterMap) math.Interval {
	// Return as interval
	return math.NewInterval(p.Constant, p.Constant)
}

// RegisterAccessExpr represents a register access within an expresion.
type RegisterAccessExpr struct {
	Register io.RegisterId
}

// Polynomial implementation for the Expr interface.
func (p *RegisterAccessExpr) Polynomial() agnostic.Polynomial {
	var (
		monomial = poly.NewMonomial(biONE, p.Register)
		result   agnostic.Polynomial
	)
	//
	return result.Set(monomial)
}

// RegistersRead implementation for the Expr interface.
func (p *RegisterAccessExpr) RegistersRead() bit.Set {
	var read bit.Set
	read.Insert(p.Register.Unwrap())
	//
	return read
}

func (p *RegisterAccessExpr) String(mapping schema.RegisterMap) string {
	return mapping.Register(p.Register).Name
}

// ValueRange implementation for the Expr interface.
func (p *RegisterAccessExpr) ValueRange(mapping schema.RegisterMap) math.Interval {
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
