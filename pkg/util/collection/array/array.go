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
)

// Predicate abstracts the notion of a function which identifies something.
type Predicate[T any] func(T) bool

// Array provides a generice interface to an array of elements.  Typically, we
// are interested in arrays of field elements here.
type Array[T any] interface {
	// Returns the number of elements in this array.
	Len() uint
	// Get returns the element at the given index in this array.
	Get(uint) T
	// Set the element at the given index in this array, overwriting the
	// original value.
	Set(uint, T)
	// Clone makes clones of this array producing an otherwise identical copy.
	Clone() Array[T]
	// Slice out a subregion of this array.
	Slice(uint, uint) Array[T]
	// Return the number of bits required to store an element of this array.
	BitWidth() uint
	// Insert n copies of T at start of the array and m copies at the back
	// producing an updated array.
	Pad(uint, uint, T) Array[T]
}

// NewArray creates a new FrArray dynamically based on the given width.
func NewArray[T any](height uint, bitWidth uint) Array[T] {
	return NewRawArray[T](height, bitWidth)
}

// ----------------------------------------------------------------------------

// Raw implements an array of elements simply using an underlying array.
type Raw[T any] struct {
	// The data stored in this column (as bytes).
	elements []T
	// Maximum number of bits required to store an element of this array.
	bitwidth uint
}

// NewRawArray constructs a new field array with a given capacity.
func NewRawArray[T any](height uint, bitwidth uint) *Raw[T] {
	elements := make([]T, height)
	return &Raw[T]{elements, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *Raw[T]) Len() uint {
	return uint(len(p.elements))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *Raw[T]) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *Raw[T]) Get(index uint) T {
	return p.elements[index]
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *Raw[T]) Set(index uint, element T) {
	p.elements[index] = element
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *Raw[T]) Clone() Array[T] {
	// Allocate sufficient memory
	ndata := make([]T, uint(len(p.elements)))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &Raw[T]{ndata, p.bitwidth}
}

// Slice out a subregion of this array.
func (p *Raw[T]) Slice(start uint, end uint) Array[T] {
	return &Raw[T]{p.elements[start:end], p.bitwidth}
}

// Pad prepend array with n copies and append with m copies of the given padding
// value.
func (p *Raw[T]) Pad(n uint, m uint, padding T) Array[T] {
	l := uint(len(p.elements))
	// Allocate sufficient memory
	ndata := make([]T, l+n+m)
	// Copy over the data
	copy(ndata[n:], p.elements)
	// Front padding!
	for i := uint(0); i < n; i++ {
		ndata[i] = padding
	}
	// Back padding!
	for i := n + l; i < n+l+m; i++ {
		ndata[i] = padding
	}
	// Copy over
	return &Raw[T]{ndata, p.bitwidth}
}

func (p *Raw[T]) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := 0; i < len(p.elements); i++ {
		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(fmt.Sprintf("%v", p.elements[i]))
	}

	sb.WriteString("]")

	return sb.String()
}
