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
func (p *Encoding) Operand() uint16 {
	return uint16(p.Encoding)
}

// Set sets the instruction opcode for this encoding.
func (p *Encoding) Set(opcode uint8, operand uint16) {
	// Clear existing opcode
	p.Encoding = p.Encoding & 0xFF_FFFF
	// Set new opcode
	p.Encoding = p.Encoding | (uint32(opcode) << 24)
	// Set new operand
	p.Encoding = p.Encoding | uint32(operand)
}

// ============================================================================
// Pool Array
// ============================================================================

func decode_pool[T word.DynamicWord[T], P Pool[T]](encoding Encoding, builder DynamicBuilder[T, P]) MutArray[T] {
	var bitwidth = encoding.Operand()
	//
	switch {
	case bitwidth <= 8:
		return decode_smallpool8(encoding, builder)
	case bitwidth <= 16:
		return decode_smallpool16(encoding, builder)
	default:
		return decode_pool32(encoding, builder)
	}
}

// Decode an array of bytes into a given array.
func decode_smallpool8[T word.DynamicWord[T], P Pool[T]](encoding Encoding, builder DynamicBuilder[T, P]) MutArray[T] {
	var arr PoolArray[uint8, T, pool.SmallPool[uint8, T]]
	//
	arr.index = encoding.Bytes
	arr.bitwidth = uint(encoding.Operand())
	arr.pool = builder.heap8
	//
	return &arr
}

func decode_smallpool16[T word.DynamicWord[T], P Pool[T]](encoding Encoding, builder DynamicBuilder[T, P]) MutArray[T] {
	var (
		arr PoolArray[uint16, T, pool.SmallPool[uint16, T]]
		n   = uint(len(encoding.Bytes) / 2)
	)
	//
	arr.index = make([]uint16, n)
	arr.bitwidth = uint(encoding.Operand())
	arr.pool = builder.heap16
	//
	for i := range n {
		var (
			offset = i * 2
			high   = uint16(encoding.Bytes[offset])
			low    = uint16(encoding.Bytes[offset+1])
		)
		// Assign ith element
		arr.index[i] = (high << 8) + low
	}
	//
	return &arr
}

func decode_pool32[T word.DynamicWord[T], P Pool[T]](encoding Encoding, builder DynamicBuilder[T, P]) MutArray[T] {
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

func encode_smallpool8[T word.DynamicWord[T], P pool.Pool[uint8, T]](array *PoolArray[uint8, T, P]) []byte {
	return array.index
}

func encode_smallpool16[T word.DynamicWord[T], P pool.Pool[uint16, T]](array *PoolArray[uint16, T, P]) []byte {
	var bytes = make([]byte, array.Len()*2)
	//
	for i := range array.Len() {
		var (
			ith    = array.index[i]
			offset = i * 2
			low    = uint8(ith)
			high   = uint8(ith >> 8)
		)
		// big endian form
		bytes[offset] = high
		bytes[offset+1] = low
	}
	//
	return bytes
}
