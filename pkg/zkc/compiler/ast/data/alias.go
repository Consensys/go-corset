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

// Alias captures the alias of a language type.
type Alias[I symbol.Symbol[I]] struct {
	Name     string
	bitwidth uint
}

// NewAlias constructs an alias for a given type.
func NewAlias[I symbol.Symbol[I]](name string, bitwidth uint) *Alias[I] {
	return &Alias[I]{name, bitwidth}
}

// BitWidth implementation for Type interface
func (p *Alias[I]) BitWidth() uint {
	return p.bitwidth
}

// Flattern implementation for Type interface
func (p *Alias[I]) Flattern(prefix string, constructor func(name string, bitwidth uint)) {
	constructor(prefix, p.bitwidth)
}

func (p *Alias[I]) String() string {
	return fmt.Sprintf("%s", p.Name)
}
