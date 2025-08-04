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
package field

import (
	"github.com/consensys/go-corset/pkg/util/word"
)

// An Element of a prime-order field.
type Element[Operand any] interface {
	// Field elements are always words
	word.Word[Operand]
	// Add x+y
	Add(y Operand) Operand
	// Bytes returns the big-endian encoded Element, possibly with leading zeros.
	Bytes() []byte
	// Cmp returns 1 if x > y, 0 if x = y, and -1 if x < y.
	Cmp(y Operand) int
	// Check whether this value is zero (or not).
	IsZero() bool
	// Check whether this value is one (or not).
	IsOne() bool
	// Compute x * y
	Mul(y Operand) Operand
	// Compute x⁻¹, or 0 if x = 0.
	Inverse() Operand
	// Set this element to a uint64 value
	Set64(uint64)
	// Compute x - y
	Sub(y Operand) Operand
	// Text returns the numerical value of x in the given base.
	Text(base int) string
}
