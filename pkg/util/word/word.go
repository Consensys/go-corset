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
	"fmt"

	"github.com/consensys/go-corset/pkg/util/collection/hash"
)

// Word abstracts a sequence of n bits.
type Word[T any] interface {
	fmt.Stringer
	hash.Hasher[T]
	// Returns the raw bytes of this word.  Observe that, if the word is encoded
	// (e.g. in Montgomerry form), then the *encoded* bytes are returned.
	Bytes() []byte
	// Check whether this word is zero, or not.
	IsZero() bool
	// Initialise this word from a set of raw bytes.
	SetBytes([]byte) T
	// Set this word to a uint64 value
	SetUint64(uint64) T
	// Returns value of word as an unsigned integer (truncated for 64bits).
	Uint64() uint64
}

// DynamicWord is a word which has a dynamically sized representation, rather
// than a fixed-size representation.  In particular, the dynamic word
// representing zero is always the empty byte array.
type DynamicWord[T any] interface {
	Word[T]
	// Return minimal number of bytes required to store this word.  This can be
	// defined as the length of bytes of this word, with all leading zero bytes
	// removed.  For example, 0x1010 has a length of 2, 0x0010 has a length of 1
	// whilst 0x0000 has a byte length of 0.  Observe that, if the word is
	// encoded (e.g. in Montgomerry form), then this is the length of the
	// encoded bytes.
	ByteWidth() uint
	// Write contents of this word into given byte array.  If the given byte
	// array is not big enough, a new array is allocated and returned.
	PutBytes([]byte) []byte
}

// Pool provides an abstraction for referring to large words by a smaller index
// value.  The pool stores the actual word data, and provides fast access via an
// index.  This makes sense when we have a relatively small number of values
// which can be referred to many times over.
type Pool[K any, T any] interface {
	// Lookup a given word in the pool using an index.
	Get(K) T
	// Allocate word into pool, returning its index.
	Put(T) K
}

// SharedPool represents a pool which can be safely shared amongst threads.
type SharedPool[K any, T any, P any] interface {
	Pool[K, T]
	// Localise this pool
	Localise() P
}

// Uint64 constructs a word from a given uint64 value.
func Uint64[W Word[W]](value uint64) W {
	var word W
	// Easy as
	return word.SetUint64(value)
}
