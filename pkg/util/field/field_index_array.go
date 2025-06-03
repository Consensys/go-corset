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
package field

import (
	"io"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// FrIndexArray provides an "indexed" array of field elements.  This applies two
// specific optimisations: (1) elements requiring only 16bits are optimised
// using a precomputed table; (2) otherwise, elements are referred to by index.
type FrIndexArray struct {
	// The data stored in this column (as indexes into the heap).
	elements []uint32
	// The set of elements in this array
	heap []fr.Element
	// Pool records value allocated to heap.  This helps ensure heap indices are
	// reused.
	pool map[[4]uint64]uint32
	// Maximum number of bits required to store an element of this array.
	bitwidth uint
}

// NewFrIndexArray constructs a new field array with a given capacity.
func NewFrIndexArray(height uint, bitwidth uint) *FrIndexArray {
	elements := make([]uint32, height)
	// NOTE: initial element is required here as otherwise, when height != 0, we
	// would be in an invalid state.
	heap := make([]fr.Element, 1)
	pool := make(map[[4]uint64]uint32)
	// Initialise pool
	pool[heap[0]] = 0
	//
	return &FrIndexArray{elements, heap, pool, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrIndexArray) Len() uint {
	return uint(len(p.elements))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *FrIndexArray) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *FrIndexArray) Get(index uint) fr.Element {
	return p.heap[p.elements[index]]
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrIndexArray) Set(index uint, element fr.Element) {
	// Lookup element in pool
	offset, ok := p.pool[element]
	// Check whether allocated already, or not
	if !ok {
		// Not allocated, so allocate now.
		offset = uint32(len(p.heap))
		p.pool[element] = offset
		p.heap = append(p.heap, element)
	}
	// Assign element
	p.elements[index] = offset
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrIndexArray) Clone() util.Array[fr.Element] {
	// Allocate sufficient memory
	elements := make([]uint32, len(p.elements))
	heap := make([]fr.Element, len(p.heap))
	pool := make(map[[4]uint64]uint32, len(p.pool))
	// Copy over the data
	copy(elements, p.elements)
	copy(heap, p.heap)
	// Initialise pool
	for i, e := range heap {
		pool[e] = uint32(i)
	}
	//
	return &FrIndexArray{elements, heap, pool, p.bitwidth}
}

// Slice out a subregion of this array.
func (p *FrIndexArray) Slice(start uint, end uint) util.Array[fr.Element] {
	// could be more efficient.
	elements := make([]uint32, len(p.elements))
	heap := make([]fr.Element, len(p.heap))
	pool := make(map[[4]uint64]uint32, len(p.pool))
	// Copy over the data
	copy(elements, p.elements)
	copy(heap, p.heap)
	// Initialise pool
	for i, e := range heap {
		pool[e] = uint32(i)
	}
	//
	return &FrIndexArray{elements[start:end], heap, pool, p.bitwidth}
}

// Pad prepend array with n copies and append with m copies of the given padding
// value.
func (p *FrIndexArray) Pad(n uint, m uint, padding fr.Element) util.Array[fr.Element] {
	l := uint(len(p.elements))
	// Allocate sufficient memory
	elements := make([]uint32, l+n+m)
	heap := make([]fr.Element, len(p.heap))
	pool := make(map[[4]uint64]uint32, len(p.pool))
	// Copy over the data
	copy(elements[n:], p.elements)
	copy(heap, p.heap)
	// Initialise pool
	for i, e := range heap {
		pool[e] = uint32(i)
	}
	//
	narr := &FrIndexArray{elements, heap, pool, p.bitwidth}
	// Front padding!
	for i := uint(0); i < n; i++ {
		narr.Set(i, padding)
	}
	// Back padding!
	for i := n + l; i < l+n+m; i++ {
		narr.Set(i, padding)
	}
	// Copy over
	return narr
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrIndexArray) Write(w io.Writer) error {
	// Determine bytewidth
	byteWidth := p.bitwidth / 8
	if p.bitwidth%8 != 0 {
		byteWidth++
	}
	//
	n := 32 - byteWidth
	//
	for _, e := range p.elements {
		element := p.heap[e]
		// Read exactly 32 bytes
		bytes := element.Bytes()
		// Write them out
		if _, err := w.Write(bytes[n:]); err != nil {
			return err
		}
	}
	//
	return nil
}

func (p *FrIndexArray) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := uint(0); i < p.Len(); i++ {
		if i != 0 {
			sb.WriteString(",")
		}

		ith := p.Get(i)
		sb.WriteString(ith.String())
	}

	sb.WriteString("]")

	return sb.String()
}
