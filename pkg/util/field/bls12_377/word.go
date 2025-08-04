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

// Bit implementation for the Word interface.
func (x Element) Bit(uint) bool {
	panic("todo")
}

// ByteWidth implementation for the Word interface.
func (x Element) ByteWidth() uint {
	switch {
	case x.Element[3] != 0:
		return word.ByteWidth64(x.Element[3])
	case x.Element[2] != 0:
		return word.ByteWidth64(x.Element[2])
	case x.Element[1] != 0:
		return word.ByteWidth64(x.Element[1])
	default:
		return word.ByteWidth64(x.Element[0])
	}
}

// Put implementation for the Word interface.
func (x Element) Put(bytes []byte) []byte {
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
		binary.BigEndian.PutUint64(bytes[24:], x.Element[3])
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

// Set implementation for the Word interface.
func (x Element) Set(bytes []byte) Element {
	x.SetBytes(bytes)
	return x
}

// Equals implementation for the Word interface.
func (x Element) Equals(other Element) bool {
	return x == other
}

// Hash implementation for the Word interface.
func (x Element) Hash() uint64 {
	hash := fnv.New64a()
	// FIXME: could do better here.
	hash.Write(x.Bytes())
	// Done
	return hash.Sum64()
}
