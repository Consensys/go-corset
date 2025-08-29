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

// ZeroArray implements an array of single bit words simply using an underlying
// array of packed bytes.  That is, where eight bits are packed into a single
// byte.
type ZeroArray[T word.Word[T]] struct {
	// Actual height of column
	height uint
}

// NewZeroArray constructs a new word array with a given capacity.
func NewZeroArray[T word.Word[T]](height uint) *ZeroArray[T] {
	return &ZeroArray[T]{height}
}

// Append new word on this array
func (p *ZeroArray[T]) Append(word T) {
	p.Pad(0, 1, word)
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *ZeroArray[T]) Clone() MutArray[T] {
	return &ZeroArray[T]{p.height}
}

// Len returns the number of elements in this word array.
func (p *ZeroArray[T]) Len() uint {
	return p.height
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *ZeroArray[T]) BitWidth() uint {
	return 0
}

// Build implementation for the array.Builder interface.  This simply means that
// a static array is its own builder.
func (p *ZeroArray[T]) Build() Array[T] {
	return p
}

// Get returns the field element at the given index in this array.
func (p *ZeroArray[T]) Get(index uint) T {
	var b T
	// Default is zero
	return b
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *ZeroArray[T]) Set(index uint, word T) {
	// do nothing
}

// Pad implementation for MutArray interface.
func (p *ZeroArray[T]) Pad(n uint, m uint, padding T) {
	p.height += n + m
}

// Slice out a subregion of this array.
func (p *ZeroArray[T]) Slice(start uint, end uint) Array[T] {
	var (
		height = end - start
	)
	// Done
	return &ZeroArray[T]{height}
}

func (p *ZeroArray[T]) String() string {
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
