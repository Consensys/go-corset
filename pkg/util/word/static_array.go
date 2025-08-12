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

// StaticArray implements an array of elements simply using an underlying array.
type StaticArray[T Word[T]] struct {
	// The data stored in this column (as bytes).
	data []byte
	// Bitwidth of each word in this array
	bitwidth uint
	// Number of bytes required to hold a word
	bytewidth uint
}

// NewStaticArray constructs a new word array with a given capacity.
func NewStaticArray[T Word[T]](height uint, bitwidth uint) *StaticArray[T] {
	var (
		bytewidth = ByteWidth(bitwidth)
		elements  = make([]byte, height*bytewidth)
	)
	//
	return &StaticArray[T]{elements, bitwidth, bytewidth}
}

// Len returns the number of elements in this word array.
func (p *StaticArray[T]) Len() uint {
	//
	return uint(len(p.data)) / p.bytewidth
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *StaticArray[T]) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *StaticArray[T]) Get(index uint) T {
	var (
		item   T
		offset = index * p.bytewidth
		bytes  = p.data[offset : offset+p.bytewidth]
	)
	//
	return item.SetBytes(bytes)
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *StaticArray[T]) Set(index uint, word T) {
	var (
		offset = index * p.bytewidth
		bytes  = p.data[offset : offset+p.bytewidth]
	)
	// Copy over
	word.PutBytes(bytes)
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *StaticArray[T]) Clone() array.MutArray[T] {
	// Allocate sufficient memory
	ndata := make([]byte, uint(len(p.data)))
	// Copy over the data
	copy(ndata, p.data)
	//
	return &StaticArray[T]{ndata, p.bitwidth, p.bytewidth}
}

// Slice out a subregion of this array.
func (p *StaticArray[T]) Slice(start uint, end uint) array.Array[T] {
	return &StaticArray[T]{
		p.data[start*p.bytewidth : end*p.bytewidth], p.bitwidth, p.bytewidth,
	}
}

// Pad prepend array with n copies and append with m copies of the given padding
// value.
func (p *StaticArray[T]) Pad(n uint, m uint, padding T) {
	var (
		// Determine new length
		l = n + m + p.Len()
		// Initialise new array
		data = make([]byte, p.bytewidth*l)
	)
	// copy
	copy(data[n*p.bytewidth:], p.data)
	p.data = data
	// Front padding!
	for i := range n {
		p.Set(i, padding)
	}
	// Back padding!
	for i := l - m; i < l; i++ {
		p.Set(i, padding)
	}
}

func (p *StaticArray[T]) String() string {
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
