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
)

// Bit implementation for the Word interface.
func (x Element) Bit(uint) bool {
	panic("todo")
}

// BitWidth implementation for the Word interface.
func (x Element) BitWidth() uint {
	return 252
}

// Put implementation for the Word interface.
func (x Element) Put(bytes []byte) []byte {
	if len(bytes) < 32 {
		return x.Bytes()
	}
	// Copy over each element without allocating new array.
	binary.BigEndian.PutUint64(bytes, x.Element[0])
	binary.BigEndian.PutUint64(bytes[8:], x.Element[1])
	binary.BigEndian.PutUint64(bytes[16:], x.Element[2])
	binary.BigEndian.PutUint64(bytes[24:], x.Element[3])
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
