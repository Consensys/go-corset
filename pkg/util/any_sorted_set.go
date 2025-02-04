package util

import (
	"cmp"
	"sort"

	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// Comparable provides an interface which types used in a AnySortedSet must implement.
type Comparable[T any] interface {
	comparable
	LessEq(T) bool
}

// Order provides a wrapper around primtive types for use with an AnySortedSet.
// This is mostly for testing purposes.
type Order[T cmp.Ordered] struct {
	Item T
}

// LessEq wraps the respective primitive comparison.
func (lhs Order[T]) LessEq(rhs Order[T]) bool {
	return lhs.Item <= rhs.Item
}

// AnySortedSet is an array of unique sorted values (i.e. no duplicates).
type AnySortedSet[T Comparable[T]] []T

// NewAnySortedSet returns an empty sorted set.
func NewAnySortedSet[T Comparable[T]]() *AnySortedSet[T] {
	return &AnySortedSet[T]{}
}

// ToArray extracts the underlying array from this sorted set.
func (p *AnySortedSet[T]) ToArray() []T {
	return *p
}

// Contains returns true if a given element is in the set.
//
//nolint:revive
func (p *AnySortedSet[T]) Contains(element T) bool {
	data := *p
	// Find index where element either does occur, or should occur.
	i := sort.Search(len(data), func(i int) bool {
		// element <= data[i]
		return element.LessEq(data[i])
	})
	// Check whether item existed or not.
	return i < len(data) && data[i] == element
}

// Insert an element into this sorted set.
//
//nolint:revive
func (p *AnySortedSet[T]) Insert(element T) {
	data := *p
	// Find index where element either does occur, or should occur.
	i := sort.Search(len(data), func(i int) bool {
		// element <= data[i]
		return element.LessEq(data[i])
	})
	// Check whether item existed or not.
	if i >= len(data) || data[i] != element {
		// No, item was not found
		ndata := make([]T, len(data)+1)
		copy(ndata, data[0:i])
		ndata[i] = element
		copy(ndata[i+1:], data[i:])
		*p = ndata
	}
}

// InsertSorted inserts all elements in a given sorted set into this set.
//
//nolint:revive
func (p *AnySortedSet[T]) InsertSorted(q *AnySortedSet[T]) {
	left := *p
	right := *q
	// Check containment
	n := anyCountDuplicates(left, right)
	// Check for total inclusion
	if n == len(right) {
		// Right set completedly included in left, so actually there is nothing
		// to do.
		return
	}
	// Allocate space
	ndata := make([]T, len(left)+len(right)-n)
	// Merge
	anyMergeSorted(ndata, left, right)
	// Finally copy over new data
	*p = ndata
}

// Iter returns an iterator over the elements of this sorted set.
//
//nolint:revive
func (p *AnySortedSet[T]) Iter() iter.Iterator[T] {
	return iter.NewArrayIterator(*p)
}

// UnionAnySortedSets unions together a number of things which can be turn into a
// sorted set using a given mapping function.  At some level, this is a
// map/reduce function.
func UnionAnySortedSets[S any, T Comparable[T]](elems []S, fn func(S) *AnySortedSet[T]) *AnySortedSet[T] {
	if len(elems) == 0 {
		return NewAnySortedSet[T]()
	}
	// Map first element
	set := fn(elems[0])
	// Map/reduce the rest
	for i := 1; i < len(elems); i++ {
		// Map ith element
		ith := fn(elems[i])
		// Reduce
		set.InsertSorted(ith)
	}
	//
	return set
}

// Determine number of duplicate elements
func anyCountDuplicates[T Comparable[T]](left []T, right []T) int {
	// Check containment
	i := 0
	j := 0
	n := 0

	for i < len(left) && j < len(right) {
		if left[i] == right[j] {
			i++
			j++
			n++ // duplicate detected
		} else if left[i].LessEq(right[j]) {
			i++
		} else {
			j++
		}
	}

	return n
}

// Merge two sets of sorted arrays (left and right) into a target array.  This
// assumes the target array is big enough.
func anyMergeSorted[T Comparable[T]](target []T, left []T, right []T) {
	i := 0
	j := 0
	k := 0
	// Merge overlap of both sets
	for ; i < len(left) && j < len(right); k++ {
		if left[i] == right[j] {
			target[k] = left[i]
			i++
			j++
		} else if left[i].LessEq(right[j]) {
			target[k] = left[i]
			i++
		} else {
			target[k] = right[j]
			j++
		}
	}
	// Handle anything left
	if i < len(left) {
		copy(target[k:], left[i:])
	} else if j < len(right) {
		copy(target[k:], right[j:])
	}
}
