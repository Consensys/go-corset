package util

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
// Append Iterator
// ===============================================================

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

// ===============================================================
// Array Iterator
// ===============================================================

// ArrayIterator provides an iterator implementation for an Array.
type arrayIterator[T any] struct {
	items []T
	index uint
}

// NewArrayIterator construct an iterator over an array of items.
func NewArrayIterator[T any](items []T) Iterator[T] {
	return &arrayIterator[T]{items, 0}
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *arrayIterator[T]) HasNext() bool {
	return p.index < uint(len(p.items))
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *arrayIterator[T]) Next() T {
	next := p.items[p.index]
	p.index++

	return next
}

// Append another iterator onto the end of this iterator.  Thus, when all
// items are visited in this iterator, iteration continues into the other.
//
//nolint:revive
func (p *arrayIterator[T]) Append(iter Iterator[T]) Iterator[T] {
	return NewAppendIterator(p, iter)
}

// Clone creates a copy of this iterator at the given cursor position.
// Modifying the clone (i.e. by calling Next) iterator will not modify the
// original.
//
//nolint:revive
func (p *arrayIterator[T]) Clone() Iterator[T] {
	return &arrayIterator[T]{p.items, p.index}
}

// Collect allocates a new array containing all items of this iterator.
// This drains the iterator.
//
//nolint:revive
func (p *arrayIterator[T]) Collect() []T {
	items := make([]T, len(p.items))
	copy(items, p.items)

	return items
}

// Count returns the number of items left in the iterator
//
//nolint:revive
func (p *arrayIterator[T]) Count() uint {
	return uint(len(p.items)) - p.index
}

// Find returns the index of the first match for a given predicate, or
// return false if no match is found.
//
//nolint:revive
func (p *arrayIterator[T]) Find(predicate Predicate[T]) (uint, bool) {
	return baseFind(p, predicate)
}

// Nth returns the nth item in this iterator
//
//nolint:revive
func (p *arrayIterator[T]) Nth(n uint) T {
	p.index = n
	return p.items[n]
}

// ===============================================================
// Cast Iterator
// ===============================================================
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

// ===============================================================
// Flatten Iterator
// ===============================================================

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

// ===============================================================
// Unit Iterator
// ===============================================================

type unitIterator[T any] struct {
	item  T
	index uint
}

// NewUnitIterator construct an iterator over an array of items.
func NewUnitIterator[T any](item T) *unitIterator[T] {
	return &unitIterator[T]{item, 0}
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *unitIterator[T]) HasNext() bool {
	return p.index < 1
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *unitIterator[T]) Next() T {
	p.index++
	return p.item
}

// Append another iterator onto the end of this iterator.  Thus, when all
// items are visited in this iterator, iteration continues into the other.
//
//nolint:revive
func (p *unitIterator[T]) Append(iter Iterator[T]) Iterator[T] {
	return NewAppendIterator(p, iter)
}

// Clone creates a copy of this iterator at the given cursor position. Modifying
// the clone (i.e. by calling Next) iterator will not modify the original.
//
//nolint:revive
func (p *unitIterator[T]) Clone() Iterator[T] {
	return &unitIterator[T]{p.item, p.index}
}

// Collect allocates a new array containing all items of this iterator.
// This drains the iterator.
//
//nolint:revive
func (p *unitIterator[T]) Collect() []T {
	items := make([]T, 1)
	items[0] = p.item

	return items
}

// Count returns the number of items left in the iterator
//
//nolint:revive
func (p *unitIterator[T]) Count() uint {
	if p.index == 0 {
		return 1
	}
	// nothing left
	return 0
}

// Find returns the index of the first match for a given predicate, or
// return false if no match is found.
//
//nolint:revive
func (p *unitIterator[T]) Find(predicate Predicate[T]) (uint, bool) {
	if predicate(p.item) {
		// Success
		return 0, true
	}
	// Failed
	return 0, false
}

// Nth returns the nth item in this iterator
//
//nolint:revive
func (p *unitIterator[T]) Nth(n uint) T {
	return p.item
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
