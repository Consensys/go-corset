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

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Array is a flat-slice implementation of ReadOnlyMemory backed by a []W.
// Reads are performed by delegating address decoding to a D (an AddressDecoder)
// which translates the incoming multi-word address tuple into a (start, end)
// index range, and then returning the corresponding sub-slice of the backing
// data.
//
// The type parameter W is the word type (e.g. a field element or big.Int), and
// D is the AddressDecoder strategy that encodes the layout of rows within the
// flat slice.
type Array[W word.Word[W]] struct {
	geometry Geometry[W]
	name     string
	data     []W
}

// newArray constructs a new array initialised with a given set of values.
func newArray[W word.Word[W]](name string, registers []register.Register, init ...W) Array[W] {
	var geometry = NewGeometry[W](registers)
	//
	return Array[W]{geometry, name, init}
}

// Name implementation for Memory interface.
func (p *Array[W]) Name() string {
	return p.name
}

// Initialise implementation for Memory interface.
func (p *Array[W]) Initialise(contents []W) {
	p.data = contents
}

// Geometry implementation for Memory interface.
func (p *Array[W]) Geometry() Geometry[W] {
	return p.geometry
}

// Read implementation for Memory interface.
func (p *Array[W]) Read(address []W) []W {
	var start, end = p.geometry.Decode(address)
	//
	return p.data[start:end]
}

// FrameRead implementation for Memory interface.
func (p *Array[W]) FrameRead(frame []W, address []register.Id, data []register.Id) error {
	var start, _ = p.geometry.FrameDecode(frame, address)
	//
	for i := range data {
		frame[data[i].Unwrap()] = p.data[uint64(i)+start]
	}
	//
	return nil
}

// FrameWrite implementation for Memory interface.
func (p *Array[W]) FrameWrite(frame []W, address []register.Id, data []register.Id) error {
	var (
		n          = uint64(len(p.data))
		start, end = p.geometry.FrameDecode(frame, address)
	)
	// expand memory if needed
	if n <= end {
		ndata := make([]W, end)
		copy(ndata, p.data)
		p.data = ndata
	}
	//
	for i := range data {
		p.data[uint64(i)+start] = frame[data[i].Unwrap()]
	}
	//
	return nil
}

// Write implementation for Memory interface.
func (p *Array[W]) Write(address []W, data []W) {
	var (
		n          = uint64(len(p.data))
		start, end = p.geometry.Decode(address)
	)
	// expand memory if needed
	if n <= end {
		ndata := make([]W, end)
		copy(ndata, p.data)
		p.data = ndata
	}
	//
	copy(p.data[start:end], data)
}

// Contents implementation for Memory interface.
func (p *Array[W]) Contents() []W {
	return p.data
}
