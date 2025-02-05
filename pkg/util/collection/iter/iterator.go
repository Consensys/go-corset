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

func baseFind[T any, S Enumerator[T]](iter S, predicate Predicate[T]) (uint, bool) {
	index := uint(0)

	for i := iter; i.HasNext(); {
		if predicate(i.Next()) {
			return index, true
		}

		index++
	}
	// Failed to find it
	return 0, false
}

func baseNth[T any, S Enumerator[T]](iter S, n uint) T {
	index := uint(0)

	for i := iter; i.HasNext(); {
		ith := i.Next()
		if index == n {
			return ith
		}

		index++
	}
	// Issue!
	panic("iterator out-of-bounds")
}
