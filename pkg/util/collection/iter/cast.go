package iter

type castIterator[S, T any] struct {
	iter Iterator[S]
}

// NewCastIterator construct an iterator over an array of items.
func NewCastIterator[S, T any](iter Iterator[S]) Iterator[T] {
	return &castIterator[S, T]{iter}
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *castIterator[S, T]) HasNext() bool {
	return p.iter.HasNext()
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *castIterator[S, T]) Next() T {
	n := any(p.iter.Next())
	return n.(T)
}

// Append another iterator onto the end of this iterator.  Thus, when all
// items are visited in this iterator, iteration continues into the other.
//
//nolint:revive
func (p *castIterator[S, T]) Append(iter Iterator[T]) Iterator[T] {
	return NewAppendIterator(p, iter)
}

// Clone creates a copy of this iterator at the given cursor position.
// Modifying the clone (i.e. by calling Next) iterator will not modify the
// original.
//
//nolint:revive
func (p *castIterator[S, T]) Clone() Iterator[T] {
	return NewCastIterator[S, T](p.iter.Clone())
}

// Collect allocates a new array containing all items of this iterator. This drains the iterator.
//
//nolint:revive
func (p *castIterator[S, T]) Collect() []T {
	items := make([]T, p.iter.Count())
	index := 0

	for i := p.iter; i.HasNext(); {
		n := any(i.Next())
		items[index] = n.(T)
		index++
	}

	return items
}

// Count returns the number of items left in the iterator
//
//nolint:revive
func (p *castIterator[S, T]) Count() uint {
	return p.iter.Count()
}

// Find returns the index of the first match for a given predicate, or
// return false if no match is found.
//
//nolint:revive
func (p *castIterator[S, T]) Find(predicate Predicate[T]) (uint, bool) {
	return p.iter.Find(func(item S) bool {
		tmp := any(item)
		return predicate(tmp.(T))
	})
}

// Nth returns the nth item in this iterator
//
//nolint:revive
func (p *castIterator[S, T]) Nth(n uint) T {
	v := any(p.iter.Nth(n))
	return v.(T)
}
