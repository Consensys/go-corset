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
package word

// SmallPool is a pool implementation for handling small domains (i.e. where all
// elements can be enumerated).
type SmallPool[K uint8 | uint16, W Word[W]] struct {
	index []W
}

// NewBytePool constructs a new pool which uses a given key type for
// indexing.
func NewBytePool[W Word[W]]() SmallPool[uint8, W] {
	index := make([]W, 256)
	//
	for i := range uint64(256) {
		var ith W
		index[i] = ith.SetUint64(i)
	}
	//
	return SmallPool[uint8, W]{index}
}

// NewWordPool constructs a new pool which uses a given key type for
// indexing.
func NewWordPool[W Word[W]]() SmallPool[uint16, W] {
	index := make([]W, 65536)
	//
	for i := range uint64(65536) {
		var ith W
		index[i] = ith.SetUint64(i)
	}
	//
	return SmallPool[uint16, W]{index}
}

// Get implementation for Pool interface.
func (p SmallPool[K, F]) Get(index K) F {
	return p.index[index]
}

// Put implementation for Pool interface.
func (p SmallPool[K, F]) Put(element F) K {
	return K(element.Uint64())
}

// Clone implementation for Pool interface.
func (p SmallPool[K, F]) Clone() Pool[K, F] {
	return p
}
