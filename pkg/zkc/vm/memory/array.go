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
package memory

// Array is a flat-slice implementation of ReadOnlyMemory backed by a []W.
// Reads are performed by delegating address decoding to a D (an AddressDecoder)
// which translates the incoming multi-word address tuple into a (start, end)
// index range, and then returning the corresponding sub-slice of the backing
// data.
//
// The type parameter W is the word type (e.g. a field element or big.Int), and
// D is the AddressDecoder strategy that encodes the layout of rows within the
// flat slice.
type Array[W any, D AddressDecoder[W]] struct {
	name    string
	decoder D
	data    []W
}

// NewArray constructs an Array with the given name and decoder.  The optional
// init values are used as the initial contents of the backing slice.
func NewArray[W any, D AddressDecoder[W]](name string, decoder D, init ...W) *Array[W, D] {
	return &Array[W, D]{
		name,
		decoder,
		init,
	}
}

// Name implementation for ReadOnlyMemory interface.
func (p *Array[W, D]) Name() string {
	return p.name
}

// Read implementation for ReadOnlyMemory interface.
func (p *Array[W, D]) Read(address []W) []W {
	var (
		start, end = p.decoder.Decode(address)
	)
	// Slice out relevant section
	return p.data[start:end]
}

// Write implementation for ReadOnlyMemory interface.
func (p *Array[W, D]) Write(address []W, data []W) {
	panic("todo")
}

// Contents implementation for Memory interface.
func (p *Array[W, D]) Contents() []W {
	return p.data
}
