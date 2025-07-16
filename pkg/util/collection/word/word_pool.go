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
package word

import (
	"math"
	"sync"
)

// HEAP_POOL_INIT_BUCKETS determines the initial number of buckets to use for
// any instance.  Since we are geared towards large pools, we set this figure
// quite high.
const HEAP_POOL_INIT_BUCKETS = 1024

// HEAP_POOL_LOADING determines the loading point, overwhich rehashing will
// occur.  This is currently set to 75% capacity forces a rehashing.
const HEAP_POOL_LOADING = 75

// Pool provides an abstraction for referring to large words by a smaller index
// value.  The pool stores the actual word data, and provides fast access via an
// index.  This makes sense when we have a relatively small number of values
// which can be referred to many times over.
type Pool[K any, T Word[T]] interface {
	// Lookup a given word in the pool using an index.
	Get(K) T
	// Allocate word into pool, returning its index.
	Put(T) K
}

// HeapPool maintains a heap of bytes representing the words.
type HeapPool[T Word[T]] struct {
	// heap of bytes
	heap []byte
	// byte lengths for each chunk in the pool
	lengths []uint8
	// hash buckets
	buckets [][]uint
	// count of words stored
	count uint
	// mutex required to ensure thread safety.
	mux sync.RWMutex
}

// NewHeapPool constructs a new heap pool with an initial number of buckets.
func NewHeapPool[T Word[T]]() *HeapPool[T] {
	var (
		// zero-sized word
		empty T
		// Initial bucket allocation
		buckets = make([][]uint, HEAP_POOL_INIT_BUCKETS)
		pool    = &HeapPool[T]{heap: nil, lengths: []uint8{0}, buckets: buckets}
	)
	// Allocate zero-sized word as the first index.  This is tricky because we
	// want to ensure the address of this object is 0.
	pool.Put(empty)
	pool.heap = []byte{0}
	// Done
	return pool
}

// Get implementation for the Pool interface.
func (p *HeapPool[T]) Get(index uint) T {
	// Obtain read lock
	p.mux.RLock()
	// Determine length of word in heap
	word := p.innerGet(index)
	// Release read lock
	p.mux.RUnlock()
	// Initialise word
	return word
}

// Put implementation for the Pool interface.  This is somewhat challenging
// because it must be thread safe.  Since we anticipate a large number of cache
// hits compared with cache misses, we employ a Read/Write lock.
func (p *HeapPool[T]) Put(word T) uint {
	p.mux.RLock()
	// Check whether word already stored
	index, _ := p.has(word)
	// Release read lock
	p.mux.RUnlock()
	// Check whether we found it
	if index != math.MaxUint {
		// Yes, therefore return it.
		return index
	}
	// No, therefore begin critical section
	p.mux.Lock()
	// Recheck whether word stored in between read lock being released
	// (unlikely, but it is possible).
	index, hash := p.has(word)
	//
	if index == math.MaxUint {
		// Word still not present, so add it.
		index = p.alloc(word)
		// Record entry in relevant bucket
		p.buckets[hash] = append(p.buckets[hash], index)
		// Rehash (if necessary)
		p.rehashIfOverloaded()
	}
	// end critical section
	p.mux.Unlock()
	//
	return index
}

// Allocate a new word into the heap, returning its address.  This method is not
// threadsafe.
func (p *HeapPool[T]) alloc(word T) uint {
	var (
		address = uint(len(p.heap))
		// Determine length of word whilst ensuring that a completely empty word
		// occupies at least one byte (as, otherwise, we'd get some kind of
		// sharing going on).
		bytewidth = ByteWidth(word.BitWidth())
	)
	// Allocate space for new word
	for range bytewidth {
		p.heap = append(p.heap, 0)
		p.lengths = append(p.lengths, 0)
	}
	// Write word data
	word.Put(p.heap[address : address+bytewidth])
	// Configure word length
	p.lengths[address] = uint8(bytewidth)
	// Record word
	p.count++
	// Done
	return address
}

// Check whether the hash map is exceed its loading factor and, if so, rehash.
func (p *HeapPool[T]) rehashIfOverloaded() {
	load := (100 * p.count) / uint(len(p.buckets))
	//
	if load > HEAP_POOL_LOADING {
		// Force a rehash
		p.rehash()
	}
}

// Has checks whether a given word is stored in this heap, or not.
func (p *HeapPool[T]) has(word T) (uint, uint64) {
	hash := word.Hash() % uint64(len(p.buckets))
	// Attempt to lookup word
	for _, index := range p.buckets[hash] {
		if p.innerGet(index).Equals(word) {
			return index, hash
		}
	}
	//
	return math.MaxUint, hash
}

// unsynchronized version of Get to be used when a lock is already acquired.
func (p *HeapPool[T]) innerGet(index uint) T {
	var (
		word T
		// Determine length of word in heap
		length = uint(p.lengths[index])
		// Identify bytes of word in the heap
		bytes = p.heap[index : index+length]
	)
	// Initialise word
	return word.Set(bytes)
}

func (p *HeapPool[T]) rehash() {
	var (
		oldBuckets = p.buckets
		n          = uint64(len(oldBuckets) * 3)
	)
	// Double number of buckets
	p.buckets = make([][]uint, n)
	// Rehash!
	for _, bucket := range oldBuckets {
		for _, index := range bucket {
			// Determine new hash
			hash := p.innerGet(index).Hash() % n
			// Record index in relevant bucket
			p.buckets[hash] = append(p.buckets[hash], index)
		}
	}
}
