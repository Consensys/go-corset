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

import (
	"github.com/consensys/go-corset/pkg/util/collection/array"
)

// Word abstracts a sequence of n bits.
type Word[T any] interface {
	// Return bitwidth of this word
	BitWidth() uint
	// Write contents of this word into given byte array.
	Put([]byte)
	// Initialise this word from a set of bytes.
	Set([]byte) T
}

// NewArray constructs a new word array with a given capacity.
func NewArray[T Word[T]](height uint, bitwidth uint) array.Array[T] {
	switch {
	case bitwidth == 1:
		panic("implement bit array")
	case bitwidth < 64:
		return NewStaticArray[T](height, bitwidth)
	default:
		panic("implement indexed array")
	}
}
