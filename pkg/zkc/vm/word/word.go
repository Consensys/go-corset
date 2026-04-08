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

import "math/big"

// Word abstracts the data type (a.k.a the "machine word") used for holding
// values within the machine.  The reason for abstracting this concept is to
// allow a machine compiled for a larger word size to be automatically lowered
// to a machine for a smaller word size.  For example, our source program might
// be written for a 64bit machine and we wish to executed it on 16bit machine
// (i.e. because our target field configuration has a maximum register size of
// 16bits).
type Word[W any] interface {
	// Add two words together, producing another (along with an overflow bit).
	Add(uint, W) (W, bool)
	// Bitwise AND of two words.
	And(uint, W) W
	// Div divides this word by another within the given bit width.  Panics on
	// division by zero.
	Div(uint, W) W
	// Return the value of this word as a big integer.
	BigInt() *big.Int
	// Cmp returns 1 if x > y, 0 if x = y, and -1 if x < y.
	Cmp(y W) int
	// Multiply two words together, producing another (along with an overflow bit).
	Mul(uint, W) (W, bool)
	// Bitwise NOT of this word within the given bit width.
	Not(uint) W
	// Bitwise OR of two words.
	Or(uint, W) W
	// Rem computes the remainder of dividing this word by another within the
	// given bit width.  Panics on division by zero.
	Rem(uint, W) W
	// Shift left word by the amount given in another word, masking to width bits.
	Shl(uint, W) W
	// Shift left word by the amount given in another word, masking to width bits.
	Shl64(uint, uint64) W
	// Shift right word by the amount given in another word.
	Shr(uint, W) W
	// Shift right word by a given number of bits.
	Shr64(uint64) W
	// Slice number of bits from this word.
	Slice(uint) W
	// Construct a fresh word with the given uint64 value, or panic (if the
	// value does not fit).
	SetUint64(uint64) W
	// Sub two words together, producing another (along with an underflow bit).
	Sub(uint, W) (W, bool)
	// Returns value of word as an unsigned integer and will panic if the value
	// does not fit.
	Uint64() uint64
	// Bitwise XOR of two words.
	Xor(uint, W) W
	// Text returns the given word formated in the given base
	Text(base int) string
}

// Uint64 initialises a given word with a 64bit value.  This will panic if the
// given value exceeds the available bandwidth of the word in question.
func Uint64[W Word[W]](val uint64) W {
	var w W
	return w.SetUint64(val)
}

// Sum a given set of words together.
func Sum[W Word[W]](bitwidth uint, values ...W) (W, bool) {
	var (
		res      W
		overflow bool
	)
	//
	for i, v := range values {
		var carry bool
		//
		if i == 0 {
			res = v
		} else {
			res, carry = res.Add(bitwidth, v)
			//
			overflow = overflow || carry
		}
	}
	//
	return res, overflow
}

// Subtract a given set of words together, producing the difference and an
// underflow indicator.
func Subtract[W Word[W]](bitwidth uint, values ...W) (W, bool) {
	var (
		res       W
		underflow bool
	)
	//
	for i, v := range values {
		var borrow bool
		//
		if i == 0 {
			res = v
		} else {
			res, borrow = res.Sub(bitwidth, v)
			//
			underflow = underflow || borrow
		}
	}
	//
	return res, underflow
}

// BitwiseAnd computes the bitwise AND of a set of words.
func BitwiseAnd[W Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.And(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseOr computes the bitwise OR of a set of words.
func BitwiseOr[W Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Or(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseXor computes the bitwise XOR of a set of words.
func BitwiseXor[W Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Xor(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseShl computes a left-shift chain over a set of words.
func BitwiseShl[W Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Shl(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseShr computes a right-shift chain over a set of words.
func BitwiseShr[W Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Shr(bitwidth, v)
		}
	}
	//
	return res
}

// Quotient divides a sequence of words left-to-right.
func Quotient[W Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Div(bitwidth, v)
		}
	}
	//
	return res
}

// Remainder computes the remainder of dividing a sequence of words left-to-right.
func Remainder[W Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Rem(bitwidth, v)
		}
	}
	//
	return res
}

// Product mulitplies a given set of words together.
func Product[W Word[W]](bitwidth uint, values ...W) (W, bool) {
	var (
		res      W
		overflow bool
	)
	//
	for i, v := range values {
		var carry bool

		if i == 0 {
			res = v
		} else {
			res, carry = res.Mul(bitwidth, v)
			//
			overflow = overflow || carry
		}
	}
	//
	return res, overflow
}
