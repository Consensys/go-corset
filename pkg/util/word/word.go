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

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
)

// Word abstracts a sequence of n bits.
type Word[T any] interface {
	fmt.Stringer
	hash.Hasher[T]
	// Get the bit at a given bit offset in this word, where offsets always
	// start with the least-significant bit.
	Bit(uint) bool
	// Return minimal number of bytes required to store this word.  This can be
	// defined as the length of bytes of this word, with all leading zero bytes
	// removed.  For example, 0x1010 has a length of 2, 0x0010 has a length of 1
	// whilst 0x0000 has a byte length of 0.
	ByteWidth() uint
	// Bytes returns the bytes of this word
	Bytes() []byte
	// Write contents of this word into given byte array.  If the given byte
	// array is not big enough, a new array is allocated and returned.
	Put([]byte) []byte
	// Initialise this word from a set of bytes.
	Set([]byte) T
}

// NewArray constructs a new word array with a given capacity.
func NewArray[T Word[T], P Pool[uint, T]](height uint, bitwidth uint, pool P) array.Builder[T] {
	switch {
	case bitwidth == 0:
		return NewZeroArray[T](height)
	case bitwidth == 1:
		return NewBitArray[T](height)
	case bitwidth < 64:
		return NewStaticArray[T](height, bitwidth)
	default:
		return NewIndexArray[T, P](height, bitwidth, pool)
	}
}
