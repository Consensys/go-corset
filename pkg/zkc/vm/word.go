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
package vm

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/zkc/vm/internal/word"
)

// Word abstracts the data type (a.k.a the "machine word") used for holding
// values within the machine.  The reason for abstracting this concept is to
// allow a machine compiled for a larger word size to be automatically lowered
// to a machine for a smaller word size.  For example, our source program might
// be written for a 64bit machine and we wish to executed it on 16bit machine
// (i.e. because our target field configuration has a maximum register size of
// 16bits).
type Word[W any] = word.Word[W]

// Uint represents an unbound unsigned integer.
type Uint = word.Uint

// ============================================================================
// Constructors
// ============================================================================

// Uint64 initialises a given word with a 64bit value.  This will panic if the
// given value exceeds the available bandwidth of the word in question.
func Uint64[W Word[W]](val uint64) W {
	var w W
	return w.SetUint64(val)
}

// ============================================================================
// Decoding
// ============================================================================

// DecodeBytes decodes the given set of bytes as big integer values according to
// the given register type(s).  Observe that values are assumed to be packed tightly
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
// If the input array is not a multiple of the bitwidth
func DecodeBytes[W Word[W]](bytes []byte, registers []register.Register) []W {
	var (
		bitwidth = bitwidthOf(registers)
		// Initially empty buffer which is expanded as necessary to accommodate
		// reading bits of the given data types.
		buffer []byte
	)
	// Decode array into
	values, _ := bit.DecodeArray(bitwidth, bytes, func(bytes []byte) (ints []big.Int) {
		var (
			reader = bit.NewReader(bytes)
		)
		// Decode the type using the given buffer
		for _, t := range registers {
			var vs []big.Int
			//
			vs, buffer = DecodeUnsignedInt(t.Width(), &reader, buffer)
			//
			ints = append(ints, vs...)
		}
		// Done
		return ints
	})
	// Flatten decoded tuples
	return array.FlatMap(values, func(ints []big.Int) []W {
		var words = make([]W, len(ints))
		//
		for i, v := range ints {
			var ith W
			//
			words[i] = ith.SetBigInt(&v)
		}
		//
		return words
	})
}

// EncodeBytes encodes the given set of word values as packed bytes according to
// the given registers type(s). This is the inverse of DecodeAll.  Consider the
// following input array of u4 values:
//
// |  0  |  1  |  2  |  3  |  4  |  5  |  6  |  7  |
// +-----+-----+-----+-----+-----+-----+-----+-----+
// | 0x3 | 0x1 | 0xf | 0x0 | 0x0 | 0xe | 0x1 | 0xd |
//
// Then, encoding this as a u4 array will produce the following bytes:
//
// |  00  |  01  |  02  |  03  |
// +------+------+------+------+
// | 0x31 | 0xf0 | 0x0e | 0x1d |
func EncodeBytes[W Word[W]](values []W, registers []register.Register) []byte {
	var (
		bitwidth   = bitwidthOf(registers)
		nElems     = uint(len(values))
		totalBits  = nElems * bitwidth
		totalBytes = (totalBits + 7) / 8
		result     = make([]byte, totalBytes)
		n          = bit.BytesRequiredFor(bitwidth)
		buf        = make([]byte, n)
		offset     uint
	)
	//
	for _, v := range values {
		for _, r := range registers {
			EncodeUnsignedInt(r.Width(), v.BigInt(), buf)
			bit.BigEndianCopy(buf, 0, result, offset, r.Width())
			offset += r.Width()
		}
	}
	//
	return buf
}

// ============================================================================
// Decoding
// ============================================================================

// DecodeUnsignedInt decodes the given set of bytes as big integer values
// according to the given register type(s).  Observe that values are assumed to
// be packed tightly (i.e. without any padding).  Consider the following input
// byte array:
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
// If the input array is not a multiple of the bitwidth
func DecodeUnsignedInt(bitwidth uint, reader *bit.Reader, buffer []byte) ([]big.Int, []byte) {
	var (
		val big.Int
		// Determine number of bytes required to hold value
		n = bit.BytesRequiredFor(bitwidth)
		// Calculate excess bits (needed for alignment)
		m = (n * 8) - bitwidth
	)
	// Expand buffer to ensure enough space
	buffer = expandBufferAsNeeded(bitwidth, buffer)
	// Read bitwidth bits out
	reader.BigEndianReadInto(bitwidth, buffer)
	// Assign (unaligned) bytes
	val.SetBytes(buffer[:n])
	// Right shift to fix alignment
	val.Rsh(&val, m)
	//
	return []big.Int{val}, buffer
}

// ============================================================================
// Encoding
// ============================================================================

// EncodeUnsignedInt encodes the given big integer as packed bytes according to
// the given bitwidths.  This is the inverse of DecodeUnsignedInt.
//
// |  0  |  1  |  2  |  3  |  4  |  5  |  6  |  7  |
// +-----+-----+-----+-----+-----+-----+-----+-----+
// | 0x3 | 0x1 | 0xf | 0x0 | 0x0 | 0xe | 0x1 | 0xd |
//
// Then, encoding this as a u4 array will produce the following bytes:
//
// |  00  |  01  |  02  |  03  |
// +------+------+------+------+
// | 0x31 | 0xf0 | 0x0e | 0x1d |
func EncodeUnsignedInt(bitwidth uint, v *big.Int, buf []byte) {
	var (
		w big.Int
		// Determine number of bytes required to hold value
		n = bit.BytesRequiredFor(bitwidth)
		// Calculate excess bits (needed for alignment)
		m = (n * 8) - bitwidth
	)
	// Clear buffer
	for i := range buf {
		buf[i] = 0
	}
	// Left-shift to account for alignment
	w.Lsh(v, m)
	// Fill with big-endian bytes of v, right-aligned in buf
	valBytes := w.Bytes()
	//
	if len(valBytes) > 0 {
		copy(buf[n-uint(len(valBytes)):], valBytes)
	}
}

// ============================================================================
// Misc
// ============================================================================

func expandBufferAsNeeded(bitwidth uint, buffer []byte) []byte {
	var n = bit.BytesRequiredFor(bitwidth)
	//
	if uint(len(buffer)) >= n {
		return buffer
	}
	//
	return make([]byte, n)
}

func bitwidthOf(registers []register.Register) uint {
	var width uint
	//
	for _, r := range registers {
		if r.IsNative() {
			panic("cannot determine bitwidth of native register")
		}
		//
		width += r.Width()
	}
	//
	return width
}
