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
package pool

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
