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
package decl

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// MemoryKind determines the type of a given memory (i.e. random access, read
// only, etc).
type MemoryKind uint8

const (
	// PUBLIC_STATIC_MEMORY represents a (public) static read-only memory.  That
	// is a ROM which never changes across all executions of a given machine.
	PUBLIC_STATIC_MEMORY = 0
	// PRIVATE_STATIC_MEMORY represents a (private) static read-only memory.  That
	// is a ROM which never changes across all executions of a given machine.
	PRIVATE_STATIC_MEMORY = 1
	// PUBLIC_READ_ONLY_MEMORY represents a (public) read-only memory which can
	// change between different executions of a given machine.
	PUBLIC_READ_ONLY_MEMORY = 2
	// PRIVATE_READ_ONLY_MEMORY represents a (private) read-only memory which
	// can change between different executions of a given machine.
	PRIVATE_READ_ONLY_MEMORY = 3
	// PUBLIC_WRITE_ONCE_MEMORY represents a (public) write-only memory which can only be
	// written once.
	PUBLIC_WRITE_ONCE_MEMORY = 4
	// PRIVATE_WRITE_ONCE_MEMORY represents a (private) write-only memory which
	// can only be written once.
	PRIVATE_WRITE_ONCE_MEMORY = 5
	// RANDOM_ACCESS_MEMORY represents the ubiquitous form of memory which
	// supports arbitrary reads / writes.  Observe that RAM is always private.
	RANDOM_ACCESS_MEMORY = 6
)

// Memory represents a declaration of some form of memory, such as random
// access, read only, etc.
type Memory[E any] struct {
	// Name given to this memory variable
	name string
	// Kind of memory (i.e. read-only, random access, etc)
	Kind MemoryKind
	// Address bus for memory (where, for random access, the first line always
	// denotes the index type used).
	Address []variable.Descriptor
	// Data bus for memory.
	Data []variable.Descriptor
	// Contents (for static memory only)
	Contents []big.Int
}

// NewMemory constructs a new memory.
func NewMemory[E any](name string, kind MemoryKind, address []variable.Descriptor, data []variable.Descriptor,
	contents []big.Int) *Memory[E] {
	// sanity checks
	if contents != nil && kind != PUBLIC_STATIC_MEMORY && kind != PRIVATE_STATIC_MEMORY {
		panic("invalid non-static memory")
	} else if contents == nil && (kind == PUBLIC_STATIC_MEMORY || kind == PRIVATE_STATIC_MEMORY) {
		panic("invalid static memory")
	}
	//
	return &Memory[E]{name: name, Kind: kind, Address: address, Data: data, Contents: contents}
}

// NewRandomAccessMemory constructs a new random access memory.
func NewRandomAccessMemory[E any](name string, address []variable.Descriptor, data []variable.Descriptor) *Memory[E] {
	return &Memory[E]{name: name, Kind: RANDOM_ACCESS_MEMORY, Address: address, Data: data}
}

// NewReadOnlyMemory constructs a new read-only access memory.
func NewReadOnlyMemory[E any](public bool, name string, address []variable.Descriptor, data []variable.Descriptor,
) *Memory[E] {
	if public {
		return &Memory[E]{name: name, Kind: PUBLIC_READ_ONLY_MEMORY, Address: address, Data: data}
	}
	//
	return &Memory[E]{name: name, Kind: PRIVATE_READ_ONLY_MEMORY, Address: address, Data: data}
}

// NewWriteOnceMemory constructs a new write-once memory.
func NewWriteOnceMemory[E any](public bool, name string, address []variable.Descriptor, data []variable.Descriptor,
) *Memory[E] {
	if public {
		return &Memory[E]{name: name, Kind: PUBLIC_WRITE_ONCE_MEMORY, Address: address, Data: data}
	}
	//
	return &Memory[E]{name: name, Kind: PRIVATE_WRITE_ONCE_MEMORY, Address: address, Data: data}
}

// NewStaticMemory constructs a new static memory.
func NewStaticMemory[E any](public bool, name string, address []variable.Descriptor, data []variable.Descriptor,
	contents []big.Int) *Memory[E] {
	//
	if public {
		return &Memory[E]{name: name, Kind: PUBLIC_STATIC_MEMORY, Address: address, Data: data, Contents: contents}
	}
	//
	return &Memory[E]{name: name, Kind: PRIVATE_STATIC_MEMORY, Address: address, Data: data, Contents: contents}
}

// Name implementation for Declaration interface
func (p *Memory[E]) Name() string {
	return p.name
}

// Externs implementation for Declaration interface
func (p *Memory[E]) Externs() []E {
	return nil
}
