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
	// Add two words together, producing another
	Add(W) W
	// Return the value of this word as a big integer.
	BigInt() *big.Int
	// Multiply two words together, producing another
	Mul(W) W
	// Construct a fresh word with the given uint64 value, or panic (if the
	// value does not fit).
	SetUint64(uint64) W
	// Returns value of word as an unsigned integer and will panic if the value
	// does not fit.
	Uint64() uint64
	// Text returns the given word formated in the given base
	Text(base int) string
}

// Uint64 initialises a given word with a 64bit value.  This will panic if the
// given value exceeds the available bandwidth of the word in question.
func Uint64[W Word[W]](val uint64) W {
	var w W
	return w.SetUint64(val)
}
