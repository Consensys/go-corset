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

// AddressDecoder translates a multi-word address tuple into a contiguous range
// within a flat data slice.  Memories in this VM address their contents using
// tuples of words (i.e. []W) rather than a single scalar, because a single
// field element may not be wide enough to express every valid address.  An
// AddressDecoder bridges that representation to the concrete (start, end) index
// pair needed to slice into the underlying []W array.
//
// The returned half-open interval [start, end) identifies the data words that
// correspond to the given address.  Both indices are measured in units of W
// (not bytes), so the caller can write data[start:end] directly.
//
// Implementations are free to impose whatever layout they require.  For
// example, a decoder for a memory that stores 4-word rows at consecutive
// indices might compute start = index*4 and end = index*4 + 4.
type AddressDecoder[W any] interface {
	// Decode maps address, a tuple of words representing a logical memory
	// address, to the half-open index range [start, end) within the backing
	// flat slice.  The length end-start must equal the width of a single row
	// in the memory.
	Decode(address []W) (start, end uint64)
}
