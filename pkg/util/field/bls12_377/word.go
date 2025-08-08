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
package bls12_377

import (
	"encoding/binary"
	"hash/fnv"

	"github.com/consensys/go-corset/pkg/util/word"
)

// Bit implementation for word.Word interface.
func (x Element) Bit(uint) bool {
	panic("todo")
}

// ByteWidth implementation for word.Word interface.
func (x Element) ByteWidth() uint {
	switch {
	case x.Element[3] != 0:
		return 24 + word.ByteWidth64(x.Element[3])
	case x.Element[2] != 0:
		return 16 + word.ByteWidth64(x.Element[2])
	case x.Element[1] != 0:
		return 8 + word.ByteWidth64(x.Element[1])
	default:
		return word.ByteWidth64(x.Element[0])
	}
}

// RawBytes implementation for word.Word interface.
func (x Element) RawBytes() []byte {
	panic("todo")
}

// PutRawBytes implementation for word.Word interface.
func (x Element) PutRawBytes(bytes []byte) []byte {
	var width = x.ByteWidth()
	// Sanity check enough space
	if uint(len(bytes)) < width {
		bytes = make([]byte, width)
	}
	// Copy over each element without allocating new array.  Do this with as few
	// branches as possible.
	switch {
	case x.Element[3] != 0:
		binary.BigEndian.PutUint64(bytes, x.Element[0])
		binary.BigEndian.PutUint64(bytes[8:], x.Element[1])
		binary.BigEndian.PutUint64(bytes[16:], x.Element[2])
		putRawBytes64(bytes[24:], x.Element[3])
	case x.Element[2] != 0:
		binary.BigEndian.PutUint64(bytes, x.Element[0])
		binary.BigEndian.PutUint64(bytes[8:], x.Element[1])
		binary.BigEndian.PutUint64(bytes[16:], x.Element[2])
	case x.Element[1] != 0:
		binary.BigEndian.PutUint64(bytes, x.Element[0])
		binary.BigEndian.PutUint64(bytes[8:], x.Element[1])
	case x.Element[0] != 0:
		binary.BigEndian.PutUint64(bytes, x.Element[0])
	}
	//
	return bytes
}

// SetRawBytes implementation for word.Word interface.
func (x Element) SetRawBytes(bytes []byte) Element {
	var y Element
	//
	switch {
	case len(bytes) >= 24:
		y.Element[3] = setRawBytes64(bytes[24:])
		y.Element[2] = binary.BigEndian.Uint64(bytes[16:24])
		y.Element[1] = binary.BigEndian.Uint64(bytes[8:16])
		y.Element[0] = binary.BigEndian.Uint64(bytes[0:8])
	case len(bytes) >= 16:
		y.Element[2] = setRawBytes64(bytes[16:])
		y.Element[1] = binary.BigEndian.Uint64(bytes[8:16])
		y.Element[0] = binary.BigEndian.Uint64(bytes[0:8])
	case len(bytes) >= 8:
		y.Element[1] = setRawBytes64(bytes[8:])
		y.Element[0] = binary.BigEndian.Uint64(bytes[0:8])
	default:
		y.Element[0] = setRawBytes64(bytes)
	}
	//
	return y
}

// Equals implementation for word.Word interface.
func (x Element) Equals(other Element) bool {
	return x == other
}

// Hash implementation for word.Word interface.
func (x Element) Hash() uint64 {
	hash := fnv.New64a()
	// FIXME: could do better here.
	hash.Write(x.Bytes())
	// Done
	return hash.Sum64()
}

func setRawBytes64(bytes []byte) uint64 {
	var val uint64
	//
	switch len(bytes) {
	case 0:
		val = 0
	case 1:
		val = uint64(bytes[0])
	case 2:
		val = uint64(bytes[1])
		val += uint64(bytes[0]) << 8
	case 3:
		val = uint64(bytes[2])
		val += uint64(bytes[1]) << 8
		val += uint64(bytes[0]) << 16
	case 4:
		val = uint64(bytes[3])
		val += uint64(bytes[2]) << 8
		val += uint64(bytes[1]) << 16
		val += uint64(bytes[0]) << 24
	case 5:
		val = uint64(bytes[4])
		val += uint64(bytes[3]) << 8
		val += uint64(bytes[2]) << 16
		val += uint64(bytes[1]) << 24
		val += uint64(bytes[0]) << 32
	case 6:
		val = uint64(bytes[5])
		val += uint64(bytes[4]) << 8
		val += uint64(bytes[3]) << 16
		val += uint64(bytes[2]) << 24
		val += uint64(bytes[1]) << 32
		val += uint64(bytes[0]) << 40
	case 7:
		val = uint64(bytes[6])
		val += uint64(bytes[5]) << 8
		val += uint64(bytes[4]) << 16
		val += uint64(bytes[3]) << 24
		val += uint64(bytes[2]) << 32
		val += uint64(bytes[1]) << 40
		val += uint64(bytes[0]) << 48
	default:
		val = uint64(bytes[7])
		val += uint64(bytes[6]) << 8
		val += uint64(bytes[5]) << 16
		val += uint64(bytes[4]) << 24
		val += uint64(bytes[3]) << 32
		val += uint64(bytes[2]) << 40
		val += uint64(bytes[1]) << 48
		val += uint64(bytes[0]) << 56
	}
	//
	return val
}

func putRawBytes64(bytes []byte, val uint64) {
	//
	switch len(bytes) {
	case 0:

	case 1:
		bytes[0] = uint8(val)
	case 2:
		bytes[0] = uint8(val >> 8)
		bytes[1] = uint8(val)
	case 3:
		bytes[0] = uint8(val >> 16)
		bytes[1] = uint8(val >> 8)
		bytes[2] = uint8(val)
	case 4:
		bytes[0] = uint8(val >> 24)
		bytes[1] = uint8(val >> 16)
		bytes[2] = uint8(val >> 8)
		bytes[3] = uint8(val)
	case 5:
		bytes[0] = uint8(val >> 32)
		bytes[1] = uint8(val >> 24)
		bytes[2] = uint8(val >> 16)
		bytes[3] = uint8(val >> 8)
		bytes[4] = uint8(val)
	case 6:
		bytes[0] = uint8(val >> 40)
		bytes[1] = uint8(val >> 32)
		bytes[2] = uint8(val >> 24)
		bytes[3] = uint8(val >> 16)
		bytes[4] = uint8(val >> 8)
		bytes[5] = uint8(val)
	case 7:
		bytes[0] = uint8(val >> 48)
		bytes[1] = uint8(val >> 40)
		bytes[2] = uint8(val >> 32)
		bytes[3] = uint8(val >> 24)
		bytes[4] = uint8(val >> 16)
		bytes[5] = uint8(val >> 8)
		bytes[6] = uint8(val)
	default:
		bytes[0] = uint8(val >> 56)
		bytes[1] = uint8(val >> 48)
		bytes[2] = uint8(val >> 40)
		bytes[3] = uint8(val >> 32)
		bytes[4] = uint8(val >> 24)
		bytes[5] = uint8(val >> 16)
		bytes[6] = uint8(val >> 8)
		bytes[7] = uint8(val)
	}
}
