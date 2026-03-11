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

// ResolvedAlias represents an alias which contains only resolved identifiers.
type ResolvedAlias = Alias[symbol.Resolved]

// UnresolvedAlias represents an alias which contains only unresolved identifiers.
type UnresolvedAlias = Alias[symbol.Unresolved]

// Alias captures the alias of a language type.
// Ref points to the symbol for the type-alias declaration if resolved.
type Alias[I symbol.Symbol[I]] struct {
	Name string
	Ref  *I
}

// NewAlias constructs an alias for a given Type.
func NewAlias[I symbol.Symbol[I]](name string, ref *I) *Alias[I] {
	return &Alias[I]{ name, ref}
}

// AsUint implementation for Type interface
func (p *Alias[S]) AsUint(Environment[S]) *UnsignedInt[S] {
	return nil
}

// AsTuple implementation for Type interface
func (p *Alias[S]) AsTuple(Environment[S]) *Tuple[S] {
	return nil
}

// AsAlias implementation for Type interface
func (p *Alias[S]) AsAlias(Environment[S]) *Alias[S] {
	return p
}

func (p *Alias[S]) String(Environment[S]) string {
	return p.Name
}

// Resolve returns the type that this alias refers to in the given environment.
func (p *Alias[S]) Resolve(env Environment[S]) Type[S] {
	if p.Ref == nil {
		panic("unresolved type alias")
	}
	return env.TypeOf(*p.Ref)
}
