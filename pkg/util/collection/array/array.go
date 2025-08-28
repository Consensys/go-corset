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
package array

import (
	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/word"
)

// Predicate abstracts the notion of a function which identifies something.
type Predicate[T any] = func(T) bool

// Array provides a generice interface to an array of elements.  Typically, we
// are interested in arrays of field elements here.
type Array[T any] interface {
	// Return the number of bits required to store an element of this array.
	BitWidth() uint
	// Clone this array producing a mutable copy
	Clone() MutArray[T]
	// Encode returns the byte encoding of this array.
	Encode() Encoding
	// Get returns the element at the given index in this array.
	Get(uint) T
	// Returns the number of elements in this array.
	Len() uint
	// Slice out a subregion of this array.
	Slice(uint, uint) Array[T]
}

// MutArray provides a generice interface to an array of elements.  Typically, we
// are interested in arrays of field elements here.
type MutArray[T any] interface {
	Array[T]
	// Append new element onto the end of array.
	Append(T)
	// Set the element at the given index in this array, overwriting the
	// original value.
	Set(uint, T)
	// Insert n copies of T at start of the array and m copies at the back
	// producing an updated array.
	Pad(uint, uint, T)
}

// Decode reconstructs an array from an array encoding, given the pool as it was
// when the encoding was made.
func Decode[K any, T word.Word[T], P pool.Pool[K, T]](encoding Encoding, p P) MutArray[T] {
	switch encoding.Encoding {
	case 1:
		var arr PoolArray[uint8, T, pool.SmallPool[uint8, T]]
		arr.Decode(encoding.Bytes)
		//
		return &arr
	default:
		panic("todo")
	}
}
