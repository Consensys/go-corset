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
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/util/collection/array"
)

// IndexArray implements an array of elements simply using an underlying array.
type IndexArray[T Word[T], P Pool[uint, T]] struct {
	// pool of values
	pool P
	// indices into pool
	index []uint
	// Bitwidth of words in this array
	bitwidth uint
}

// NewIndexArray constructs a new indexed array.
func NewIndexArray[T Word[T], P Pool[uint, T]](height uint, bitwidth uint, pool P) *IndexArray[T, P] {
	var (
		undefined      T
		undefinedIndex = pool.Put(undefined)
		index          = make([]uint, height)
	)
	//
	for i := range height {
		index[i] = undefinedIndex
	}
	//
	return &IndexArray[T, P]{pool, index, bitwidth}
}

// Build implementation for the array.Builder interface.  This simply means that
// a index array is its own builder.
func (p *IndexArray[T, P]) Build() array.Array[T] {
	return p
}

// Len returns the number of elements in this word array.
func (p *IndexArray[T, P]) Len() uint {
	return uint(len(p.index))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *IndexArray[T, P]) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *IndexArray[T, P]) Get(index uint) T {
	return p.pool.Get(p.index[index])
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *IndexArray[T, P]) Set(index uint, word T) {
	p.index[index] = p.pool.Put(word)
}

// Slice out a subregion of this array.
func (p *IndexArray[T, P]) Slice(start uint, end uint) array.Array[T] {
	return &IndexArray[T, P]{
		p.pool,
		p.index[start:end],
		p.bitwidth,
	}
}

func (p *IndexArray[T, P]) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := range p.Len() {
		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(fmt.Sprintf("%v", p.Get(i)))
	}

	sb.WriteString("]")

	return sb.String()
}
