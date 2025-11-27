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

// DecodeU1 decodes an array of 1bit values represented in a "dense encoding".
// That is, where each value is stored consecutively with 8 values being packed
// into a single byte.  The number of "unused values" is stored in an additional
// final byte.
func DecodeU1(bytes []byte) ([]byte, uint) {
	var (
		n      = uint(len(bytes) - 1)
		unused = uint(bytes[n])
		data   = bytes[:n]
		height = uint(len(data)*8) - unused
	)
	//
	return data, height
}

// DecodeU1Dense decodes an array of 1bit values represented in a "dense
// encoding".  That is, where each value is stored consecutively with 8 values
// being packed into a single byte.  The number of "unused values" is stored in
// an additional final byte.
func DecodeU1Dense[T uint8 | uint16 | uint32](bytes []byte) []T {
	var (
		n      = uint(len(bytes) - 1)
		unused = uint(bytes[n])
		height = (n * 8) - unused
		data   = make([]T, height)
	)
	//
	for i := range height {
		var (
			ith    = bytes[i/8]
			offset = i % 8
		)
		//
		data[i] = T((ith >> offset) & 0x1)
	}
	//
	return data
}

// DecodeU2Dense decodes an array of 2bit values represented in a "dense
// encoding".  That is, where value is stored consecutively with 4 values being
// packed into a single byte.  The number of "unused values" is stored in an
// additional final byte.
func DecodeU2Dense[T uint8 | uint16 | uint32](bytes []byte) []T {
	var (
		n      = uint(len(bytes) - 1)
		unused = uint(bytes[n])
		height = (n * 4) - unused
		data   = make([]T, height)
	)
	//
	for i := range height {
		var (
			ith    = bytes[i/4]
			offset = 2 * (i % 4)
		)
		//
		data[i] = T((ith >> offset) & 0x3)
	}
	//
	return data
}

// DecodeU4Dense decodes an array of 2bit values represented in a "dense
// encoding".  That is, where value is stored consecutively with 2 values being
// packed into a single byte.  The number of "unused values" is stored in an
// additional final byte.
func DecodeU4Dense[T uint8 | uint16 | uint32](bytes []byte) []T {
	var (
		n      = uint(len(bytes) - 1)
		unused = uint(bytes[n])
		height = (n * 2) - unused
		data   = make([]T, height)
	)
	//
	for i := range height {
		var (
			ith    = bytes[i/2]
			offset = 4 * (i % 2)
		)
		//
		data[i] = T((ith >> offset) & 0xf)
	}
	//
	return data
}

// DecodeU8Dense decode an array of bytes represented in a "dense encoding".
// That is, where value is stored consecutively.
func DecodeU8Dense[T uint8 | uint16 | uint32 | uint64](bytes []byte) []T {
	var data = make([]T, len(bytes))
	//
	for i := range len(bytes) {
		data[i] = T(bytes[i])
	}
	//
	return data
}

// DecodeU16Dense decode an array of 16bit values represented in a "dense encoding".
// That is, where value is stored consecutively.
func DecodeU16Dense[T uint16 | uint32 | uint64](bytes []byte) []T {
	var (
		n    = uint(len(bytes) / 2)
		data = make([]T, n)
	)
	//
	for i := range n {
		var (
			offset = i * 2
			b1     = uint16(bytes[offset])
			b0     = uint16(bytes[offset+1])
		)
		// Assign ith element
		data[i] = T((b1 << 8) | b0)
	}
	//
	return data
}

// DecodeU32Dense decode an array of 32bit values represented in a "dense encoding".
// That is, where value is stored consecutively.
func DecodeU32Dense[T uint32 | uint64](bytes []byte) []T {
	var (
		n    = uint(len(bytes) / 4)
		data = make([]T, n)
	)
	//
	for i := range n {
		var (
			offset = i * 4
			b3     = uint32(bytes[offset])
			b2     = uint32(bytes[offset+1])
			b1     = uint32(bytes[offset+2])
			b0     = uint32(bytes[offset+3])
		)
		// Assign ith element
		data[i] = T((b3 << 24) | (b2 << 16) | (b1 << 8) | b0)
	}
	//
	return data
}
