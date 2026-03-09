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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// Constant represents a named constant at the source level.
type Constant[S symbol.Symbol[S]] struct {
	name      string
	DataType  data.Type[S]
	ConstExpr expr.Expr[S]
}

// NewConstant creates a new named constant in a given base
func NewConstant[S symbol.Symbol[S]](name string, datatype data.Type[S], constexpr expr.Expr[S]) *Constant[S] {
	return &Constant[S]{name, datatype, constexpr}
}

// Arity implementation for Declaration interface
func (p *Constant[S]) Arity() (nInputs, nOutputs uint) {
	return 0, 0
}

// Name implementation for AssemblyComponent interface
func (p *Constant[I]) Name() string {
	return p.name
}

// Externs implementation for Declaration interface.
func (p *Constant[I]) Externs() []I {
	return nil
}
