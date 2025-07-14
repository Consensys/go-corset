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
package bit

// LeastSignificantReader provides a mechanism for reading bits from a given
// array of bytes, where the least significant bits are read first.
type LeastSignificantReader struct {
	bitoffset uint
	bytes     []byte
}

// NewLeastSignificantReader constructs a new bit reader.
func NewLeastSignificantReader(bytes []byte) LeastSignificantReader {
	return LeastSignificantReader{0, bytes}
}

// ReadInto reads the n least significant bits from the underlying array into a
// given target array, returning the total number of bytes affected.
func (p *LeastSignificantReader) ReadInto(nbits uint, buf []byte) uint {
	var nread = nbits / 8
	// Determine how many bytes affected.
	if nbits%8 != 0 {
		// Clear final byte
		buf[nread] = 0
		nread++
	}
	//
	bitCopy(p.bytes, p.bitoffset, buf, 0, nbits)
	//
	p.bitoffset += nbits
	//
	return nread
}
