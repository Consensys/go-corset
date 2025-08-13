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

import "hash/fnv"

// Equals implementation for hash.Hasher interface
func (x Element) Equals(o Element) bool {
	return x.Element == o.Element
}

// Hash implementation for hash.Hasher interface
func (x Element) Hash() uint64 {
	var (
		bytes = x.Element.Bytes()
		hash  = fnv.New64a()
	)
	// Write hash bytes
	hash.Write(bytes[:])
	// Done
	return hash.Sum64()
}
