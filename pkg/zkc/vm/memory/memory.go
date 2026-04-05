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

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Memory represents (in many ways) the simplest form of memory
// which can be read or written without restrictions.  Initially, all locations
// of a RAM can be considered to hold zero.  Thus, reading a location which has
// not yet been written will return zero; otherwise, it will return the last
// value written.
type Memory[W word.Word[W]] interface {
	// Name returns the name of this RAM
	Name() string
	Geometry() Geometry[W]
	// Initialise this memory with the given contents.  This will overwrite any
	// existing contents.
	Initialise(contents []W)
	// Read a given data-tuple from a given address-tuple.
	Read(address []W) []W
	// Write a given data-tuple to a given address-tuple, overwriting the
	// previous value stored at that address.
	Write(address []W, value []W)
	// Read a given data-tuple from a given address-tuple.
	FrameRead(frame []W, address []register.Id, data []register.Id) error
	// Read a given data-tuple from a given address-tuple.
	FrameWrite(frame []W, address []register.Id, data []register.Id) error
	// Return the contents of this memory as a sequence of words, where all rows
	// are simply appended together.
	Contents() []W
}
