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
package hir

import (
	"encoding/gob"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// Term represents a component of an AIR expression.
type Term interface {
	util.Boundable
}

// ============================================================================
// Addition
// ============================================================================

// Add represents the addition of zero or more expressions.
type Add struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Add) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Cast
// ============================================================================

// Cast attempts to narrow the width a given expression.
type Cast struct {
	Arg      Term
	BitWidth uint
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Cast) Bounds() util.Bounds { return p.Arg.Bounds() }

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
// Connective
// ============================================================================

// Connective represents a logical connective (either conjunction or
// disjunction).
type Connective struct {
	Sign bool
	Args []Term
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Connective) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Constant
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Value fr.Element }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *Constant) Bounds() util.Bounds { return util.EMPTY_BOUND }

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

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *ColumnAccess) Bounds() util.Bounds {
	if p.Shift >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.Shift))
	}
	// Negative shift
	return util.NewBounds(uint(-p.Shift), 0)
}

// ============================================================================
// Equation
// ============================================================================

const (
	// EQUALS indicates an equals (==) relationship
	EQUALS uint8 = 0
	// NOT_EQUALS indicates a not-equals (!=) relationship
	NOT_EQUALS uint8 = 1
	// LESS_THAN indicates a less-than (<) relationship
	LESS_THAN uint8 = 2
	// LESS_THAN_EQUALS indicates a less-than-or-equals (<=) relationship
	LESS_THAN_EQUALS uint8 = 3
	// GREATER_THAN indicates a greater-than (>) relationship
	GREATER_THAN uint8 = 4
	// GREATER_THAN_EQUALS indicates a greater-than-or-equals (>=) relationship
	GREATER_THAN_EQUALS uint8 = 5
)

// Equation represents an equality (e.g. X == Y) or non-equality (e.g. X != Y)
// relationship between two terms.
type Equation struct {
	Kind uint8
	Lhs  Term
	Rhs  Term
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Equation) Bounds() util.Bounds {
	l := p.Lhs.Bounds()
	r := p.Rhs.Bounds()
	//
	l.Union(&r)
	//
	return l
}

// ============================================================================
// Exponentiation
// ============================================================================

// Exp represents the a given value taken to a power.
type Exp struct {
	Arg Term
	Pow uint64
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Exp) Bounds() util.Bounds { return p.Arg.Bounds() }

// ============================================================================
// IfZero
// ============================================================================

// IfZero returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero struct {
	// Elements contained within this list.
	Condition Term
	// True branch (optional).
	TrueBranch Term
	// False branch (optional).
	FalseBranch Term
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *IfZero) Bounds() util.Bounds {
	c := p.Condition.Bounds()
	// Get bounds for true branch (if applicable)
	if p.TrueBranch != nil {
		tbounds := p.TrueBranch.Bounds()
		c.Union(&tbounds)
	}
	// Get bounds for false branch (if applicable)
	if p.FalseBranch != nil {
		fbounds := p.FalseBranch.Bounds()
		c.Union(&fbounds)
	}
	// Done
	return c
}

// ============================================================================
// LabelledConstant
// ============================================================================

// LabelledConstant represents a constant value which is labelled with a given
// name.  The purpose of this is to allow labelled constants to be substituted
// for different values when desired.
type LabelledConstant struct {
	Label string
	Value fr.Element
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *LabelledConstant) Bounds() util.Bounds { return util.EMPTY_BOUND }

// ============================================================================
// List
// ============================================================================

// List represents a block of zero or more expressions.
type List struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *List) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Mul) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Normalise
// ============================================================================

// Norm represents the normalisation operator which reduces the value of an
// expression to either zero (if it was zero) or one (otherwise).
type Norm struct{ Arg Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Norm) Bounds() util.Bounds { return p.Arg.Bounds() }

// ============================================================================
// Not
// ============================================================================

// Not represents the logical negation of its argument.
type Not struct{ Arg Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Not) Bounds() util.Bounds { return p.Arg.Bounds() }

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Term }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Sub) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(Term(&Add{}))
	gob.Register(Term(&Mul{}))
	gob.Register(Term(&Sub{}))
	gob.Register(Term(&Cast{}))
	gob.Register(Term(&Connective{}))
	gob.Register(Term(&Equation{}))
	gob.Register(Term(&Exp{}))
	gob.Register(Term(&IfZero{}))
	gob.Register(Term(&List{}))
	gob.Register(Term(&Constant{}))
	gob.Register(Term(&LabelledConstant{}))
	gob.Register(Term(&Norm{}))
	gob.Register(Term(&Not{}))
	gob.Register(Term(&ColumnAccess{}))
}
