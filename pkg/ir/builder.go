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

// SchemaBuilder is a mechanism for constructing mixed schemas which attempts to
// simplify the problem of mapping source-level names to e.g. module-specific
// register indexes.
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

// NewSchemaBuilder constructs a new schema builder with a given number of
// externally defined modules.  Such modules are allocated module indices first.
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

// Module returns the builder for the given module based on its index.
func (p *SchemaBuilder[C, T]) Module(mid uint) *ModuleBuilder[C, T] {
	var n uint = uint(len(p.externs))
	// Sanity check
	if mid < n {
		panic("no builder for external module")
	}
	//
	return &p.modules[mid-n]
}

// ModuleOf returns the builder for the given module based on its name.
func (p *SchemaBuilder[C, T]) ModuleOf(name string) *ModuleBuilder[C, T] {
	return p.Module(p.modmap[name])
}

// Build returns an array of tables constructed by this builder.
func (p *SchemaBuilder[C, T]) Build() []schema.Table[C] {
	modules := make([]schema.Table[C], len(p.modules))
	//
	for i, m := range p.modules {
		modules[i] = m.buildTable()
	}
	//
	return modules
}

// ModuleBuilder provides a mechanism to ease the construction of modules for
// use in schemas.  For example, it maintains a mapping from register names to
// their relevant indices.  It also provides a mechanism for constructing a
// register access based on the register name, etc.
type ModuleBuilder[C schema.Constraint, T Term[T]] struct {
	// Name of the module being constructed
	name string
	// Maps register names (including aliases) to the register number.
	regmap map[string]uint
	// Registers declared for this module
	registers []schema.Register
	// Constraints for this module
	constraints []C
}

// NewModuleBuilder constructs a new builder for a module with the given name.
func NewModuleBuilder[C schema.Constraint, T Term[T]](name string) ModuleBuilder[C, T] {
	regmap := make(map[string]uint, 0)
	return ModuleBuilder[C, T]{name, regmap, nil, nil}
}

// Name returns the name of the module being constructed.
func (p *ModuleBuilder[C, T]) Name() string {
	return p.name
}

// Register returns the register details given an appropriate register
// identifier.
func (p *ModuleBuilder[C, T]) Register(rid uint) schema.Register {
	return p.registers[rid]
}

// Registers returns the set of declared registers in the module being
// constructed.
func (p *ModuleBuilder[C, T]) Registers() []schema.Register {
	return p.registers
}

// RegisterAccessOf returns a register accessor for the register with the given name.
func (p *ModuleBuilder[C, T]) RegisterAccessOf(name string, shift int) *RegisterAccess[T] {
	// Lookup register associated with this name
	rid := p.regmap[name]
	//
	return &RegisterAccess[T]{
		Register: rid,
		Shift:    shift,
	}
}

// NewRegister declares a new register within the module being built.  This will
// panic if a register of the same name already exists.
func (p *ModuleBuilder[C, T]) NewRegister(register schema.Register) uint {
	// Determine identifier
	id := uint(len(p.registers))
	// Sanity check
	if _, ok := p.regmap[register.Name]; ok {
		panic(fmt.Sprintf("register \"%s\" already declared", register.Name))
	}
	//
	p.registers = append(p.registers, register)
	p.regmap[register.Name] = id
	//
	return id
}

// AddConstraint adds a new constraint to this module.
func (p *ModuleBuilder[C, T]) AddConstraint(constraint C) {
	p.constraints = append(p.constraints, constraint)
}

// Build constructs a table module from this module builder.
func (p *ModuleBuilder[C, T]) buildTable() schema.Table[C] {
	table := schema.NewTable[C](p.name)
	table.AddRegisters(p.registers...)
	table.AddConstraints(p.constraints...)
	//
	return table
}
