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
package data

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/util"
)

// Type provides an abstraction over raw words which, in principle, can be used
// to support richer forms of type (e.g. structs).
type Type interface {
	fmt.Stringer
	// Return the number of bits to represent an element of this type.
	BitWidth() uint
	// Flattern this type into a set of one or more registers, using a given
	// prefix.  For example, a variable "x [2]u8" is flatterned into "x$0 u8"
	// and "x$1 u8", etc.
	Flattern(prefix string, constructor func(name string, bitwidth uint))
}

// Struct represents a (non-recursive) data structure composed of one or more
// named fields, each of which has a declared type.
type Struct struct {
	Fields []util.Pair[string, Type]
}

// Flattern implementation for Type interface
func (p *Struct) Flattern(prefix string, constructor func(name string, bitwidth uint)) {
	panic("todo")
}

// Array represents a fixed-size array of a given type.
type Array struct {
	Width   uint
	Element Type
}

// Flattern implementation for Type interface
func (p *Array) Flattern(prefix string, constructor func(name string, bitwidth uint)) {
	panic("todo")
}

// BitWidth implementation for Type interface
func (p *Array) BitWidth() uint {
	return p.Width * p.Element.BitWidth()
}
