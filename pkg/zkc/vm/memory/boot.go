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
	"math/big"

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Boot is the concrete memory type used by the boot machine: a pointer to a
// flat Array of big.Int words addressed via an AddressDecoder.
type Boot = *Array[big.Int, BootDecoder]

// BootDecoder translates a multi-dimensional logical address into the half-open
// index range [start, end) within the backing flat slice of a memory.Array.
// The address tuple arrives as a slice of big.Int values, decoded from the
// memory's address data type.  addressGeometry records the bit width of each
// address component; dataGeometry records how many data words make up a single
// row, so that multi-word rows are addressed contiguously.
type BootDecoder struct {
	addressGeometry []uint
	dataGeometry    uint
}

// NewBootDecoder constructs an AddressDecoder for a memory whose address bus
// has the given address type and whose data bus has the given data type.
// addressGeometry is populated by flattening the address type and collecting
// each leaf's bit width.  dataGeometry is the number of leaves produced by
// flattening the data type (i.e. the number of data words per row).
func NewBootDecoder(addressLines []variable.Descriptor, dataLines []variable.Descriptor) BootDecoder {
	var (
		addressGeometry []uint
		dataGeometry    uint
	)
	// flattern address lines
	for _, address := range addressLines {
		address.DataType.Flattern(address.Name, func(_ string, bitwidth uint) {
			addressGeometry = append(addressGeometry, bitwidth)
		})
	}
	// flattern data lines
	for _, data := range dataLines {
		data.DataType.Flattern(data.Name, func(_ string, _ uint) {
			dataGeometry++
		})
	}

	return BootDecoder{addressGeometry, dataGeometry}
}

// Decode maps address (a tuple of big.Int values representing a logical memory
// address) to the half-open index range [start, end) within the backing flat
// slice.  The length end-start always equals dataGeometry, i.e. the number of
// data words per row.
//
// The linear row index is computed by packing the address components
// big-endian: each component is shifted left by the total bit width of all
// subsequent components, then OR-ed in.  For a scalar address this reduces to
// index = address[0]; for a tuple (u8, u16) it gives
// index = address[0]<<16 | address[1].
func (p BootDecoder) Decode(address []big.Int) (uint64, uint64) {
	var index uint64
	for i, component := range address {
		index = (index << p.addressGeometry[i]) | component.Uint64()
	}

	start := index * uint64(p.dataGeometry)

	return start, start + uint64(p.dataGeometry)
}
