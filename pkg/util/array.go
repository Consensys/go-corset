package util

// Array is an abstract representation of an array of zero or more elements.
// Typically this is backed by a single Go array, but this does not have to be
// the case.
type Array[T any] interface {
	// Add a given element to this array, returning the index of the new item.
	Add(T) uint
	// Clone the given array
	Clone() Array[T]
	// Find the first element matching a given predicate, returning its index.
	// Otherwise, returns false when no such element exists.
	Find(Predicate[T]) (uint, bool)
	// Get the ith element of this array.
	Get(uint) T
	// Has determines whether an element exists for which the given predicate holds.
	Has(Predicate[T]) bool
	// Return an iterator over the items in this array.
	Iter() Iterator[T]
	// Returns the number of items in this array.
	Len() uint
	// Swap two elements in this array.
	Swap(uint, uint)
}

// Array_1 is an implementation of Array which is backed by a single underlying
// array.
type Array_1[T any] struct {
	items []T
}

// NewArray_1 constructs a new array from an underlying Go array.
func NewArray_1[T any](items []T) Array_1[T] {
	return Array_1[T]{items}
}

// Add a given element to this array, returning the index of the new item.
func (p *Array_1[T]) Add(item T) uint {
	index := p.Len()
	p.items = append(p.items, item)

	return index
}

// Clone the given array
func (p *Array_1[T]) Clone() Array[T] {
	arr := p.Copy()
	return &arr
}

// Copy creates a copy of this array which, in particular, clones the underlying
// array itself.  Thus, modifications to the copy should not affect the original.
func (p *Array_1[T]) Copy() Array_1[T] {
	nitems := make([]T, len(p.items))
	copy(nitems, p.items)

	return Array_1[T]{nitems}
}

// Find the first element matching a given predicate, returning its index.
// Otherwise, returns false when no such element exists.
//
//nolint:revive
func (p *Array_1[T]) Find(predicate Predicate[T]) (uint, bool) {
	return p.Iter().Find(predicate)
}

// Get the ith element of this array.
func (p *Array_1[T]) Get(index uint) T {
	return p.items[index]
}

// Has determines whether an element exists for which the given predicate holds.
func (p *Array_1[T]) Has(predicate Predicate[T]) bool {
	_, r := p.Find(predicate)
	return r
}

// Iter returns an iterator over the items in this array.
func (p *Array_1[T]) Iter() Iterator[T] {
	return NewArrayIterator(p.items)
}

// Len returns the number of items in this array.
func (p *Array_1[T]) Len() uint {
	return uint(len(p.items))
}

// Swap two elements in this array.
func (p *Array_1[T]) Swap(l uint, r uint) {
	lth := p.items[l]
	p.items[l] = p.items[r]
	p.items[r] = lth
}
