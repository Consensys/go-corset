package iter

type appendIterator[T any] struct {
	left  Iterator[T]
	right Iterator[T]
}

// NewAppendIterator construct an iterator over an array of items.
func NewAppendIterator[T any](left Iterator[T], right Iterator[T]) Iterator[T] {
	return &appendIterator[T]{left, right}
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *appendIterator[T]) HasNext() bool {
	return p.left.HasNext() || p.right.HasNext()
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *appendIterator[T]) Next() T {
	if p.left.HasNext() {
		return p.left.Next()
	}

	return p.right.Next()
}

// Append another iterator onto the end of this iterator.  Thus, when all
// items are visited in this iterator, iteration continues into the other.
//
//nolint:revive
func (p *appendIterator[T]) Append(iter Iterator[T]) Iterator[T] {
	return NewAppendIterator(p, iter)
}

// Clone creates a copy of this iterator at the given cursor position.
// Modifying the clone (i.e. by calling Next) iterator will not modify the
// original.
//
//nolint:revive
func (p *appendIterator[T]) Clone() Iterator[T] {
	return NewAppendIterator[T](p.left.Clone(), p.right.Clone())
}

// Collect allocates a new array containing all items of this iterator.
// This drains the iterator.
//
//nolint:revive
func (p *appendIterator[T]) Collect() []T {
	lhs := p.left.Collect()
	rhs := p.right.Collect()

	return append(lhs, rhs...)
}

// Count returns the number of items left in the iterator
//
//nolint:revive
func (p *appendIterator[T]) Count() uint {
	return p.left.Count() + p.right.Count()
}

// Find returns the index of the first match for a given predicate, or
// return false if no match is found.
//
//nolint:revive
func (p *appendIterator[T]) Find(predicate Predicate[T]) (uint, bool) {
	return baseFind(p, predicate)
}

// Nth returns the nth item in this iterator
//
//nolint:revive
func (p *appendIterator[T]) Nth(n uint) T {
	// TODO: improve performance.
	return baseNth(p, n)
}
