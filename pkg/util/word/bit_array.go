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
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// bitOne is the byte-level binary representation of 1.
var bitOne = []byte{1}

// =================================================================================
// Implementation
// =================================================================================

// BitArray implements an array of single bit words simply using an underlying
// array of packed bytes.  That is, where eight bits are packed into a single
// byte.
type BitArray[T Word[T]] struct {
	// The data stored in this column (as bytes).
	data []byte
	// Actual height of column
	height uint
}

// NewBitArray constructs a new word array with a given capacity.
func NewBitArray[T Word[T]](height uint) *BitArray[T] {
	var (
		bytewidth = ByteWidth(height)
		elements  = make([]byte, bytewidth)
	)
	//
	return &BitArray[T]{elements, height}
}

// Len returns the number of elements in this word array.
func (p *BitArray[T]) Len() uint {
	return p.height
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *BitArray[T]) BitWidth() uint {
	return 1
}

// Build implementation for the array.Builder interface.  This simply means that
// a static array is its own builder.
func (p *BitArray[T]) Build() array.Array[T] {
	return p
}

// Get returns the field element at the given index in this array.
func (p *BitArray[T]) Get(index uint) T {
	var b T
	//
	if bit.Read(p.data, index) {
		return b.Set(bitOne)
	}
	// Default is zero
	return b
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *BitArray[T]) Set(index uint, word T) {
	bit.Write(word.Bit(0), p.data, index)
}

// Slice out a subregion of this array.
func (p *BitArray[T]) Slice(start uint, end uint) array.Array[T] {
	var (
		height    = end - start
		bytewidth = ByteWidth(height)
	)
	// Check for aligned slice (since this is a fast case).
	if start%8 == 0 {
		// Yes, easy case
		start = start / 8
		//
		return &BitArray[T]{p.data[start : start+bytewidth], height}
	}
	// No, hard case.  We'll just do a bitcopy for now.  In theory we could
	// improve performance by allowing BitArray to have a starting offset.  But,
	// the use cases for Slice() are very limited at this time, so no need.
	bytes := make([]byte, bytewidth)
	// Copy height bits over
	bit.Copy(p.data, start, bytes, height)
	// Done
	return &BitArray[T]{bytes, height}
}

func (p *BitArray[T]) String() string {
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
