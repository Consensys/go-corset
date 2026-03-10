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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// ResolvedType represents a type which contains only resolved identifiers.
type ResolvedType = Type[symbol.Resolved]

// UnresolvedType represents a type which may contain unresolved identifiers.
type UnresolvedType = Type[symbol.Unresolved]

// Type provides an abstraction over raw words which, in principle, can be used
// to support richer forms of type (e.g. structs).
type Type[S symbol.Symbol[S]] interface {
	// AsUint determines whether or not this is an unsigned int.
	AsUint(Environment[S]) *UnsignedInt[S]
	// AsTuple determines whether or not this is a tuple
	AsTuple(Environment[S]) *Tuple[S]
	// AsUint determines whether or not this is an alias.
	AsAlias(Environment[S]) *Alias[S]
	// String returns a string representation of this type.
	String(Environment[S]) string
}
