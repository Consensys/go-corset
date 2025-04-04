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
package enum

import "math/rand/v2"

// Sample constructs a sampling enumerator from a given enumerator.
// Specifically, the sampling enumerator samples exactly n items from the given
// enumerator (assuming it held at least n items).  Otherwise, if n is larger or
// equal to the number of elements in the original numerator it simply returns
// that.
func Sample[T any](n uint, enum Enumerator[T]) Enumerator[T] {
	if n >= enum.Count() {
		return enum
	}
	//
	return &samplingEnumerator[T]{n, enum}
}

type samplingEnumerator[T any] struct {
	// Number of items left to sample
	count uint
	//
	enum Enumerator[T]
}

// HasNext checks whether or not there are any items remaining to visit.
//
//nolint:revive
func (p *samplingEnumerator[E]) HasNext() bool {
	return p.count > 0
}

// Count returns the number of items left in this enumeration.
//
//nolint:revive
func (p *samplingEnumerator[E]) Count() uint {
	return p.count
}

// Nth returns the nth item in this iterator.  This will mutate the iterator.
func (p *samplingEnumerator[E]) Nth(n uint) E {
	return Nth(p, n)
}

// Next returns the next item, and advance the iterator.
//
//nolint:revive
func (p *samplingEnumerator[E]) Next() E {
	// Determine gap to next item
	gap := rand.UintN(p.enum.Count() - p.count)
	// Decrement for item returned
	p.count--
	//
	return p.enum.Nth(gap)
}
