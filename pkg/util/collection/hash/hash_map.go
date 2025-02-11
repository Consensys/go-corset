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
package hash

import (
	"fmt"
	"strings"
)

// Map defines a generic set implementation backed by a map.  This is a true
// hashtable in that collisions are handle gracefully using buckets, rather than
// simply discarding them.
type Map[K Hasher[K], V any] struct {
	// items maps hashcodes to *buckets* of items.
	buckets map[uint64]hashMapBucket[K, V]
}

// NewMap creates a new HashMap with a given underlying capacity.
func NewMap[K Hasher[K], V any](size uint) *Map[K, V] {
	items := make(map[uint64]hashMapBucket[K, V], size)
	return &Map[K, V]{items}
}

// Size returns the number of unique items stored in this HashMap.
//
//nolint:revive
func (p *Map[K, V]) Size() uint {
	count := uint(0)
	for _, b := range p.buckets {
		count += b.size()
	}

	return count
}

// MaxBucket returns the size of the largest bucket.
//
//nolint:revive
func (p *Map[K, V]) MaxBucket() uint {
	m := uint(0)
	for _, b := range p.buckets {
		m = max(m, b.size())
	}

	return m
}

// Insert a new item into this map, returning true if it was already contained
// and false otherwise.
//
//nolint:revive
func (p *Map[K, V]) Insert(key K, value V) bool {
	var b1 hashMapBucket[K, V]
	// Compute item's hashcode
	hash := key.Hash()
	// Lookup existing bucket
	b1 = p.buckets[hash]
	// Insert new item
	r := b1.insert(key, value)
	// Update map
	p.buckets[hash] = b1
	// Done
	return r
}

// Contains checks whether the given item is contained within this map, or not.
//
//nolint:revive
func (p *Map[K, V]) ContainsKey(key K) bool {
	hash := key.Hash()

	if bucket, ok := p.buckets[hash]; ok {
		return bucket.containsKey(key)
	}

	return false
}

// Get item from bucket, or return false otherwise.
//
//nolint:revive
func (p *Map[K, V]) Get(key K) (V, bool) {
	var (
		empty V
		hash  uint64 = key.Hash()
	)
	// Look for bucket
	if bucket, ok := p.buckets[hash]; ok {
		return bucket.get(key)
	}

	return empty, false
}

//nolint:revive
func (p *Map[K, V]) String() string {
	var r strings.Builder
	//
	first := true
	// Write opening brace
	r.WriteString("{")
	// Iterate all buckets
	for _, b := range p.buckets {
		// Iterate all items in bucket
		for i, k := range b.keys {
			if !first {
				r.WriteString(",")
			}

			first = false

			r.WriteString(fmt.Sprintf("%s:=%s", any(k), any(b.values[i])))
		}
	}
	// Write closing brace
	r.WriteString("}")
	// Done
	return r.String()
}

// ============================================================================
// Bucket
// ============================================================================

type hashMapBucket[K Hasher[K], V any] struct {
	keys   []K
	values []V
}

// Get the number of items in this bucket.
//
//nolint:revive
func (b *hashMapBucket[K, V]) size() uint {
	return uint(len(b.keys))
}

// Insert a new item into this bucket
//
//nolint:revive
func (b *hashMapBucket[K, V]) insert(key K, value V) bool {
	// Determine whether key already present
	for i, k := range b.keys {
		if key.Equals(k) {
			b.values[i] = value
			return true
		}
	}
	// Append item
	b.keys = append(b.keys, key)
	b.values = append(b.values, value)
	// Item not present
	return false
}

// Check whether this bucket contains a given item, or not.
//
//nolint:revive
func (b *hashMapBucket[K, V]) containsKey(key K) bool {
	for _, k := range b.keys {
		if key.Equals(k) {
			return true
		}
	}

	return false
}

// Get item from bucket, or return false otherwise.
//
//nolint:revive
func (b *hashMapBucket[K, V]) get(key K) (V, bool) {
	var empty V

	for i, k := range b.keys {
		if key.Equals(k) {
			return b.values[i], true
		}
	}

	return empty, false
}
