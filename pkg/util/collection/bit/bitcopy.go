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

// Copy copies n bits starting a given bit offset from a given byte array
// source into a given destination (at a given offset).
func Copy(src []byte, srcOffset uint, dst []byte, dstOffset uint, nbits uint) {
	// Check for aligned read / write
	if srcOffset%8 == 0 && dstOffset%8 == 0 {
		var (
			srcByteOffset = srcOffset / 8
			dstByteOffset = dstOffset / 8
			nBytes        = nbits / 8
		)
		// Copy bytes
		copy(dst[dstByteOffset:dstByteOffset+nBytes], src[srcByteOffset:srcByteOffset+nBytes])
		// Calculate residue
		nbits = nbits % 8
		srcOffset += nBytes * 8
		dstOffset += nBytes * 8
	}
	// Continue with any remaining
	for i := range nbits {
		ith := Read(src, srcOffset+i)
		Write(ith, dst, dstOffset+i)
	}
}

// Read reads the bit at a given bit offset out of an array of bytes arranged in
// little endian format.
func Read(src []byte, bitoffset uint) bool {
	var (
		byte = bitoffset / 8
		bit  = bitoffset % 8
		mask = uint8(1) << bit
	)
	//
	return src[byte]&mask != 0
}

// Write writes a bit to a given bit offset in an array of bytes arranged in
// little endian format.
func Write(val bool, src []byte, bitoffset uint) {
	var (
		byte = bitoffset / 8
		bit  = bitoffset % 8
		mask = uint8(1) << bit
	)
	//
	if val {
		// set bit
		src[byte] = src[byte] | mask
	} else {
		// Clear bit
		src[byte] = src[byte] & ^mask
	}
}

// ReadBigEndian reads the bit at a given bit offset out of an array of bytes
// arranged in big endian format.
func ReadBigEndian(src []byte, bitoffset uint) bool {
	var (
		n    = uint(len(src))
		byte = bitoffset / 8
		bit  = bitoffset % 8
		mask = uint8(1) << bit
	)
	//
	return src[n-byte-1]&mask != 0
}
