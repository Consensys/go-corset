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
package util

import (
	"cmp"
	"fmt"
	"io"
	"math"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Predicate abstracts the notion of a function which identifies something.
type Predicate[T any] func(T) bool

// Array provides a generice interface to an array of elements.  Typically, we
// are interested in arrays of field elements here.
type Array[T comparable] interface {
	// Returns the number of elements in this array.
	Len() uint
	// Get returns the element at the given index in this array.
	Get(uint) T
	// Set the element at the given index in this array, overwriting the
	// original value.
	Set(uint, T)
	// Clone makes clones of this array producing an otherwise identical copy.
	Clone() Array[T]
	// Slice out a subregion of this array.
	Slice(uint, uint) Array[T]
	// Return the number of bits required to store an element of this array.
	BitWidth() uint
	// Insert n copies of T at start of the array and m copies at the back
	// producing an updated array.
	Pad(uint, uint, T) Array[T]
	// Write out the contents of this array, assuming a minimal unit of 1 byte
	// per element.
	Write(w io.Writer) error
}

// Prepend creates a new slice containing the result of prepending the given
// item onto the end of the given slice.  Observe that, unlike the built-in
// append() function, this will never modify the given slice.
func Prepend[T any](item T, slice []T) []T {
	n := len(slice)
	// Make space for new slice
	nslice := make([]T, n+1)
	// Copy existing values
	copy(nslice[1:], slice)
	// Set first value
	nslice[0] = item
	// Done
	return nslice
}

// Append creates a new slice containing the result of appending the given item
// onto the end of the given slice.  Observe that, unlike the built-in append()
// function, this will never modify the given slice.
//
//nolint:revive
func Append[T any](slice []T, item T) []T {
	n := len(slice)
	// Make space for new slice
	nslice := make([]T, n+1)
	// Copy existing values
	copy(nslice[:n], slice)
	// Set last value
	nslice[n] = item
	// Done
	return nslice
}

// Append creates a new slice containing the result of appending the given item
// onto the end of the given slice.  Observe that, unlike the built-in append()
// function, this will never modify the given slice.
//
//nolint:revive
func AppendAll[T any](lhs []T, rhs ...T) []T {
	n := len(lhs)
	m := len(rhs)
	// Make space for new slice
	nslice := make([]T, n+m)
	// Copy left values
	copy(nslice[:n], lhs)
	// Copy right values
	copy(nslice[n:], rhs)
	// Done
	return nslice
}

// CountUnique counts the number of unique items within a given slice.
func CountUnique[T cmp.Ordered](items []T) uint {
	// First sort them
	slices.Sort(items)
	//
	fmt.Printf("got %v\n", items)
	//
	count := uint(0)
	//
	for i, v := range items {
		if i == 0 || items[i-1] != v {
			count++
		}
	}
	//
	return count
}

// ReplaceFirstOrPanic replaces the first occurrence of a given item (from) in an
// array with another item (to).  If not match is found, then this will panic.
// In otherwords, we are expecting a match.
func ReplaceFirstOrPanic[T comparable](columns []T, from T, to T) {
	for i, c := range columns {
		if c == from {
			// Success
			columns[i] = to
			return
		}
	}
	// Failure
	panic(fmt.Sprintf("invalid replace (item %s not found)", any(from)))
}

// FindMatching determines the index of first matching item in a given array, or
// returns max.MaxUint otherwise.
func FindMatching[T any](items []T, predicate Predicate[T]) uint {
	for i, item := range items {
		if predicate(item) {
			return uint(i)
		}
	}
	//
	return math.MaxUint
}

// ContainsMatching checks whether a given array contains an item matching a given predicate.
func ContainsMatching[T any](items []T, predicate Predicate[T]) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	//
	return false
}

