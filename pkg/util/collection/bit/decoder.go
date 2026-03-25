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

// DecodeArray decodes the given set of bytes into an array of arbitrary
// fixed-width values.  Observe that values are assumed to be packed tightly
// (i.e. without any padding).  Consider the following input byte array:
//
// |  00  |  01  |  02  |  03  |
// +------+------+------+------+
// | 0x31 | 0xf0 | 0x0e | 0x1d |
//
// Then, decoding this into a u4 array will produce the following:
//
// |  0  |  1  |  2  |  3  |  4  |  5  |  6  |  7  |
// +-----+-----+-----+-----+-----+-----+-----+-----+
// | 0x3 | 0x1 | 0xf | 0x0 | 0x0 | 0xe | 0x1 | 0xd |
//
// Finally, the number of unused (i.e. remaining) bits is returned (which will
// be zero if (len(bytes)*8)%bitwidth == 0).
//
// NOTE: the array parsed into the decoder function is reused across different
// elements and, hence, the decoder function should clone it if necessary.
func DecodeArray[T any](bitwidth uint, bytes []byte, decoder func([]byte) T) ([]T, uint) {
	var (
		// Calculate how size of final array
		nelems = (uint(len(bytes)) * 8) / bitwidth
		// Construct buffer of sufficient width
		buffer = NewBuffer(bitwidth)
		// Construct reader for bytes
		reader = NewReader(bytes)
		// Preallocate final array
		values = make([]T, nelems)
	)
	// Keep going whilst at least one element left
	for index := 0; reader.Remaining() >= bitwidth; index++ {
		reader.BigEndianReadInto(bitwidth, buffer)
		values[index] = decoder(buffer)
	}
	// Done
	return values, reader.Remaining()
}
