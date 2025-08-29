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
package array

import (
	"fmt"
	"reflect"

	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/word"
)

// Builder is a mechanism for constructing arrays which aims to select the
// right representation for a given array.
type Builder[T any] interface {
	// NewArray constructs a new array of the given height holding elements of the given bitwidth
	NewArray(height uint, bitwidth uint) MutArray[T]
}

// ============================================================================
// Static Builder
// ============================================================================

// NewStaticBuilder constructs a new array builder for dynamic words.
func NewStaticBuilder[T word.Word[T]]() Builder[T] {
	var builder = &staticArrayBuilder[T]{}
	//
	builder.heap8 = pool.NewBytePool[T]()
	builder.heap16 = pool.NewWordPool[T]()
	builder.heap = pool.NewSharedIndex[T]()
	//
	return builder
}

// staticArrayBuilder is for handling static words only.
type staticArrayBuilder[T word.Word[T]] struct {
	heap8  pool.SmallPool[uint8, T]
	heap16 pool.SmallPool[uint16, T]
	heap   *pool.SharedIndex[T]
}

// NewArray constructs a new word array with a given capacity.
func (p *staticArrayBuilder[T]) NewArray(height uint, bitwidth uint) MutArray[T] {
	switch {
	case bitwidth == 0:
		return NewZeroArray[T](height)
	case bitwidth == 1:
		return NewBitArray[T](height)
	case bitwidth <= 8:
		return NewPoolArray(height, bitwidth, p.heap8)
	case bitwidth <= 16:
		return NewPoolArray(height, bitwidth, p.heap16)
	default:
		// FIXME: for now, this actually defeats the only purpose of the shared
		// array builder.  Each array getting its own heap is sub-optimal.
		// However, at this stage, this is done for performance reasons.
		return NewPoolArray(height, bitwidth, pool.NewLocalIndex[T]())
	}
}

// ============================================================================
// Dynamic Builder
// ============================================================================

// NewDynamicBuilder constructs a new array builder for dynamic words.
func NewDynamicBuilder[T word.DynamicWord[T], P pool.Pool[uint32, T]](heap P) DynamicBuilder[T, P] {
	return DynamicBuilder[T, P]{
		heap8:  pool.NewBytePool[T](),
		heap16: pool.NewWordPool[T](),
		heap:   heap,
	}
}

// DynamicBuilder is for handling dynamic words only.
type DynamicBuilder[T word.DynamicWord[T], P pool.Pool[uint32, T]] struct {
	heap8  pool.SmallPool[uint8, T]
	heap16 pool.SmallPool[uint16, T]
	heap   P
}

// NewArray constructs a new word array with a given capacity.
func (p *DynamicBuilder[T, P]) NewArray(height uint, bitwidth uint) MutArray[T] {
	switch {
	case bitwidth == 0:
		return NewZeroArray[T](height)
	case bitwidth == 1:
		return NewBitArray[T](height)
	case bitwidth <= 8:
		return NewPoolArray(height, bitwidth, p.heap8)
	case bitwidth <= 16:
		return NewPoolArray(height, bitwidth, p.heap16)
	default:
		return NewPoolArray(height, bitwidth, p.heap)
	}
}

// Decode reconstructs an array from an array encoding, given the pool as it was
// when the encoding was made.
func (p *DynamicBuilder[T, P]) Decode(encoding Encoding) MutArray[T] {
	switch encoding.OpCode() {
	case ENCODING_POOL:
		return decode_pool(encoding, *p)
	default:
		panic("todo")
	}
}

// Encode a given array as a sequence of bytes suitable for serialisation.
func (p *DynamicBuilder[T, P]) Encode(array Array[T]) Encoding {
	var encoding Encoding
	//
	switch t := array.(type) {
	case *PoolArray[uint8, T, pool.SmallPool[uint8, T]]:
		encoding.Bytes = encode_smallpool8(t)
		encoding.Set(ENCODING_POOL, uint16(t.bitwidth))
	case *PoolArray[uint16, T, pool.SmallPool[uint16, T]]:
		encoding.Bytes = encode_smallpool16(t)
		encoding.Set(ENCODING_POOL, uint16(t.bitwidth))
	case *PoolArray[uint32, T, P]:
		encoding.Bytes = encode_pool(t)
		encoding.Set(ENCODING_POOL, uint16(t.bitwidth))
	default:
		panic(fmt.Sprintf("unknown array type: %s", reflect.TypeOf(t).String()))
	}
	//
	return encoding
}
