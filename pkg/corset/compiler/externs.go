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
package compiler

import (
	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/file"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// DeclareExterns adds externally defined symbols to the given scope in such a
// way that they can then be resolved against.
func DeclareExterns[M schema.RegisterMap](scope *ModuleScope, externs ...M) {
	for _, e := range externs {
		path := file.NewAbsolutePath(e.Name())
		// Declare external module
		scope.Declare(e.Name(), util.None[string]())
		// Define external symbol
		for _, r := range e.Registers() {
			scope.Define(NewExternSymbolDefinition(path, r))
		}
	}
}

// ExternSymbolDefinition provides an implementation of SymbolDefinition
// designed specifically for externally defined symbols.
type ExternSymbolDefinition struct {
	binding ast.ColumnBinding
}

// NewExternSymbolDefinition creates a new external symbol definition for a
// given register in a given module.
func NewExternSymbolDefinition(path file.Path, reg schema.Register) *ExternSymbolDefinition {
	var kind uint8 = ast.NOT_COMPUTED
	//
	if reg.IsComputed() {
		kind = ast.COMPUTED
	}
	//
	return &ExternSymbolDefinition{
		binding: ast.ColumnBinding{
			ColumnContext: path,
			Path:          *path.Extend(reg.Name),
			DataType:      ast.NewUintType(reg.Width),
			MustProve:     true,
			Multiplier:    1,
			Kind:          kind,
			Display:       "hex", // default
		},
	}
}

// Arity implementation for SymbolDefinition interface.
func (p *ExternSymbolDefinition) Arity() util.Option[uint] {
	return util.None[uint]()
}

// Binding implementation for SymbolDefinition interface.
func (p *ExternSymbolDefinition) Binding() ast.Binding {
	return &p.binding
}

// Name implementation for SymbolDefinition interface.
func (p *ExternSymbolDefinition) Name() string {
	return p.binding.Path.Tail()
}

// Path implementation for SymbolDefinition interface.
func (p *ExternSymbolDefinition) Path() *file.Path {
	return &p.binding.Path
}

// Lisp returns a lisp view of this
func (p *ExternSymbolDefinition) Lisp() sexp.SExp {
	panic("unreachable")
}
