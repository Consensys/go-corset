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
package iter

// Predicate abstracts the notion of a function which identifies something.
type Predicate[T any] func(T) bool

// Iterator is an adapter which sits on top of a BaseIterator and provides
// various useful and reusable functions.
type Iterator[T any] interface {
	Enumerator[T]

	// Append another iterator onto the end of this iterator.  Thus, when all
	// items are visited in this iterator, iteration continues into the other.
	Append(Iterator[T]) Iterator[T]

	// Clone creates a copy of this iterator at the given cursor position.
	// Modifying the clone (i.e. by calling Next) iterator will not modify the
	// original.
	Clone() Iterator[T]

	// Collect allocates a new array containing all items of this iterator.
	// This drains the iterator.
	Collect() []T

	// Find returns the index of the first match for a given predicate, or
	// return false if no match is found.  This will mutate the iterator.
	Find(Predicate[T]) (uint, bool)

	// Count the number of items left.  Note, this does not modify the iterator.
	Count() uint

	// Get the nth item in this iterator.  This will mutate the iterator.
	Nth(uint) T
}

// ===============================================================
// Base Iterator
// ===============================================================

// Find provides a default implementation of Iterator.Find which can be used by
// other iterator implementations.
//
//nolint:revive
func Find[T any, S Enumerator[T]](iter S, predicate Predicate[T]) (uint, bool) {
	index := uint(0)

	for iter.HasNext() {
		if predicate(iter.Next()) {
			return index, true
		}

		index++
	}
	// Failed to find it
	return 0, false
}

// Nth provides a default implementation of Iterator.Nth which can be used by
// other iterator implementations.
//
//nolint:revive
func Nth[T any, S Enumerator[T]](iter S, n uint) T {
	index := uint(0)

	for iter.HasNext() {
		ith := iter.Next()
		if index == n {
			return ith
		}

		index++
	}
	// Issue!
	panic("iterator out-of-bounds")
}

// Count provides a default implementation of Iterator.Count which can be used by
// other iterator implementations.
//
//nolint:revive
func Count[T any, S Enumerator[T]](iter S) uint {
	count := uint(0)

	for iter.HasNext() {
		iter.Next()
		//
		count++
	}
	// Issue!
	return count
}

// Collect provides a default implementation of Iterator.Collect which can be used by
// other iterator implementations.
//
//nolint:revive
func Collect[T any, S Enumerator[T]](iter S) []T {
	var items []T = make([]T, 0)
	//
	for iter.HasNext() {
		items = append(items, iter.Next())
	}
	//
	return items
}
