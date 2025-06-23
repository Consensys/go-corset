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
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// SchemaBuilder is a mechanism for constructing mixed schemas which attempts to
// simplify the problem of mapping source-level names to e.g. module-specific
// register indexes.
type SchemaBuilder[C schema.Constraint, T Term[T]] struct {
	// Modmap maps modules identifers to modules
	modmap map[string]uint
	// Externs represent modules which have already been constructed.  These
	// will be given the lower module identifiers, since they are already
	// packaged and, hence, we must avoid breaking thein linkage.
	externs []schema.Module
	// Modules being constructed
	modules []*ModuleBuilder[C, T]
}

// NewSchemaBuilder constructs a new schema builder with a given number of
// externally defined modules.  Such modules are allocated module indices first.
func NewSchemaBuilder[C schema.Constraint, T Term[T], E schema.Module](externs ...E) SchemaBuilder[C, T] {
	var (
		modmap   = make(map[string]uint, 0)
		nexterns = make([]schema.Module, len(externs))
	)
	// Initialise module map
	for i, m := range externs {
		// Quick sanity check
		if _, ok := modmap[m.Name()]; ok {
			panic(fmt.Sprintf("duplicate module \"%s\" detected", m.Name()))
		}
		//
		modmap[m.Name()] = uint(i)
	}
	// Convert externs
	for i, m := range externs {
		nexterns[i] = m
	}
	//
	return SchemaBuilder[C, T]{modmap, nexterns, nil}
}

// NewModule constructs a new, empty module and returns its unique module
// identifier.
func (p *SchemaBuilder[C, T]) NewModule(name string, multiplier uint) uint {
	var mid = uint(len(p.externs) + len(p.modules))
	// Sanity check this module is not already declared
	if _, ok := p.modmap[name]; ok {
		panic(fmt.Sprintf("module \"%s\" already declared", name))
	}
	//
	p.modules = append(p.modules, NewModuleBuilder[C, T](name, mid, multiplier))
	p.modmap[name] = mid
	//
	return mid
}

// HasModule checks whether a moduleregister of the given name exists already
// and,if so, returns its index.
func (p *SchemaBuilder[C, T]) HasModule(name string) (uint, bool) {
	// Lookup module associated with this name
	mid, ok := p.modmap[name]
	// That's it.
	return mid, ok
}

// Module returns the builder for the given module based on its index.
func (p *SchemaBuilder[C, T]) Module(mid uint) *ModuleBuilder[C, T] {
	var n uint = uint(len(p.externs))
	// Sanity check
	if mid < n {
		return NewExternModuleBuilder[C, T](mid, p.externs[mid])
	}
	//
	return p.modules[mid-n]
}

// ModuleOf returns the builder for the given module based on its name.
func (p *SchemaBuilder[C, T]) ModuleOf(name string) *ModuleBuilder[C, T] {
	return p.Module(p.modmap[name])
}

// Build returns an array of tables constructed by this builder.
func (p *SchemaBuilder[C, T]) Build() []*schema.Table[C] {
	modules := make([]*schema.Table[C], len(p.modules))
	//
	for i, m := range p.modules {
		modules[i] = m.BuildTable()
	}
	//
	return modules
}

// ModuleBuilder provides a mechanism to ease the construction of modules for
// use in schemas.  For example, it maintains a mapping from register names to
// their relevant indices.  It also provides a mechanism for constructing a
// register access based on the register name, etc.
type ModuleBuilder[C schema.Constraint, T Term[T]] struct {
	extern bool
	// Name of the module being constructed
	name string
	// Id of this module
	moduleId schema.ModuleId
	// Length multiplier for this module
	multiplier uint
	// Maps register names (including aliases) to the register number.
	regmap map[string]uint
	// Registers declared for this module
	registers []schema.Register
	// Constraints for this module
	constraints []C
	// Assignments for computed registers
	assignments []schema.Assignment
}

// NewModuleBuilder constructs a new builder for a module with the given name.
func NewModuleBuilder[C schema.Constraint, T Term[T]](name string, mid schema.ModuleId,
	multiplier uint) *ModuleBuilder[C, T] {
	//
	regmap := make(map[string]uint, 0)
	return &ModuleBuilder[C, T]{false, name, mid, multiplier, regmap, nil, nil, nil}
}

