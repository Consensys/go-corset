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
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// DynamicArray is a memory implementation backed by a dynamically sizing []W,
// meaning that an out-of-bound read will return 0.  Reads are performed by
// delegating address decoding to a D (an AddressDecoder) which translates the
// incoming multi-word address tuple into a (start, end) index range, and then
// returning the corresponding sub-slice of the backing data.
//
// The type parameter W is the word type (e.g. a field element or big.Int), and
// D is the AddressDecoder strategy that encodes the layout of rows within the
// flat slice.
type DynamicArray[W word.Word[W]] struct {
	StaticArray[W]
}

// newDynamicArray constructs a new array initialised with a given set of values.
func newDynamicArray[W word.Word[W]](name string, registers []register.Register, init ...W) DynamicArray[W] {
	var geometry = NewGeometry[W](registers)
	//
	return DynamicArray[W]{StaticArray[W]{geometry, name, init}}
}

// Read implementation for Memory interface.
func (p *DynamicArray[W]) Read(address []W) []W {
	var (
		n          = uint64(len(p.data))
		start, end = p.geometry.Decode(address)
	)
	// Check for out-of-bounds read
	if end <= n {
		// In-bounds
		return p.data[start:end]
	}
	// Out-of-bounds case
	var (
		zero   W
		values []W
	)
	//
	if start < n {
		// Partially out-of-bounds
		values = p.data[start:n]
	}
	// done
	return array.BackPad(values, uint(end-start), zero)
}

// FrameRead implementation for Memory interface.
func (p *DynamicArray[W]) FrameRead(frame []W, address []register.Id, data []register.Id) error {
	var start, _ = p.geometry.FrameDecode(frame, address)
	//
	for i := range data {
		frame[data[i].Unwrap()] = p.read(uint64(i) + start)
	}
	//
	return nil
}

// Internal read function handles out-of-bounds accesses.
func (p *DynamicArray[W]) read(address uint64) W {
	if address < uint64(len(p.data)) {
		return p.data[address]
	}
	// out-of-bounds access
	var zero W
	//
	return zero
}
