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
package field

import (
	"fmt"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// FrPool captures a pool of field elements which are used to reduce unnecessary
// duplication of elements.
type FrPool[K any] interface {
	// Allocate an item into the pool, returning its index.
	Put(fr.Element) K

	// Lookup a given item in the pool using an index.
	Get(K) fr.Element
}

// ----------------------------------------------------------------------------

// FrBitPool is a pool implementation indexed using a single bit which is backed
// by an array of a fixed size.  This is ideally suited for representing bit
// columns.
type FrBitPool struct{}

// NewFrBitPool constructs a new pool which uses a single bit for indexing.
func NewFrBitPool() FrBitPool {
	return FrBitPool{}
}

// Get looks up the given item in the pool.
func (p FrBitPool) Get(index bool) fr.Element {
	if index {
		return pool16bit[1]
	}
	//
	return pool16bit[0]
}

// Put allocates an item into the pool, returning its index.  Since the pool is
// fixed, then so is the index.
func (p FrBitPool) Put(element fr.Element) bool {
	val := element.Uint64()
	// Sanity checks
	if !element.IsUint64() || val >= 2 {
		panic(fmt.Sprintf("invalid field element for bit pool (%d)", val))
	} else if val == 1 {
		return true
	}
	// Done
	return false
}

// ----------------------------------------------------------------------------

// FrIndexPool is a pool implementation which is backed by an array of a fixed
// size.
type FrIndexPool[K uint8 | uint16] struct{}

// NewFrIndexPool constructs a new pool which uses a given key type for
// indexing.
func NewFrIndexPool[K uint8 | uint16]() FrIndexPool[K] {
	return FrIndexPool[K]{}
}

// Get looks up the given item in the pool.
func (p FrIndexPool[K]) Get(index K) fr.Element {
	return pool16bit[index]
}

// Put allocates an item into the pool, returning its index.  Since the pool is
// fixed, then so is the index.
func (p FrIndexPool[K]) Put(element fr.Element) K {
	val := element.Uint64()
	// Sanity checks
	if !element.IsUint64() || val >= 65536 {
		panic(fmt.Sprintf("invalid field element for bit pool (%d)", val))
	}

	return K(val)
}

// -------------------------------------------------------------------------------

var pool16init sync.Once
var pool16bit []fr.Element

// Initialise the index pool.
func init() {
	// Singleton pattern for initialisation.
	pool16init.Do(func() {
		// Construct empty array
		tmp := make([]fr.Element, 65536)
		// Initialise array
		for i := uint(0); i < 65536; i++ {
			tmp[i] = fr.NewElement(uint64(i))
		}
		// Should not race
		pool16bit = tmp
	})
}
