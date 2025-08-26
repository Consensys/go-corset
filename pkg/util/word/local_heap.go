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
	"slices"
)

// LocalHeap maintains a heap of bytes representing the words which is *not* thread
// safe.
type LocalHeap[T DynamicWord[T]] struct {
	// heap of bytes
	heap []byte
	// byte lengths for each chunk in the pool
	lengths []uint8
	// hash buckets
	buckets [][]uint32
	// count of words stored
	count uint
}

// NewLocalHeap constructs a new thread-unsafe heap
func NewLocalHeap[T DynamicWord[T]]() *LocalHeap[T] {
	// zero-sized word
	var (
		empty T
		p     = &LocalHeap[T]{
			lengths: []uint8{0},
			buckets: make([][]uint32, HEAP_POOL_INIT_BUCKETS),
			heap:    nil,
			count:   0,
		}
	)
	// Ensure address of this object is 0.
	p.Put(empty)
	p.heap = []byte{0}
	//
	return p
}

// Clone implementation for SharedPool interface.
func (p *LocalHeap[T]) Clone() LocalHeap[T] {
	var (
		heap    = make([]byte, len(p.heap))
		lengths = make([]uint8, len(p.lengths))
		buckets = make([][]uint32, len(p.buckets))
	)
	//
	copy(heap, p.heap)
	copy(lengths, p.lengths)
	//
	for i := range len(p.buckets) {
		buckets[i] = slices.Clone(p.buckets[i])
	}
	//
	return LocalHeap[T]{
		heap:    heap,
		lengths: lengths,
		buckets: buckets,
		count:   p.count,
	}
}

// Get implementation for the Pool interface.
func (p *LocalHeap[T]) Get(index uint32) T {
	// Determine length of word in heap
	return p.innerGet(index)
}

// Put implementation for the Pool interface.  This is somewhat challenging
// because it must be thread safe.  Since we anticipate a large number of cache
// hits compared with cache misses, we employ a Read/Write lock.
func (p *LocalHeap[T]) Put(word T) uint32 {
	// Check whether word stored in heap already
	index, hash := p.has(word)
	//
	if index == math.MaxUint32 {
		// Word not present, so add it.
		index = p.alloc(word)
		// Record entry in relevant bucket
		p.buckets[hash] = append(p.buckets[hash], index)
		// Rehash (if necessary)
		p.rehashIfOverloaded()
	}
	//
	return index
}

// MarshalBinary converts this heap into a sequence of bytes.
func (p *LocalHeap[T]) MarshalBinary() ([]byte, error) {
	panic("todo")
}

// UnmarshalBinary initialises this heap from a given set of data bytes. This
// should match exactly the encoding above.
func (p *LocalHeap[T]) UnmarshalBinary(data []byte) error {
	panic("todo")
}

// Allocate a new word into the heap, returning its address.  This method is not
// threadsafe.
func (p *LocalHeap[T]) alloc(word T) uint32 {
	var (
		address = uint32(len(p.heap))
		// Determine length of word
		bytewidth = uint32(word.ByteWidth())
	)
	// Allocate space for new word
	for range bytewidth {
		p.heap = append(p.heap, 0)
		p.lengths = append(p.lengths, 0)
	}
	// Write word data
	word.PutBytes(p.heap[address : address+bytewidth])
	// Configure word length
	p.lengths[address] = uint8(bytewidth)
	// Record word
	p.count++
	// Done
	return address
}

// Check whether the hash map is exceed its loading factor and, if so, rehash.
func (p *LocalHeap[T]) rehashIfOverloaded() {
	load := (100 * p.count) / uint(len(p.buckets))
	//
	if load > HEAP_POOL_LOADING {
		// Force a rehash
		p.rehash()
	}
}

// Has checks whether a given word is stored in this heap, or not.
func (p *LocalHeap[T]) has(word T) (uint32, uint64) {
	hash := word.Hash() % uint64(len(p.buckets))
	// Attempt to lookup word
	for _, index := range p.buckets[hash] {
		if p.innerGet(index).Equals(word) {
			return index, hash
		}
	}
	//
	return math.MaxUint32, hash
}

// unsynchronized version of Get to be used when a lock is already acquired.
func (p *LocalHeap[T]) innerGet(index uint32) T {
	var (
		word T
		// Determine length of word in heap
		length = uint32(p.lengths[index])
		// Identify bytes of word in the heap
		bytes = p.heap[index : index+length]
	)
	// Initialise word
	return word.SetBytes(bytes)
}

func (p *LocalHeap[T]) rehash() {
	var (
		oldBuckets = p.buckets
		n          = uint64(len(oldBuckets) * 3)
	)
	// Double number of buckets
	p.buckets = make([][]uint32, n)
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
