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

import "github.com/consensys/go-corset/pkg/util/collection/array"

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
func (p *castIterator[S, T]) Find(predicate array.Predicate[T]) (uint, bool) {
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
