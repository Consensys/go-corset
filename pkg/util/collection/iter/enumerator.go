package iter

// Enumerator abstracts the process of iterating over a sequence of elements.
type Enumerator[T any] interface {
	// Check whether or not there are any items remaining to visit.
	HasNext() bool

	// Get the next item, and advanced the iterator.
	Next() T
}

// EnumerateElements returns an iterator which enumerates all arrays of size n
// over the given set of elements.  For example, if n==2 and elems contained two
// elements A and B, then this will return [[A,A],[A,B][B,A],[B,B]].
func EnumerateElements[E any](n uint, elems []E) Enumerator[[]E] {
	counters := make([]uint, n)
	return &enumerator[E]{counters, elems}
}

type enumerator[E any] struct {
	counters []uint
	elements []E
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *enumerator[E]) HasNext() bool {
	return p.counters != nil
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *enumerator[E]) Next() []E {
	rs := make([]E, len(p.counters))
	// Copy over elements
	for i := 0; i < len(rs); i++ {
		rs[i] = p.elements[p.counters[i]]
	}
	//
	carry := false
	// Increment counters
	for i := 0; i < len(p.counters); i++ {
		ithp1 := p.counters[i] + 1
		// Check for oveflow
		if ithp1 != uint(len(p.elements)) {
			// No overflow
			p.counters[i] = ithp1
			carry = false
			// Done incrementing
			break
		}
		// overflow
		p.counters[i] = 0
		carry = true
	}
	// Check whether finished
	if carry {
		// Yes, signal end of enumeration
		p.counters = nil
	}
	//
	return rs
}
