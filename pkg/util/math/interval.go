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
package math

import (
	"fmt"
	"math/big"
)

// Interval provides a discrete range of integers, such as 0..1, 1..18, etc.  An
// interval can be used to approximate the possible values that a given
// expression could evaluate to.
type Interval struct {
	min big.Int
	max big.Int
}

// NewInterval creates an interval representing a given range.
func NewInterval(lower *big.Int, upper *big.Int) *Interval {
	var (
		min big.Int
		max big.Int
	)
	//
	min.Set(lower)
	max.Set(upper)
	//
	return &Interval{min, max}
}

// NewInterval64 creates an interval representing a given range.
func NewInterval64(lower int64, upper int64) *Interval {
	return NewInterval(big.NewInt(lower), big.NewInt(upper))
}

// MinValue returns the minimum value that this interval includes.
func (p *Interval) MinValue() big.Int {
	return p.min
}

// MaxValue returns the maximum value that this interval includes.
func (p *Interval) MaxValue() big.Int {
	return p.max
}

// BitWidth returns the minimum number of bits required to store all elements in
// this interval.
func (p *Interval) BitWidth() uint {
	// If range includes a negative number, then bitwidth is maximum possible.
	if p.min.Cmp(big.NewInt(0)) < 0 {
		panic("todo")
	}
	//
	return uint(p.max.BitLen())
}

// Set assigns a given value to this interval.
func (p *Interval) Set(val *Interval) {
	p.min.Set(&val.min)
	p.max.Set(&val.max)
}

// Size returns the number of elements contained within this interval.
func (p *Interval) Size() big.Int {
	var diff big.Int
	//
	diff.Set(&p.max)
	diff.Sub(&diff, &p.min)
	diff.Add(&diff, big.NewInt(1))
	//
	return diff
}

// Contains checks whether a given value is contained with this interval
func (p *Interval) Contains(val *big.Int) bool {
	return p.min.Cmp(val) <= 0 && p.max.Cmp(val) >= 0
}

// Within checks whether this interval is contained within the given bounds.
func (p *Interval) Within(other *Interval) bool {
	return p.min.Cmp(&other.min) >= 0 && p.max.Cmp(&other.max) <= 0
}

// Insert a given value into this interval
func (p *Interval) Insert(val *Interval) {
	// Check lower bound
	if p.min.Cmp(&val.min) > 0 {
		p.min.Set(&val.min)
	}
	// Check upper bound
	if p.max.Cmp(&val.max) < 0 {
		p.max.Set(&val.max)
	}
}

// Add two intervals together
func (p *Interval) Add(q *Interval) {
	p.min.Add(&p.min, &q.min)
	p.max.Add(&p.max, &q.max)
}

// Sub subtracts another interval from this.
func (p *Interval) Sub(q *Interval) {
	p.min.Sub(&p.min, &q.max)
	p.max.Sub(&p.max, &q.min)
}

// Mul multiplies this interval by another.
func (p *Interval) Mul(q *Interval) {
	var (
		x1 big.Int
		x2 big.Int
		x3 big.Int
		x4 big.Int
	)
	//
	x1.Mul(&p.min, &q.min)
	x2.Mul(&p.min, &q.max)
	x3.Mul(&p.max, &q.min)
	x4.Mul(&p.max, &q.max)
	// Compute min / max
	min := bigMin(bigMin(x1, x2), bigMin(x3, x4))
	max := bigMax(bigMax(x1, x2), bigMax(x3, x4))
	//
	p.min.Set(&min)
	p.max.Set(&max)
}

// Exp raises this interval to a fix exponent.
func (p *Interval) Exp(pow uint) {
	var val Interval
	// Clone p
	val.Set(p)
	// This can be computed more efficiently perhaps by using a recursive
	// decomposition, 2^n = 2^n/2 * 2^n/2.
	for i := uint(1); i < pow; i++ {
		p.Mul(&val)
	}
}

func (p *Interval) String() string {
	return fmt.Sprintf("(%s..%s)", p.min.String(), p.max.String())
}

func bigMin(x1 big.Int, x2 big.Int) big.Int {
	if x1.Cmp(&x2) < 0 {
		return x1
	}

	return x2
}

func bigMax(x1 big.Int, x2 big.Int) big.Int {
	if x1.Cmp(&x2) >= 0 {
		return x1
	}

	return x2
}
