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
package bit

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// Set provides a straightforward bitset implementation.  That is, a set of
// (unsigned) integer values implemented as an array of bits.
type Set struct {
	words []uint64
}

// Insert a given value into this set.
func (p *Set) Insert(val uint) {
	word := val / 64
	bit := val % 64
	//
	for uint(len(p.words)) <= word {
		p.words = append(p.words, 0)
	}
	// Set bit
	mask := uint64(1) << bit
	p.words[word] = p.words[word] | mask
}

// InsertAll inserts zero or more elements into this bitset.
func (p *Set) InsertAll(vals ...uint) {
	for _, v := range vals {
		p.Insert(v)
	}
}

// Union inserts all elements from a given bitset into this bitset.
func (p *Set) Union(bits Set) {
	for len(p.words) < len(bits.words) {
		p.words = append(p.words, 0)
	}
	// Insert all
	for w := 0; w < len(bits.words); w++ {
		p.words[w] = p.words[w] | bits.words[w]
	}
}

// Contains checks whether a given value is contained, or not.
func (p *Set) Contains(val uint) bool {
	word := val / 64
	bit := val % 64
	//
	if uint(len(p.words)) <= word {
		return false
	}
	// Set mask
	mask := uint64(1) << bit
	//
	return (p.words[word] & mask) != 0
}

// Count returns the number of bits in the bitset which are set to one.
func (p *Set) Count() uint {
	count := uint(0)
	//
	for word := uint(0); word < uint(len(p.words)); word++ {
		bits := p.words[word]
		//
		for bits != 0 {
			if bits&1 == 1 {
				count++
			}
			//
			bits = bits >> 1
		}
	}
	//
	return count
}

// Iter returns an iterator over the elements of this bitset.
func (p *Set) Iter() iter.Iterator[uint] {
	return &iterator{p.words, 0}
}

func (p *Set) String() string {
	var (
		builder strings.Builder
		first   = true
	)
	//
	builder.WriteString("[")
	//
	for word := uint(0); word < uint(len(p.words)); word++ {
		for bit := uint(0); bit < 64; bit++ {
			value := (word * 64) + bit

			if p.Contains(value) {
				if !first {
					builder.WriteString(", ")
				}
				//
				first = false
				//
				builder.WriteString(fmt.Sprintf("%d", value))
			}
		}
	}
	//
	builder.WriteString("]")
	//
	return builder.String()
}

// ============================================================================
// Iterator
// ============================================================================
type iterator struct {
	words []uint64
	value uint
}

func (p *iterator) HasNext() bool {
	n := uint(len(p.words))
	word := p.value / 64
	bit := p.value % 64
	mask := uint64(1) << bit
	// skip empty words
	for word < n && (p.words[word] == 0 || p.words[word] < mask) {
		bit = 0
		mask = 1
		word = word + 1
	}
	//
	if word < n {
		for i := bit; i < 64; i++ {
			mask := uint64(1) << i
			if (p.words[word] & mask) != 0 {
				p.value = (word * 64) + i
				return true
			}
		}
	}
	//
	p.value = n * 64
	// Done
	return false
}

func (p *iterator) Next() uint {
	next := p.value
	p.value = p.value + 1
	//
	return next
}

// Append another iterator onto the end of this iterator.  Thus, when all
// items are visited in this iterator, iteration continues into the other.
//
//nolint:revive
func (p *iterator) Append(other iter.Iterator[uint]) iter.Iterator[uint] {
	return iter.NewAppendIterator[uint](p, other)
}

// Clone creates a copy of this iterator at the given cursor position.
// Modifying the clone (i.e. by calling Next) iterator will not modify the
// original.
//
//nolint:revive
func (p *iterator) Clone() iter.Iterator[uint] {
	return &iterator{p.words, p.value}
}

// Collect allocates a new array containing all items of this iterator.
// This drains the iterator.
//
//nolint:revive
func (p *iterator) Collect() []uint {
	return iter.Collect(p)
}

// Count returns the number of items left in the iterator
//
//nolint:revive
func (p *iterator) Count() uint {
	return iter.Count(p)
}

// Find returns the index of the first match for a given predicate, or
// return false if no match is found.
//
//nolint:revive
func (p *iterator) Find(predicate iter.Predicate[uint]) (uint, bool) {
	return iter.Find(p, predicate)
}

// Nth returns the nth item in this iterator
//
//nolint:revive
func (p *iterator) Nth(n uint) uint {
	return iter.Nth(p, n)
}
