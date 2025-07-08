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
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util/collection/array"
)

// FrArray represents an array of field elements.
type FrArray = array.Array[fr.Element]

// NewFrArray creates a new FrArray dynamically based on the given width.
func NewFrArray(height uint, bitWidth uint) FrArray {
	switch {
	case bitWidth <= 1:
		var pool = NewFrBitPool()
		return NewFrPoolArray(height, bitWidth, pool)
	case bitWidth <= 8:
		var pool = NewFrIndexPool[uint8]()
		return NewFrPoolArray(height, bitWidth, pool)
	case bitWidth <= 16:
		var pool = NewFrIndexPool[uint16]()
		return NewFrPoolArray(height, bitWidth, pool)
	default:
		return NewFrIndexArray(height, bitWidth)
	}
}

// FrArrayFromBigInts converts an array of big integers into an array of
// field elements.
func FrArrayFromBigInts(bitWidth uint, ints []big.Int) FrArray {
	elements := NewFrArray(uint(len(ints)), bitWidth)
	// Convert each integer in turn.
	for i, v := range ints {
		var element fr.Element

		element.SetBigInt(&v)
		elements.Set(uint(i), element)
	}

	// Done.
	return elements
}

// ----------------------------------------------------------------------------

// FrElementArray implements an array of field elements using an underlying
// byte array.  Each element occupies a fixed number of bytes, known as the
// width.  This is space efficient when a known upper bound holds for the given
// elements.  For example, when storing elements which always fit within 16bits,
// etc.
type FrElementArray struct {
	// The data stored in this column (as bytes).
	elements []fr.Element
	// Maximum number of bits required to store an element of this array.
	bitwidth uint
}

// NewFrElementArray constructs a new field array with a given capacity.
func NewFrElementArray(height uint, bitwidth uint) *FrElementArray {
	elements := make([]fr.Element, height)
	return &FrElementArray{elements, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrElementArray) Len() uint {
	return uint(len(p.elements))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *FrElementArray) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *FrElementArray) Get(index uint) fr.Element {
	return p.elements[index]
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrElementArray) Set(index uint, element fr.Element) {
	p.elements[index] = element
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrElementArray) Clone() array.Array[fr.Element] {
	// Allocate sufficient memory
	ndata := make([]fr.Element, uint(len(p.elements)))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &FrElementArray{ndata, p.bitwidth}
}

// Slice out a subregion of this array.
func (p *FrElementArray) Slice(start uint, end uint) array.Array[fr.Element] {
	return &FrElementArray{p.elements[start:end], p.bitwidth}
}

// Pad prepend array with n copies and append with m copies of the given padding
// value.
func (p *FrElementArray) Pad(n uint, m uint, padding fr.Element) array.Array[fr.Element] {
	l := uint(len(p.elements))
	// Allocate sufficient memory
	ndata := make([]fr.Element, l+n+m)
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
	return &FrElementArray{ndata, p.bitwidth}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrElementArray) Write(w io.Writer) error {
	for _, e := range p.elements {
		// Read exactly 32 bytes
		bytes := e.Bytes()
		// Write them out
		if _, err := w.Write(bytes[:]); err != nil {
			return err
		}
	}
	//
	return nil
}

func (p *FrElementArray) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := 0; i < len(p.elements); i++ {
		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(p.elements[i].String())
	}

	sb.WriteString("]")

	return sb.String()
}
