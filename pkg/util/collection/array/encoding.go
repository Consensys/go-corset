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
package array

import (
	"encoding/binary"

	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ENCODING_CONSTANT is for arrays which hold constant values.
const ENCODING_CONSTANT = 0

// ENCODING_STATIC is for arrays which hold their values explicitly.
const ENCODING_STATIC = 1

// ENCODING_POOL is for arrays whose values are indexes into a pool.
const ENCODING_POOL = 2

// Pool provides a convenient alias
type Pool[T any] = pool.Pool[uint32, T]

// Encoding represents an encoded form of a word array useful for long term
// storage (e.g. in a file).
type Encoding struct {
	// Indicates what encoding method is used for this encoding.
	Encoding uint32
	// Bytes of the encoding itself
	Bytes []byte
}

// OpCode returns the instruction opcode for this encoding.
func (p *Encoding) OpCode() uint8 {
	return uint8(p.Encoding >> 24)
}

// Operand returns the instruction operand for this encoding.
func (p *Encoding) Operand() uint32 {
	// Operand is actually 24bits
	return p.Encoding & 0xFF_FFFF
}

// Set sets the instruction opcode for this encoding.
func (p *Encoding) Set(opcode uint8, operand uint32) {
	// Set new opcode & operand
	p.Encoding = (uint32(opcode) << 24) | (operand & 0xFF_FFFF)
}

// ============================================================================
// Constant Arrays
// ============================================================================

func decode_constant[T word.DynamicWord[T]](encoding Encoding) MutArray[T] {
	var (
		value    T
		len      = binary.BigEndian.Uint32(encoding.Bytes)
		constant = encoding.Operand()
		tmp      = constant
		bitwidth uint
	)
	// Determine bitwidth of constant
	for tmp > 0 {
		tmp >>= 1
		bitwidth++
	}
	//
	value = value.SetUint64(uint64(constant))
	//
	return NewConstantArray(uint(len), bitwidth, value)
}

func encode_constant[T word.DynamicWord[T]](array *ConstantArray[T]) []byte {
	var bytes [4]byte
	//
	binary.BigEndian.PutUint32(bytes[:], uint32(array.Len()))
	//
	return bytes[:]
}

// ============================================================================
// Static Arrays
// ============================================================================

func decode_static[T word.DynamicWord[T]](encoding Encoding) MutArray[T] {
	var bitwidth = encoding.Operand()
	//
	switch {
	case bitwidth == 1:
		return decode_bits[T](encoding.Bytes)
	case bitwidth <= 8:
		return decode_small8[T](encoding)
	case bitwidth <= 16:
		return decode_small16[T](encoding)
	case bitwidth <= 32:
		return decode_small32[T](encoding)
	default:
		panic("unsupported static array")
	}
}

func decode_bits[T word.DynamicWord[T]](bytes []byte) MutArray[T] {
	var (
		n      = uint(len(bytes) - 1)
		unused = uint(bytes[n])
		arr    BitArray[T]
	)
	//
	arr.data = bytes[:n]
	arr.height = uint(len(arr.data)*8) - unused
	//
	return &arr
}

func encode_bits[T word.DynamicWord[T]](array *BitArray[T]) []byte {
	var (
		unused = (len(array.data) * 8) - int(array.height)
	)
	//
	return append(array.data, uint8(unused))
}

// Decode an array of bytes into a given array.
func decode_small8[T word.DynamicWord[T]](encoding Encoding) MutArray[T] {
	var arr SmallArray[uint8, T]
	//
	arr.data = encoding.Bytes
	arr.bitwidth = uint(encoding.Operand())
	//
	return &arr
}

func encode_small8[T word.DynamicWord[T]](array *SmallArray[uint8, T]) []byte {
	return array.data
}

func decode_small16[T word.DynamicWord[T]](encoding Encoding) MutArray[T] {
	var (
		arr SmallArray[uint16, T]
		n   = uint(len(encoding.Bytes) / 2)
	)
	//
	arr.data = make([]uint16, n)
	arr.bitwidth = uint(encoding.Operand())
	//
	for i := range n {
		var (
			offset = i * 2
			b1     = uint16(encoding.Bytes[offset])
			b0     = uint16(encoding.Bytes[offset+1])
		)
		// Assign ith element
		arr.data[i] = (b1 << 8) | b0
	}
	//
	return &arr
}

func encode_small16[T word.DynamicWord[T]](array *SmallArray[uint16, T]) []byte {
	var bytes = make([]byte, array.Len()*2)
	//
	for i := range array.Len() {
		var (
			ith    = array.data[i]
			offset = i * 2
		)
		// big endian form
		bytes[offset] = uint8(ith >> 8)
		bytes[offset+1] = uint8(ith)
	}
	//
	return bytes
}

func decode_small32[T word.DynamicWord[T]](encoding Encoding) MutArray[T] {
	var (
		arr SmallArray[uint32, T]
		n   = uint(len(encoding.Bytes) / 4)
	)
	//
	arr.data = make([]uint32, n)
	arr.bitwidth = uint(encoding.Operand())
	//
	for i := range n {
		var (
			offset = i * 4
			b3     = uint32(encoding.Bytes[offset])
			b2     = uint32(encoding.Bytes[offset+1])
			b1     = uint32(encoding.Bytes[offset+2])
			b0     = uint32(encoding.Bytes[offset+3])
		)
		// Assign ith element
		arr.data[i] = (b3 << 24) | (b2 << 16) | (b1 << 8) | b0
	}
	//
	return &arr
}

// Encode returns the byte encoding of this array.
func encode_small32[T word.DynamicWord[T]](array *SmallArray[uint32, T]) []byte {
	var bytes = make([]byte, array.Len()*4)
	//
	for i := range array.Len() {
		var (
			ith    = array.data[i]
			offset = i * 4
		)
		// big endian form
		bytes[offset] = uint8(ith >> 24)
		bytes[offset+1] = uint8(ith >> 16)
		bytes[offset+2] = uint8(ith >> 8)
		bytes[offset+3] = uint8(ith)
	}
	//
	return bytes
}

// ============================================================================
// Pool Array
// ============================================================================

func decode_pool[T word.DynamicWord[T], P Pool[T]](encoding Encoding, builder DynamicBuilder[T, P]) MutArray[T] {
	var (
		arr PoolArray[uint32, T, P]
		n   = uint(len(encoding.Bytes) / 4)
	)
	//
	arr.index = make([]uint32, n)
	arr.bitwidth = uint(encoding.Operand())
	arr.pool = builder.heap
	//
	for i := range n {
		var (
			offset = i * 4
			b3     = uint32(encoding.Bytes[offset])
			b2     = uint32(encoding.Bytes[offset+1])
			b1     = uint32(encoding.Bytes[offset+2])
			b0     = uint32(encoding.Bytes[offset+3])
		)
		// Assign ith element
		arr.index[i] = (b3 << 24) + (b2 << 16) + (b1 << 8) + b0
	}
	//
	return &arr
}

// Encode returns the byte encoding of this array.
func encode_pool[T word.DynamicWord[T], P Pool[T]](array *PoolArray[uint32, T, P]) []byte {
	var bytes = make([]byte, array.Len()*4)
	//
	for i := range array.Len() {
		var (
			ith    = array.index[i]
			offset = i * 4
		)
		// big endian form
		bytes[offset] = uint8(ith >> 24)
		bytes[offset+1] = uint8(ith >> 16)
		bytes[offset+2] = uint8(ith >> 8)
		bytes[offset+3] = uint8(ith)
	}
	//
	return bytes
}
