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
package word

import (
	"fmt"
	"math/big"
)

// Uint represents an unbound unsigned integer.
type Uint struct {
	value big.Int
}

// BigInt implementation for Word interface.
func (p Uint) BigInt() *big.Int {
	return &p.value
}

// Uint64 implementation for Word interface.
func (p Uint) Uint64() uint64 {
	if p.value.IsUint64() {
		return p.value.Uint64()
	}
	//
	panic(fmt.Sprintf("word cannot be expressed as uint64 (0x%s)", p.value.Text(16)))
}

// SetBigInt assigns a given big integer to this unsigned integer; observe that
// this will panic if the given big integer is negative.
func (p Uint) SetBigInt(val *big.Int) Uint {
	// Sanity check
	if val.Sign() < 0 {
		panic("cannot assign negatve integer")
	}
	// Assign
	p.value = *val

	return p
}

// Text implementation for Word interface
func (p Uint) Text(base int) string {
	return p.value.Text(base)
}
