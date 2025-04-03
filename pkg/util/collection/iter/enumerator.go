// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package iter

import (
	utilmath "github.com/consensys/go-corset/pkg/util/math"
)

// Enumerator abstracts the process of iterating over a sequence of elements.
type Enumerator[T any] interface {
	// Check whether or not there are any items remaining to visit.
	HasNext() bool

	// Get the next item, and advanced the iterator.
	Next() T

	// Count the number of items left.  Note, this does not modify the iterator.
	Count() uint
}

// EnumerateElements returns an iterator which enumerates all arrays of size n
// over the given set of elements.  For example, if n==2 and elems contained two
// elements A and B, then this will return [[A,A],[A,B],[B,A],[B,B]].
func EnumerateElements[E any](n uint, elems []E) Enumerator[[]E] {
	counters := make([]uint, n)
	remaining := utilmath.PowUint64(uint64(len(elems)), uint64(n))
	//
	return &enumerator[E]{remaining, counters, elems}
}

type enumerator[E any] struct {
	remaining uint64
	counters  []uint
	elements  []E
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *enumerator[E]) HasNext() bool {
	return p.counters != nil
}

// Count returns the number of items left in this enumeration.
//
//nolint:revive
func (p *enumerator[E]) Count() uint {
	return uint(p.remaining)
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
	p.remaining -= 1
	//
	return rs
}
