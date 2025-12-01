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
package codec

import (
	"encoding/binary"
	"fmt"
)

// DecodeU8Sparse decodes an array of bytes represented in a "sparse encoding".
// Specifically, data is encoded as an array of tuples (value, n) which
// represent n copies of the given value.
func DecodeU8Sparse[T uint8 | uint16 | uint32 | uint64](bytes []byte, blockSizeWidth uint) []T {
	switch blockSizeWidth {
	case 8:
		return decodeU8Sparse8[T](bytes)
	case 16:
		return decodeU8Sparse16[T](bytes)
	case 32:
		return decodeU8Sparse32[T](bytes)
	default:
		panic(fmt.Sprintf("unsupported static array type (u8,u%d)", blockSizeWidth))
	}
}

func decodeU8Sparse8[T uint8 | uint16 | uint32 | uint64](bytes []byte) []T {
	var (
		index int
		data  = make([]T, countSparseRows8(1, bytes))
	)
	//
	for i := 0; i < len(bytes); i += 2 {
		// Read ith tuple
		value := bytes[i]
		count := bytes[i+1]
		// Measure it out
		for range count {
			data[index] = T(value)
			index++
		}
	}
	//
	return data
}

func decodeU8Sparse16[T uint8 | uint16 | uint32 | uint64](bytes []byte) []T {
	var (
		index int
		data  = make([]T, countSparseRows16(1, bytes))
	)
	//
	for i := 0; i < len(bytes); i += 3 {
		// Read ith tuple
		value := bytes[i]
		count := binary.BigEndian.Uint16(bytes[i+1 : i+3])
		// Measure it out
		for range count {
			data[index] = T(value)
			index++
		}
	}
	//
	return data
}

func decodeU8Sparse32[T uint8 | uint16 | uint32 | uint64](bytes []byte) []T {
	var (
		index int
		data  = make([]T, countSparseRows32(1, bytes))
	)
	//
	for i := 0; i < len(bytes); i += 5 {
		// Read ith tuple
		value := bytes[i]
		count := binary.BigEndian.Uint32(bytes[i+1 : i+5])
		// Measure it out
		for range count {
			data[index] = T(value)
			index++
		}
	}
	//
	return data
}
