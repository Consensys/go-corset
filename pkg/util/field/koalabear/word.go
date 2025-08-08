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
package koalabear

import "hash/fnv"

// SetUint64 implementation for Element interface
func (x Element) SetUint64(val uint64) Element {
	var elem Element
	//
	return elem.AddUint32(uint32(val))
}

// Bit implementation for the Word interface.
func (x Element) Bit(uint) bool {
	panic("todo")
}

// ByteWidth implementation for the Word interface.
func (x Element) ByteWidth() uint {
	panic("todo")
}

// PutRawBytes implementation for the Word interface.
func (x Element) PutRawBytes([]byte) []byte {
	panic("todo")
}

// RawBytes implementation for word.Word interface.
func (x Element) RawBytes() []byte {
	panic("todo")
}

// SetRawBytes implementation for the Word interface.
func (x Element) SetRawBytes([]byte) Element {
	panic("todo")
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