// NewExternModuleBuilder constructs a new builder suitable for external modules.
func NewExternModuleBuilder[C schema.Constraint, T Term[T]](mid schema.ModuleId,
	module schema.Module) *ModuleBuilder[C, T] {
	//
	regmap := make(map[string]uint, 0)
	// Initialise register map
	for i, r := range module.Registers() {
		regmap[r.Name] = uint(i)
	}
	// Done
	return &ModuleBuilder[C, T]{true, module.Name(), mid, 1, regmap, module.Registers(), nil, nil}
}

// AddAssignment adds a new assignment to this module.  Assignments are
// responsible for computing the values of computed columns.
func (p *ModuleBuilder[C, T]) AddAssignment(assignment schema.Assignment) {
	if p.extern {
		panic("cannot add assignment to external module")
	}
	//
	p.assignments = append(p.assignments, assignment)
}

// AddConstraint adds a new constraint to this module.
func (p *ModuleBuilder[C, T]) AddConstraint(constraint C) {
	if p.extern {
		panic("cannot add constraint to external module")
	}
	//
	p.constraints = append(p.constraints, constraint)
}

// Assignments returns an iterator over the assignments of this schema.
// These are the computations used to assign values to all computed columns
// in this module.
func (p *ModuleBuilder[C, T]) Assignments() iter.Iterator[schema.Assignment] {
	return iter.NewArrayIterator(p.assignments)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p *ModuleBuilder[C, T]) Consistent(schema schema.AnySchema) []error {
	var errors []error
	// Check constraints
	for _, c := range p.constraints {
		errors = append(errors, c.Consistent(schema)...)
	}
	// Check assignments
	for _, a := range p.assignments {
		errors = append(errors, a.Consistent(schema)...)
	}
	// Done
	return errors
}

// Constraints provides access to those constraints associated with this
// module.
func (p *ModuleBuilder[C, T]) Constraints() iter.Iterator[schema.Constraint] {
	i := iter.NewArrayIterator(p.constraints)
	return iter.NewCastIterator[C, schema.Constraint](i)
}

// Id returns the module index of this module.
func (p *ModuleBuilder[C, T]) Id() uint {
	return p.moduleId
}

// LengthMultiplier identifies the length multiplier for this module.  For every
// trace, the height of the corresponding module must be a multiple of this.
// This is used specifically to support interleaving constraints.
func (p *ModuleBuilder[C, T]) LengthMultiplier() uint {
	return p.multiplier
}

// Width returns the number of registers in this module.
func (p *ModuleBuilder[C, T]) Width() uint {
	return uint(len(p.registers))
}

// HasRegister checks whether a register of the given name exists already and,
// if so, returns its index.
func (p *ModuleBuilder[C, T]) HasRegister(name string) (schema.RegisterId, bool) {
	// Lookup register associated with this name
	rid, ok := p.regmap[name]
	//
	return schema.NewRegisterId(rid), ok
}

// Name returns the name of the module being constructed.
func (p *ModuleBuilder[C, T]) Name() string {
	return p.name
}

// NewRegister declares a new register within the module being built.  This will
// panic if a register of the same name already exists.
func (p *ModuleBuilder[C, T]) NewRegister(register schema.Register) schema.RegisterId {
	// Determine identifier
	id := uint(len(p.registers))
	// Sanity check
	if _, ok := p.regmap[register.Name]; ok {
		panic(fmt.Sprintf("register \"%s\" already declared", register.Name))
	} else if p.extern {
		panic("cannot add register to external module")
	}
	//
	p.registers = append(p.registers, register)
	p.regmap[register.Name] = id
	//
	return schema.NewRegisterId(id)
}

// NewRegisters declares zero or more new registers within the module being
// built.  This will panic if a register of the same name already exists.
func (p *ModuleBuilder[C, T]) NewRegisters(registers ...schema.Register) {
	for _, r := range registers {
		p.NewRegister(r)
	}
}

// Register returns the register details given an appropriate register
// identifier.
func (p *ModuleBuilder[C, T]) Register(rid schema.RegisterId) schema.Register {
	return p.registers[rid.Unwrap()]
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
		Register: schema.NewRegisterId(rid),
		Shift:    shift,
	}
}

// BuildTable constructs a table module from this module builder.
func (p *ModuleBuilder[C, T]) BuildTable() *schema.Table[C] {
	if p.extern {
		panic("cannot build externally defined module")
	}
	//
	table := schema.NewTable[C](p.name, p.multiplier)
	table.AddRegisters(p.registers...)
	table.AddConstraints(p.constraints...)
	table.AddAssignments(p.assignments...)
	//
	return table
}
