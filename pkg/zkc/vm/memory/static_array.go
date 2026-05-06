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
	"slices"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util"
)

// StaticArray is a memory implementation backed by a fixed-size []W, meaning
// that an out-of-bound read will panic. Reads are performed by delegating
// address decoding to a D (an AddressDecoder) which translates the incoming
// multi-word address tuple into a (start, end) index range, and then returning
// the corresponding sub-slice of the backing data.
//
// The type parameter W is the word type (e.g. a field element or big.Int), and
// D is the AddressDecoder strategy that encodes the layout of rows within the
// flat slice.
type StaticArray[W util.Uinter64] struct {
	geometry Geometry[W]
	name     string
	data     []W
}

// newStaticArray constructs a new array initialised with a given set of values.
func newStaticArray[W util.Uinter64](name string, registers []register.Register, init ...W) StaticArray[W] {
	var geometry = NewGeometry[W](registers)
	//
	return StaticArray[W]{geometry, name, init}
}

// Name implementation for Memory interface.
func (p *StaticArray[W]) Name() string {
	return p.name
}

// Initialise implementation for Memory interface.
func (p *StaticArray[W]) Initialise(contents []W) {
	p.data = contents
}

// Geometry implementation for Memory interface.
func (p *StaticArray[W]) Geometry() Geometry[W] {
	return p.geometry
}

// Read implementation for Memory interface.
func (p *StaticArray[W]) Read(frame []W, address []register.Id, data []register.Id) error {
	var start, _ = p.geometry.FrameDecode(frame, address)
	//
	for i := range data {
		frame[data[i].Unwrap()] = p.data[uint64(i)+start]
	}
	//
	return nil
}

// Write implementation for Memory interface.
func (p *StaticArray[W]) Write(frame []W, address []register.Id, data []register.Id) error {
	var start, end = p.geometry.FrameDecode(frame, address)
	// expand memory if needed
	p.data = expand(p.data, end)
	// copy over data
	for i := range data {
		p.data[uint64(i)+start] = frame[data[i].Unwrap()]
	}
	//
	return nil
}

// Expand a slice to ensure it has at least length n.  If the slice already has
// at least n elements it is returned as-is.  Otherwise capacity is grown if
// needed (via slices.Grow, which uses the runtime's append-style growth
// heuristic) and the length is extended to n.
func expand[T any](slice []T, n uint64) []T {
	m := uint64(len(slice))
	if n <= m {
		return slice
	}
	//
	return slices.Grow(slice, int(n-m))[:n]
}
