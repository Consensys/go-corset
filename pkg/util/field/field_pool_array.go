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

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// ----------------------------------------------------------------------------

// FrPoolArray implements an array of field elements using an index to pool
// identical elements.  Specificially, an array of all elements upto a certain
// bound is used as the index.
type FrPoolArray[K any, P FrPool[K]] struct {
	pool P
	// Elements in this array, where each is an index into the pool.
	elements []K
	// Determines how many bits are required to hold an element of this array.
	bitwidth uint
}

// NewFrPoolArray constructs a new pooled field array using a given pool.
func NewFrPoolArray[K any, P FrPool[K]](height uint, bitwidth uint, pool P) *FrPoolArray[K, P] {
	// Create empty array
	elements := make([]K, height)
	// Done
	return &FrPoolArray[K, P]{pool, elements, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrPoolArray[K, P]) Len() uint {
	return uint(len(p.elements))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *FrPoolArray[K, P]) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
//
//nolint:revive
func (p *FrPoolArray[K, P]) Get(index uint) fr.Element {
	key := p.elements[index]
	return p.pool.Get(key)
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrPoolArray[K, P]) Set(index uint, element fr.Element) {
	p.elements[index] = p.pool.Put(element)
}

// Clone makes clones of this array producing an otherwise identical copy.
// nolint: revive
func (p *FrPoolArray[K, P]) Clone() util.Array[fr.Element] {
	// Allocate sufficient memory
	ndata := make([]K, len(p.elements))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &FrPoolArray[K, P]{p.pool, ndata, p.bitwidth}
}

// Slice out a subregion of this array.
func (p *FrPoolArray[K, P]) Slice(start uint, end uint) util.Array[fr.Element] {
	return &FrPoolArray[K, P]{p.pool, p.elements[start:end], p.bitwidth}
}

// Pad prepend array with n copies and append with m copies of the given padding
// value.
func (p *FrPoolArray[K, P]) Pad(n uint, m uint, padding fr.Element) util.Array[fr.Element] {
	l := uint(len(p.elements))
	// Ensure padding in pool
	key := p.pool.Put(padding)
	// Allocate sufficient memory
	nelements := make([]K, l+n+m)
	// Copy over the data
	copy(nelements[n:], p.elements)
	// Front padding!
	for i := uint(0); i < n; i++ {
		nelements[i] = key
	}
	// Back padding!
	for i := l + n; i < l+n+m; i++ {
		nelements[i] = key
	}
	// Copy over
	return &FrPoolArray[K, P]{p.pool, nelements, p.bitwidth}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrPoolArray[K, P]) Write(w io.Writer) error {
	// Determine bytewidth
	byteWidth := p.bitwidth / 8
	if p.bitwidth%8 != 0 {
		byteWidth++
	}
	//
	n := 32 - byteWidth
	//
	for _, i := range p.elements {
		ith := p.pool.Get(i)
		// Read exactly 32 bytes
		bytes := ith.Bytes()
		// Write them out
		if _, err := w.Write(bytes[n:]); err != nil {
			return err
		}
	}
	//
	return nil
}

//nolint:revive
func (p *FrPoolArray[K, P]) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := 0; i < len(p.elements); i++ {
		if i != 0 {
			sb.WriteString(",")
		}

		index := p.elements[i]
		ith := p.pool.Get(index)
		sb.WriteString(ith.String())
	}

	sb.WriteString("]")

	return sb.String()
}
