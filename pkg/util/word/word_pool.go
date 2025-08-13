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
	"sync"
)

// StaticPool represents a pool which stores words "as is", and does not attempt
// to compress them into shorter byte sequences.
type StaticPool[T Word[T]] struct {
	// heap of bytes
	words []T
	// hash buckets
	buckets [][]uint
	// mutex required to ensure thread safety.
	mux sync.RWMutex
}

var _ Pool[uint, BigEndian] = &StaticPool[BigEndian]{}

// NewStaticPool constructs a new heap pool with an initial number of buckets.
func NewStaticPool[T Word[T]]() *StaticPool[T] {
	var (
		// Initial bucket allocation
		buckets = make([][]uint, HEAP_POOL_INIT_BUCKETS)
		pool    = &StaticPool[T]{words: nil, buckets: buckets}
	)
	// Done
	return pool
}

// Clone implementation for Pool interface
func (p *StaticPool[T]) Clone() Pool[uint, T] {
	panic("todo")
}

// Get implementation for the Pool interface.
func (p *StaticPool[T]) Get(index uint) T {
	panic("todo")
}

// IndexOf implementation for the Pool interface.
func (p *StaticPool[T]) IndexOf(word T) (uint, bool) {
	panic("todo")
}

// Put implementation for the Pool interface.  This is somewhat challenging
// because it must be thread safe.  Since we anticipate a large number of cache
// hits compared with cache misses, we employ a Read/Write lock.
func (p *StaticPool[T]) Put(word T) uint {
	panic("todo")
}
