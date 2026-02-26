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
package memory

// WriteOnceMemory (WOM) represents a form of memory where each cell can be
// written exactly once and, furthermore, cells must be written consecutively
// starting from zero.  Thus, a WOM can be viewed as an output stream (which is
// exactly what they are typically used for).
type WriteOnceMemory[W any] interface {
	// Name returns the name of this WOM
	Name() string
	// Write a given data-tuple to a given address-tuple.  Observe some
	// constraints around this:  firstly, no address can be written more than
	// once; secondly, addresses must be written consecutively starting from 0.
	// If these requirements are not met, then this will panic.
	Write(address []W, value []W)
	// Return the contents of this WOM as a sequence of words, where all rows
	// are simply appended together.
	Contents() []W
}
