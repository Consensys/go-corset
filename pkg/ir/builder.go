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
package ir

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
)

type RegisterId struct {
	Module   uint
	Register uint
}

type SchemaBuilder[C schema.Constraint, T Term[T]] struct {
	// Modmap maps modules to registers
	modmap map[string]uint
	// Externs represent modules which have already been constructed.  These
	// will be given the lower module identifiers, since they are already
	// packaged and, hence, we must avoid breaking thein linkage.
	externs []schema.Module
	// Modules being constructed
	modules []ModuleBuilder[C, T]
}

func NewSchemaBuilder[C schema.Constraint, T Term[T]](externs ...schema.Module) SchemaBuilder[C, T] {
	modmap := make(map[string]uint, 0)
	// Initialise module map
	for i, m := range externs {
		// Quick sanity check
		if _, ok := modmap[m.Name()]; ok {
			panic(fmt.Sprintf("duplicate module \"%s\" detected", m.Name()))
		}
		//
		modmap[m.Name()] = uint(i)
	}
	//
	return SchemaBuilder[C, T]{modmap, externs, nil}
}

// NewModule constructs a new, empty module and returns its unique module
// identifier.
func (p *SchemaBuilder[C, T]) NewModule(name string) uint {
	var mid = uint(len(p.externs) + len(p.modules))
	// Sanity check this module is not already declared
	if _, ok := p.modmap[name]; ok {
		panic(fmt.Sprintf("module \"%s\" already declared", name))
	}
	//
	p.modules = append(p.modules, NewModuleBuilder[C, T](name))
	p.modmap[name] = mid
	//
	return mid
}

func (p *SchemaBuilder[C, T]) Module(mid uint) *ModuleBuilder[C, T] {
	var n uint = uint(len(p.externs))
	// Sanity check
	if mid < n {
		panic("no builder for external module")
	}
	//
	return &p.modules[mid-n]
}

func (p *SchemaBuilder[C, T]) ModuleOf(name string) *ModuleBuilder[C, T] {
	return p.Module(p.modmap[name])
}

func (p *SchemaBuilder[C, T]) Modules() []schema.Table[C] {
	modules := make([]schema.Table[C], len(p.modules))
	//
	for i, m := range p.modules {
		modules[i] = m.build()
	}
	//
	return modules
}

type ModuleBuilder[C schema.Constraint, T Term[T]] struct {
	// Name of the module being constructed
	name string
	// Maps register names (including aliases) to the register number.
	regmap map[string]uint
	// Registers declared for this module
	registers []schema.Column
	// Constraints for this module
	constraints []C
}

// NewModuleBuilder constructs a new builder for a module with the given name.
func NewModuleBuilder[C schema.Constraint, T Term[T]](name string) ModuleBuilder[C, T] {
	regmap := make(map[string]uint, 0)
	return ModuleBuilder[C, T]{name, regmap, nil, nil}
}

// ColumnAccessOf returns a column accessor for the column with the given name.
func (p *ModuleBuilder[C, T]) ColumnAccessOf(name string) *ColumnAccess[T] {
	panic("todo")
}

// NewColumn declares a new column within the module being built.  This will
// panic if a column of the same name already exists.
func (p *ModuleBuilder[C, T]) NewColumn(column schema.Column) uint {
	// Determine identifier
	id := uint(len(p.registers))
	// Sanity check
	if _, ok := p.regmap[column.Name]; ok {
		panic(fmt.Sprintf("column \"%s\" already declared", column.Name))
	}
	//
	p.registers = append(p.registers, column)
	p.regmap[column.Name] = id
	//
	return id
}

func (p *ModuleBuilder[C, T]) AddConstraint(constraint C) {
	panic("todo")
}

// Build constructs a table module from this module builder.
func (p *ModuleBuilder[C, T]) build() schema.Table[C] {
	return schema.NewTable(p.name, p.registers, p.constraints)
}
