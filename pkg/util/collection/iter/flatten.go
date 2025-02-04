package iter

// FlattenIterator provides an iterator implementation for an Array.
type flattenIterator[S, T comparable] struct {
	// Outermost iterator
	iter Iterator[S]
	// Innermost iterator
	curr Iterator[T]
	// Mapping function
	fn func(S) Iterator[T]
}

// NewFlattenIterator adapts a sequence of items S which themselves can be
// iterated as items T, into a flat sequence of items T.
func NewFlattenIterator[S, T comparable](iter Iterator[S], fn func(S) Iterator[T]) Iterator[T] {
	return &flattenIterator[S, T]{iter, nil, fn}
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *flattenIterator[S, T]) HasNext() bool {
	if p.curr != nil && p.curr.HasNext() {
		return true
	}
	// Find next hit
	for p.iter.HasNext() {
		p.curr = p.fn(p.iter.Next())
		if p.curr.HasNext() {
			return true
		}
	}
	// Failed
	return false
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *flattenIterator[S, T]) Next() T {
	// Can assume HasNext called, otherwise this is undefined anyway :)
	return p.curr.Next()
}

// Append another iterator onto the end of this iterator.  Thus, when all
// items are visited in this iterator, iteration continues into the other.
//
//nolint:revive
func (p *flattenIterator[S, T]) Append(iter Iterator[T]) Iterator[T] {
	return NewAppendIterator[T](p, iter)
}

// Clone creates a copy of this iterator at the given cursor position.
// Modifying the clone (i.e. by calling Next) iterator will not modify the
// original.
//
//nolint:revive
func (p *flattenIterator[S, T]) Clone() Iterator[T] {
	var curr Iterator[T]
	if p.curr != nil {
		curr = p.curr.Clone()
	}

	return &flattenIterator[S, T]{p.iter.Clone(), curr, p.fn}
}

// Collect allocates a new array containing all items of this iterator.
//
//nolint:revive
func (p *flattenIterator[S, T]) Collect() []T {
	items := make([]T, 0)
	if p.curr != nil {
		items = p.curr.Collect()
	}
	// Flatten each group in turn
	for p.iter.HasNext() {
		ith_items := p.fn(p.iter.Next()).Collect()
		items = append(items, ith_items...)
	}
	// Done
	return items
}

// Count returns the number of items left in the iterator
//
//nolint:revive
func (p *flattenIterator[S, T]) Count() uint {
	count := uint(0)

	for i := p.Clone(); i.HasNext(); {
		i.Next()

		count++
	}

	return count
}

// Find returns the index of the first match for a given predicate, or
// return false if no match is found.
//
//nolint:revive
func (p *flattenIterator[S, T]) Find(predicate Predicate[T]) (uint, bool) {
	return baseFind(p, predicate)
}

// Nth returns the nth item in this iterator
//
//nolint:revive
func (p *flattenIterator[S, T]) Nth(n uint) T {
	panic("todo")
}
