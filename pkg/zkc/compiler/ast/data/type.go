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

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

type ResolvedType = Type[symbol.Resolved]

type UnresolvedType = Type[symbol.Unresolved]

// Type provides an abstraction over raw words which, in principle, can be used
// to support richer forms of type (e.g. structs).
type Type[S symbol.Symbol[S]] interface {
	fmt.Stringer
	// AsUint determines whether or not this is an unsigned int.
	AsUint() *UnsignedInt[S]
	// Return the number of bits to represent an element of this type.
	BitWidth(Environment[S]) uint
	// Flattern this type into a set of one or more registers, using a given
	// prefix.  For example, a variable "x [2]u8" is flatterned into "x$0 u8"
	// and "x$1 u8", etc.
	Flattern(prefix string, env Environment[S], constructor func(name string, bitwidth uint))
}
