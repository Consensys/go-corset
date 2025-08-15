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
	"github.com/consensys/go-corset/pkg/util/collection/array"
)

// ArrayBuilder is a mechanism for constructing arrays which aims to select the
// right representation for a given array.
type ArrayBuilder[T any] interface {
	// Clone this array builder, including any pools contained therein.
	Clone() ArrayBuilder[T]
	// NewArray constructs a new array of the given height holding elements of the given bitwidth
	NewArray(height uint, bitwidth uint) array.MutArray[T]
}

// NewDynamicArrayBuilder constructs a new array builder for dynamic words.
func NewDynamicArrayBuilder[T DynamicWord[T]]() ArrayBuilder[T] {
	var builder = &dynamicArrayBuilder[T]{}
	//
	builder.heap8 = NewBytePool[T]()
	builder.heap16 = NewWordPool[T]()
	builder.heap = NewSharedHeap[T]()
	//
	return builder
}

// NewStaticArrayBuilder constructs a new array builder for dynamic words.
func NewStaticArrayBuilder[T Word[T]]() ArrayBuilder[T] {
	var builder = &staticArrayBuilder[T]{}
	//
	builder.heap8 = NewBytePool[T]()
	builder.heap16 = NewWordPool[T]()
	builder.heap = NewSharedIndex[T]()
	//
	return builder
}

// dynamicArrayBuilder is for handling static words only.
type dynamicArrayBuilder[T DynamicWord[T]] struct {
	heap8  SmallPool[uint8, T]
	heap16 SmallPool[uint16, T]
	heap   *SharedHeap[T]
}

// NewArray constructs a new word array with a given capacity.
func (p *dynamicArrayBuilder[T]) Clone() ArrayBuilder[T] {
	return &dynamicArrayBuilder[T]{
		p.heap8,
		p.heap16,
		p.heap.Clone(),
	}
}

// NewArray constructs a new word array with a given capacity.
func (p *dynamicArrayBuilder[T]) NewArray(height uint, bitwidth uint) array.MutArray[T] {
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

// staticArrayBuilder is for handling static words only.
type staticArrayBuilder[T Word[T]] struct {
	heap8  SmallPool[uint8, T]
	heap16 SmallPool[uint16, T]
	heap   *SharedIndex[T]
}

// NewArray constructs a new word array with a given capacity.
func (p *staticArrayBuilder[T]) Clone() ArrayBuilder[T] {
	return &staticArrayBuilder[T]{
		p.heap8,
		p.heap16,
		p.heap.Clone(),
	}
}

// NewArray constructs a new word array with a given capacity.
func (p *staticArrayBuilder[T]) NewArray(height uint, bitwidth uint) array.MutArray[T] {
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
		return NewPoolArray(height, bitwidth, NewLocalIndex[T]())
	}
}
