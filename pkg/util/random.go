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
package util

import "math/rand/v2"

// GenerateRandomUints generates n random unsigned integers in the range 0..m.
func GenerateRandomUints(n, m uint) []uint {
	items := make([]uint, n)

	for i := uint(0); i < n; i++ {
		items[i] = rand.UintN(m)
	}

	return items
}

// GenerateRandomInts generates n random unsigned integers in the range -m..m.
func GenerateRandomInts(n uint, m int) []int {
	items := make([]int, n)

	for i := uint(0); i < n; i++ {
		items[i] = rand.IntN(2*m) - m
	}

	return items
}

// GenerateRandomElements generates n elements selected at random from the given array.
func GenerateRandomElements[E any](n uint, elems []E) []E {
	items := make([]E, n)
	m := uint(len(elems))

	for i := uint(0); i < n; i++ {
		index := rand.UintN(m)
		items[i] = elems[index]
	}

	return items
}