// InsertAt constructs an identical slice, except with the element inserted at
// the given index.  If the index is beyond the bounds of the array, then the
// element is simply appended.
func InsertAt[T any](items []T, element T, index uint) []T {
	n := uint(len(items))
	//
	if index < n {
		first := items[:index]
		second := items[index:]
		items = make([]T, n+1)
		copy(items, first)
		copy(items[index+1:], second)
		items[index] = element
	} else {
		items = append(items, element)
	}
	//
	return items
}

// RemoveAt constructs an identical slice, except with the element at the given
// index removed.  If the index is beyond the bounds of the array, then there is
// no change.
func RemoveAt[T any](items []T, index uint) []T {
	n := uint(len(items))
	//
	if index < n {
		first := items[0:index]
		second := items[index+1:]
		items = append(first, second...)
	}
	//
	return items
}

// RemoveMatching removes all elements from an array matching the given item.
func RemoveMatching[T any](items []T, predicate Predicate[T]) []T {
	count := 0
	// Check how many matches we have
	for _, r := range items {
		if !predicate(r) {
			count++
		}
	}
	// Check for stuff to remove
	if count != len(items) {
		nitems := make([]T, count)
		j := 0
		// Remove items
		for i, r := range items {
			if !predicate(r) {
				nitems[j] = items[i]
				j++
			}
		}
		//
		items = nitems
	}
	//
	return items
}

// RemoveMatchingIndexed removes all elements from an array matching the given item.
func RemoveMatchingIndexed[T any](items []T, predicate func(int, T) bool) []T {
	count := 0
	// Check how many matches we have
	for i, r := range items {
		if !predicate(i, r) {
			count++
		}
	}
	// Check for stuff to remove
	if count != len(items) {
		nitems := make([]T, count)
		j := 0
		// Remove items
		for i, r := range items {
			if !predicate(i, r) {
				nitems[j] = items[i]
				j++
			}
		}
		//
		items = nitems
	}
	//
	return items
}

// Flatten flattens items from an array which expand into arrays of terms.
func Flatten[T any](items []T, fn func(T) []T) []T {
	for _, t := range items {
		if fn(t) != nil {
			return forceFlatten(items, fn)
		}
	}
	// no change
	return items
}

func forceFlatten[T any](items []T, fn func(T) []T) []T {
	nitems := make([]T, 0)
	//
	for _, t := range items {
		if ts := fn(t); ts != nil {
			nitems = append(nitems, ts...)
		} else {
			nitems = append(nitems, t)
		}
	}
	// no change
	return nitems
}

// Equals2d returns true if two 2D arrays are equal.
func Equals2d(lhs [][]fr.Element, rhs [][]fr.Element) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i := 0; i < len(lhs); i++ {
		lhs_i := lhs[i]
		rhs_i := rhs[i]
		// Check lengths match
		if len(lhs_i) != len(rhs_i) {
			return false
		}
		// Check elements match
		for j := 0; j < len(lhs_i); j++ {
			if lhs_i[j].Cmp(&rhs_i[j]) != 0 {
				return false
			}
		}
	}
	//
	return true
}

// FlatArrayIndexOf_2 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_2[A any, B any](index int, as []A, bs []B) any {
	if index < len(as) {
		return as[index]
	}
	// Otherwise bs
	return bs[index-len(as)]
}

// FlatArrayIndexOf_3 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_3[A any, B any, C any](index int, as []A, bs []B, cs []C) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_2(index-len(as), bs, cs)
}

// FlatArrayIndexOf_4 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_4[A any, B any, C any, D any](index int, as []A, bs []B, cs []C, ds []D) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_3(index-len(as), bs, cs, ds)
}

// FlatArrayIndexOf_5 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_5[A any, B any, C any, D any, E any](index int, as []A, bs []B, cs []C, ds []D, es []E) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_4(index-len(as), bs, cs, ds, es)
}

// FlatArrayIndexOf_6 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_6[A any, B any, C any, D any, E any, F any](index int, as []A, bs []B, cs []C, ds []D, es []E,
	fs []F) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_5(index-len(as), bs, cs, ds, es, fs)
}
