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

// Add implementation for Word interface.
func (p Uint) Add(w Uint) Uint {
	var res big.Int
	res.Add(&p.value, &w.value)
	//
	return Uint{res}
}

// Cmp implementation for Word interface.
func (p Uint) Cmp(o Uint) int {
	return p.value.Cmp(&o.value)
}

// BigInt implementation for Word interface.
func (p Uint) BigInt() *big.Int {
	return &p.value
}

// Mul implementation for Word interface.
func (p Uint) Mul(w Uint) Uint {
	var res big.Int
	res.Mul(&p.value, &w.value)
	//
	return Uint{res}
}

// Shr64 implementation for Word interface.
func (p Uint) Shr64(n uint64) Uint {
	var val big.Int
	val.Rsh(&p.value, uint(n))
	//
	return Uint{val}
}

// Slice implementation for Word interface.
func (p Uint) Slice(width uint) Uint {
	val := readBitSlice(0, width, p.value, true)
	return Uint{val}
}

// Uint64 implementation for Word interface.
func (p Uint) Uint64() uint64 {
	if p.value.IsUint64() {
		return p.value.Uint64()
	}
	//
	panic(fmt.Sprintf("word cannot be expressed as uint64 (0x%s)", p.value.Text(16)))
}

// SetUint64 assigns a given big integer to this unsigned integer.
func (p Uint) SetUint64(val uint64) Uint {
	var w big.Int
	w.SetUint64(val)
	//
	return Uint{w}
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

// ReadBitSlice reads a slice of bits starting at a given offset in a give
// value.  For example, consider the value is 10111000 and we have offset=1 and
// width=4, then the result is 1100.
func readBitSlice(offset uint, width uint, value big.Int, sign bool) big.Int {
	var (
		slice big.Int
		bit   uint
		n     = int(offset + width)
		m     = value.BitLen()
		i     = int(offset)
		j     = 0
	)
	// Read bits upto end
	for ; i < min(n, m); i, j = i+1, j+1 {
		// Read appropriate bit
		bit = value.Bit(i)
		// set appropriate bit
		slice.SetBit(&slice, j, bit)
	}
	// Sign extend (negative values)
	if !sign {
		// Negative value
		for ; i < n; i, j = i+1, j+1 {
			// set appropriate bit
			slice.SetBit(&slice, j, 1)
		}
	}
	//
	return slice
}
