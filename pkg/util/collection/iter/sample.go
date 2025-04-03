package iter

import "math/rand/v2"

// SampleElements samples exactly n elements from a given enumerator (unless
// the enumerator has fewer items, in which case it always returns the original
// enumerator).
func SampleElements[E any](n uint, e Enumerator[E]) Enumerator[E] {
	if e.Count() <= n {
		return e
	}
	//
	return &samplingEnumerator[E]{e, n}
}

type samplingEnumerator[E any] struct {
	enumerator Enumerator[E]
	// number of items left to sample
	left uint
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *samplingEnumerator[E]) HasNext() bool {
	return p.left > 0
}

// Count returns the number of items left in this enumeration.
//
//nolint:revive
func (p *samplingEnumerator[E]) Count() uint {
	return p.left
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *samplingEnumerator[E]) Next() E {
	// Decrease number of items
	p.left--
	// Number of items can choose from
	left := p.enumerator.Count() - p.left
	// Advance enumerator
	return Nth(p.enumerator, rand.UintN(left))
}
