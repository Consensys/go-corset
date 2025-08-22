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
package gf8209

const (
	offset64 uint64 = 14695981039346656037
	prime64  uint64 = 1099511628211
)

// Equals implementation for hash.Hasher interface
func (x Element) Equals(o Element) bool {
	return x == o
}

// Hash implementation for hash.Hasher interface
func (x Element) Hash() uint64 {
	// FNV1a hash implementation (unrolled)
	hash := offset64
	//
	return (hash ^ uint64(x[0])) * prime64
}

// SetBytes implementation for word.Word interface.
func (x Element) SetBytes(b []byte) Element {
	var y Element
	//
	y.AddBytes(b)
	//
	return y
}

// SetUint64 implementation for word.Word interface
func (x Element) SetUint64(val uint64) Element {
	var elem Element
	//
	return elem.AddUint32(uint32(val))
}

// Uint64 implementation for word.Word interface.
func (x Element) Uint64() uint64 {
	return uint64(x.ToUint32())
}
