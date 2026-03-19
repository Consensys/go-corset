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
package parser

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Environment captures useful information used during the assembling process.
type Environment struct {
	effects []*symbol.Unresolved
	// Variables identifies set of declared variables.
	variables []VariableDescriptor
}

// DeclareEffect declares a new effect.  If an effect with the same name
// already exists, this panics.
func (p *Environment) DeclareEffect(effect *symbol.Unresolved) {
	//
	if p.IsDeclared(effect.Name) {
		panic(fmt.Sprintf("effect %s already declared", effect.Name))
	}
	//
	p.effects = append(p.effects, effect)
}

// DeclareVariable declares a new register with the given name and bitwidth.  If
// a register with the same name already exists, this panics.
func (p *Environment) DeclareVariable(kind variable.Kind, name string, datatype Type) {
	//
	if p.IsDeclared(name) {
		panic(fmt.Sprintf("variable %s already declared", name))
	}
	//
	p.variables = append(p.variables, variable.New(kind, name, datatype))
}

// IsDeclared checks whether or not a given name is already declared (either as
// an effect or a variable).
func (p *Environment) IsDeclared(name string) bool {
	// check effects
	for _, effect := range p.effects {
		if effect.Name == name {
			return true
		}
	}
	// check local variables
	return p.IsDeclaredVariable(name)
}

// IsDeclaredVariable checks whether or not a given name is already declared as
// a variable.
func (p *Environment) IsDeclaredVariable(name string) bool {
	// check local variables
	for _, variable := range p.variables {
		if variable.Name == name {
			return true
		}
	}
	//
	return false
}

// LookupVariable looks up the index for a given register.
func (p *Environment) LookupVariable(name string) variable.Id {
	for i, variable := range p.variables {
		if variable.Name == name {
			return uint(i)
		}
	}
	//
	panic(fmt.Sprintf("unknown register %s", name))
}
