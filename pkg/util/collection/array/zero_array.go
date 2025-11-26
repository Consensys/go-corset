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
	"strings"

	"github.com/consensys/go-corset/pkg/util/word"
)

// =================================================================================
// Implementation
// =================================================================================

// ConstantArray implements an array of a constant value.
type ConstantArray[T word.Word[T]] struct {
	// Actual height of column
	height uint
	value  T
}

// NewConstantArray constructs a new word array with a given capacity.
func NewConstantArray[T word.Word[T]](height uint, value T) *ConstantArray[T] {
	return &ConstantArray[T]{height, value}
}

// Append new word on this array
func (p *ConstantArray[T]) Append(word T) {
	p.Pad(0, 1, word)
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *ConstantArray[T]) Clone() MutArray[T] {
	return &ConstantArray[T]{p.height, p.value}
}

// Len returns the number of elements in this word array.
func (p *ConstantArray[T]) Len() uint {
	return p.height
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *ConstantArray[T]) BitWidth() uint {
	return 0
}

// Build implementation for the array.Builder interface.  This simply means that
// a static array is its own builder.
func (p *ConstantArray[T]) Build() Array[T] {
	return p
}

// Get returns the field element at the given index in this array.
func (p *ConstantArray[T]) Get(index uint) T {
	return p.value
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *ConstantArray[T]) Set(index uint, word T) {
	// do nothing
}

// Pad implementation for MutArray interface.
func (p *ConstantArray[T]) Pad(n uint, m uint, padding T) {
	p.height += n + m
}

// Slice out a subregion of this array.
func (p *ConstantArray[T]) Slice(start uint, end uint) Array[T] {
	var (
		height = end - start
	)
	// Done
	return &ConstantArray[T]{height, p.value}
}

func (p *ConstantArray[T]) String() string {
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
