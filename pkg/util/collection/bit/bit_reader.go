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

// Reader provides a mechanism for reading bits from a given array of bytes.
type Reader struct {
	bitoffset uint
	bytes     []byte
}

// NewReader constructs a new bit reader.
func NewReader(bytes []byte) Reader {
	return Reader{0, bytes}
}

// ReadInto reads n bits from the underlying array into a given target array,
// returning the total number of bytes affected.
func (p *Reader) ReadInto(nbits uint, buf []byte) uint {
	var (
		aligned_offset = p.bitoffset%8 == 0
		aligned_count  = nbits%8 == 0
	)
	// Sanity check for fast cases
	if aligned_offset && aligned_count {
		return p.fullyAlignedRead(nbits/8, buf)
	} else if aligned_offset {
		return p.partiallyAlignedRead(nbits, buf)
	}
	// unaligned read
	n := bitcopy(p.bytes, p.bitoffset, buf, nbits)
	//
	p.bitoffset += nbits
	//
	return n
}

func (p *Reader) fullyAlignedRead(nbytes uint, buf []byte) uint {
	var (
		start = p.bitoffset / 8
		end   = start + nbytes
	)
	// Copy over bytes
	copy(buf, p.bytes[start:end])
	// Update the offset
	p.bitoffset += nbytes * 8
	// Done
	return nbytes
}

func (p *Reader) partiallyAlignedRead(nbits uint, buf []byte) uint {
	var (
		nbytes   = nbits / 8
		bitsLeft = nbits % 8
	)
	// Initial (fast) read
	p.fullyAlignedRead(nbytes, buf)
	// Final (tidy up) read
	bitcopy(p.bytes, p.bitoffset, buf[nbytes:], bitsLeft)
	// Update offset
	p.bitoffset += bitsLeft
	//
	return nbytes + 1
}

func bitcopy(src []byte, srcOffset uint, dst []byte, nbits uint) uint {
	var nread = nbits / 8
	// Determine how many bytes affected.
	if nbits%8 != 0 {
		// Clear final byte
		dst[nread] = 0
		nread++
	}
	//
	for i := range nbits {
		ith := readBit(src, srcOffset+i)
		writeBit(ith, dst, i)
	}
	//
	return nread
}

func readBit(src []byte, bitoffset uint) bool {
	var (
		byte = bitoffset / 8
		bit  = bitoffset % 8
		mask = uint8(1) << bit
	)
	//
	return src[byte]&mask != 0
}

func writeBit(val bool, src []byte, bitoffset uint) {
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
