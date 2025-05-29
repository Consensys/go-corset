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
package mir

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// Term represents a component of an AIR expression.
type Term interface {
	// Normalised returns true if the given term is normalised.  For example, an
	// product containing a product argument is not normalised.
	Normalised() bool
}

// ============================================================================
// Addition
// ============================================================================

// Add represents the addition of zero or more expressions.
type Add struct{ Args []Term }

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *Add) Normalised() bool {
	panic("todo")
}

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Term }

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *Sub) Normalised() bool {
	panic("todo")
}

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Term }

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *Mul) Normalised() bool {
	panic("todo")
}

// ============================================================================
// Cast
// ============================================================================

// Cast attempts to narrow the width a given expression.
type Cast struct {
	Arg      Term
	BitWidth uint
}

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *Cast) Normalised() bool {
	panic("todo")
}

// Range returns the range of values which this cast represents.
func (p *Cast) Range() *util.Interval {
	var (
		zero  = big.NewInt(0)
		bound = big.NewInt(2)
	)
	// Determine bound for static type check
	bound.Exp(bound, big.NewInt(int64(p.BitWidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	// Determine casted interval
	return util.NewInterval(zero, bound)
}

// ============================================================================
// Exponentiation
// ============================================================================

// Exp represents the a given value taken to a power.
type Exp struct {
	Arg Term
	Pow uint64
}

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *Exp) Normalised() bool {
	panic("todo")
}

// ============================================================================
// Constant
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Value fr.Element }

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *Constant) Normalised() bool {
	return true
}

// ============================================================================
// Normalise
// ============================================================================

// Norm reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Norm struct{ Arg Term }

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *Norm) Normalised() bool {
	panic("todo")
}

// ============================================================================
// ColumnAccess
// ============================================================================

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount. Suppose we are evaluating a constraint on row k=5 which
// contains the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)
// accesses the STAMP column at row 5, whilst CT(-1) accesses the CT column at
// row 4.
type ColumnAccess struct {
	Column uint
	Shift  int
}

// Normalised returns true if the given term is normalised.  For example, an
// product containing a product argument is not normalised.
func (p *ColumnAccess) Normalised() bool {
	return true
}
