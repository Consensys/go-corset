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

type indexEntry struct {
	offset uint32
	length uint32
}

// IndexArray implements an array of elements simply using an underlying array.
type IndexArray[T Word[T]] struct {
	// heap of bytes
	heap []byte
	// indices into heap
	index []indexEntry
	// Bitwidth of words in this array
	bitwidth uint
	// Number of bytes required to hold a word
	bytewidth uint
}

// NewIndexArray constructs a new word array with a given capacity.
func NewIndexArray[T Word[T]](height uint, bitwidth uint) *IndexArray[T] {
	var (
		bytewidth = ByteWidth(bitwidth)
		heap      = []byte{}
		index     = make([]indexEntry, height)
	)
	//
	return &IndexArray[T]{heap, index, bitwidth, bytewidth}
}

// Len returns the number of elements in this word array.
func (p *IndexArray[T]) Len() uint {
	return uint(len(p.index))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *IndexArray[T]) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *IndexArray[T]) Get(index uint) T {
	var (
		item  T
		entry = p.index[index]
		bytes = p.heap[entry.offset : entry.offset+entry.length]
	)
	//
	return item.Set(bytes)
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *IndexArray[T]) Set(index uint, word T) {
	panic("todo")
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *IndexArray[T]) Clone() array.MutArray[T] {
	panic("todo")
}

// Slice out a subregion of this array.
func (p *IndexArray[T]) Slice(start uint, end uint) array.MutArray[T] {
	panic("todo")
}

// Pad prepend array with n copies and append with m copies of the given padding
// value.
func (p *IndexArray[T]) Pad(n uint, m uint, padding T) array.MutArray[T] {
	panic("todo")
}

func (p *IndexArray[T]) String() string {
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
