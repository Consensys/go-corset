package set

import (
	"cmp"
	"sort"

	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// SortedSet is an array of unique sorted values (i.e. no duplicates).
type SortedSet[T cmp.Ordered] []T

// NewSortedSet returns an empty sorted set.
func NewSortedSet[T cmp.Ordered]() *SortedSet[T] {
	return &SortedSet[T]{}
}

// Contains returns true if a given element is in the set.
//
//nolint:revive
func (p *SortedSet[T]) Contains(element T) bool {
	data := *p
	// Find index where element either does occur, or should occur.
	i := sort.Search(len(*p), func(i int) bool {
		return element <= data[i]
	})
	// Check whether item existed or not.
	return i < len(data) && data[i] == element
}

// Insert an element into this sorted set.
//
//nolint:revive
func (p *SortedSet[T]) Insert(element T) {
	data := *p
	// Find index where element either does occur, or should occur.
	i := sort.Search(len(*p), func(i int) bool {
		return element <= data[i]
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
func (p *SortedSet[T]) InsertSorted(q *SortedSet[T]) {
	left := *p
	right := *q
	// Check containment
	n := countDuplicates(left, right)
	// Check for total inclusion
	if n == len(right) {
		// Right set completedly included in left, so actually there is nothing
		// to do.
		return
	}
	// Allocate space
	ndata := make([]T, len(left)+len(right)-n)
	// Merge
	mergeSorted(ndata, left, right)
	// Finally copy over new data
	*p = ndata
}

// Iter returns an iterator over the elements of this sorted set.
//
//nolint:revive
func (p *SortedSet[T]) Iter() iter.Iterator[T] {
	return iter.NewArrayIterator(*p)
}

// UnionSortedSets unions together a number of things which can be turn into a
// sorted set using a given mapping function.  At some level, this is a
// map/reduce function.
func UnionSortedSets[S any, T cmp.Ordered](elems []S, fn func(S) *SortedSet[T]) *SortedSet[T] {
	if len(elems) == 0 {
		return NewSortedSet[T]()
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
func countDuplicates[T cmp.Ordered](left []T, right []T) int {
	// Check containment
	i := 0
	j := 0
	n := 0

	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
			i++
		} else if left[i] > right[j] {
			j++
		} else {
			i++
			j++
			n++ // duplicate detected
		}
	}

	return n
}

// Merge two sets of sorted arrays (left and right) into a target array.  This
// assumes the target array is big enough.
func mergeSorted[T cmp.Ordered](target []T, left []T, right []T) {
	i := 0
	j := 0
	k := 0
	// Merge overlap of both sets
	for ; i < len(left) && j < len(right); k++ {
		if left[i] < right[j] {
			target[k] = left[i]
			i++
		} else if left[i] > right[j] {
			target[k] = right[j]
			j++
		} else {
			target[k] = left[i]
			i++
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
