package util

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

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

// Equals returns true if both arrays contain equivalent elements.
func Equals(lhs []*fr.Element, rhs []*fr.Element) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i := 0; i < len(lhs); i++ {
		// Check lengths match
		if lhs[i].Cmp(rhs[i]) != 0 {
			return false
		}
	}
	//
	return true
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
