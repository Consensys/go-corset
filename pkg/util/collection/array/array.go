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

// Predicate abstracts the notion of a function which identifies something.
type Predicate[T any] func(T) bool

// Array provides a generice interface to an array of elements.  Typically, we
// are interested in arrays of field elements here.
type Array[T any] interface {
	// Returns the number of elements in this array.
	Len() uint
	// Get returns the element at the given index in this array.
	Get(uint) T
	// Return the number of bits required to store an element of this array.
	BitWidth() uint
	// Slice out a subregion of this array.
	Slice(uint, uint) Array[T]
}

// MutArray provides a generice interface to an array of elements.  Typically, we
// are interested in arrays of field elements here.
type MutArray[T any] interface {
	Array[T]
	// Clone makes clones of this array producing an otherwise identical copy.
	Clone() MutArray[T]
	// Set the element at the given index in this array, overwriting the
	// original value.
	Set(uint, T)
	// Insert n copies of T at start of the array and m copies at the back
	// producing an updated array.
	Pad(uint, uint, T) MutArray[T]
}

// Builder represents a general mechanism for construct arrays.  This helps to
// separate the construction of an array from its continued existence.  Thus,
// any additional data required for its construction can be discarded once the
// array is built.
type Builder[T any] interface {
	// Fix the element at the given index in array being constructed.
	Set(uint, T)
	// Build constructs the final array.  After this point, the builder should
	// be discarded.
	Build() Array[T]
}
