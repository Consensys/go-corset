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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// ResolvedType represents a type whose contents uses only external
// identifiers which are resolved. As such, it should not be possible that such
// a declaration refers to unknown (or otherwise incorrect) external components.
type ResolvedType = Type[symbol.Resolved]

// UnresolvedType represents a type whose contents may contain string
// identifiers for external (i.e. unlinked) components.  As such, its possible
// that such an expression may fail with an error at link time due to an
// unresolvable reference to an external component (e.g. function, RAM, ROM,
// etc).
type UnresolvedType = Type[symbol.Unresolved]

// Type represents a type alias at the source level.
type Type[S symbol.Symbol[S]] struct {
	name     string
	DataType data.Type[S]
}

// NewType creates a new named type over arbitrary symbol identifiers.
func NewType[S symbol.Symbol[S]](name string, datatype data.Type[S]) *Type[S] {
	return &Type[S]{name, datatype}
}

// Arity implementation for Declaration interface
func (p *Type[S]) Arity() (nInputs, nOutputs uint) {
	return 0, 0
}

// Name implementation for AssemblyComponent interface
func (p *Type[I]) Name() string {
	return p.name
}

// Externs implementation for Declaration interface.
func (p *Type[I]) Externs() []I {
	return nil
}
