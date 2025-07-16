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
	// lengths for each chunk in the pool
	lengths []uint8
}

// NewHeapPool constructs a new heap oriented pool.
func NewHeapPool[T Word[T]]() *HeapPool[T] {
	panic("todo")
}

// Get implementation for the Pool interface.
func (p *HeapPool[T]) Get(index uint) T {
	panic("got here")
}

// Put implementation for the Pool interface.
func (p *HeapPool[T]) Put(word T) uint {
	panic("got here")
}
