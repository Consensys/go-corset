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
	"github.com/consensys/go-corset/pkg/util/field"
)

// BuildableModule embodies the notion of a module which can be initialised from
// the various required components.  This provides a useful way for constructing
// modules once all the various pieces of information have been finalised.
type BuildableModule[F any, C schema.Constraint[F], M any] interface {
	Init(name string, multiplier uint, padding bool) M
	// Add one or more assignments to this buildable module
	AddAssignments(assignments ...schema.Assignment[F])
	// Add one or more constraints to this buildable module
	AddConstraints(constraints ...C)
	// Add one or more registers to this buildable module.
	AddRegisters(registers ...schema.Register)
}

// BuildSchema builds all modules defined within a give SchemaBuilder instance.
func BuildSchema[M BuildableModule[F, C, M], F field.Element[F], C schema.Constraint[F], T Term[F, T]](
	p SchemaBuilder[F, C, T]) []M {
	//
	var modules = make([]M, len(p.modules))
	//
	for i, m := range p.modules {
		modules[i] = BuildModule[F, C, T, M](*m)
	}
	//
	return modules
}

// BuildModule builds a module from a given ModuleBuilder instance.
func BuildModule[F field.Element[F], C schema.Constraint[F], T Term[F, T], M BuildableModule[F, C, M]](
	m ModuleBuilder[F, C, T]) M {
	//
	var module M
	// Build it
	module = module.Init(m.name, m.multiplier, m.padding)
	module.AddRegisters(m.registers...)
	module.AddAssignments(m.assignments...)
	module.AddConstraints(m.constraints...)
	// Done
	return module
}

// SchemaBuilder is a mechanism for constructing mixed schemas which attempts to
// simplify the problem of mapping source-level names to e.g. module-specific
// register indexes.
type SchemaBuilder[F field.Element[F], C schema.Constraint[F], T Term[F, T]] struct {
	// Modmap maps modules identifers to modules
	modmap map[string]uint
	// Externs represent modules which have already been constructed.  These
	// will be given the lower module identifiers, since they are already
	// packaged and, hence, we must avoid breaking thein linkage.
	externs []schema.Module[F]
	// Modules being constructed
	modules []*ModuleBuilder[F, C, T]
}

// NewSchemaBuilder constructs a new schema builder with a given number of
// externally defined modules.  Such modules are allocated module indices first.
func NewSchemaBuilder[F field.Element[F], C schema.Constraint[F], T Term[F, T], E schema.Module[F]](externs ...E,
) SchemaBuilder[F, C, T] {
	var (
		modmap   = make(map[string]uint, 0)
		nexterns = make([]schema.Module[F], len(externs))
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
	return SchemaBuilder[F, C, T]{modmap, nexterns, nil}
}

// NewModule constructs a new, empty module and returns its unique module
// identifier.
func (p *SchemaBuilder[F, C, T]) NewModule(name string, multiplier uint, padding bool) uint {
	var mid = uint(len(p.externs) + len(p.modules))
	// Sanity check this module is not already declared
	if _, ok := p.modmap[name]; ok {
		panic(fmt.Sprintf("module \"%s\" already declared", name))
	}
	//
	p.modules = append(p.modules, NewModuleBuilder[F, C, T](name, mid, multiplier, padding))
	p.modmap[name] = mid
	//
	return mid
}

// HasModule checks whether a moduleregister of the given name exists already
// and,if so, returns its index.
func (p *SchemaBuilder[F, C, T]) HasModule(name string) (uint, bool) {
	// Lookup module associated with this name
	mid, ok := p.modmap[name]
	// That's it.
	return mid, ok
}

// Module returns the builder for the given module based on its index.
func (p *SchemaBuilder[F, C, T]) Module(mid uint) *ModuleBuilder[F, C, T] {
	var n uint = uint(len(p.externs))
	// Sanity check
	if mid < n {
		return NewExternModuleBuilder[F, C, T](mid, p.externs[mid])
	}
	//
	return p.modules[mid-n]
}

// ModuleOf returns the builder for the given module based on its name.
func (p *SchemaBuilder[F, C, T]) ModuleOf(name string) *ModuleBuilder[F, C, T] {
	return p.Module(p.modmap[name])
}

// ModuleBuilder provides a mechanism to ease the construction of modules for
// use in schemas.  For example, it maintains a mapping from register names to
// their relevant indices.  It also provides a mechanism for constructing a
// register access based on the register name, etc.
type ModuleBuilder[F field.Element[F], C schema.Constraint[F], T Term[F, T]] struct {
	extern bool
	// Name of the module being constructed
	name string
	// Id of this module
	moduleId schema.ModuleId
	// Length multiplier for this module
	multiplier uint
	// Indicates whether padding supported for this module
	padding bool
	// Maps register names (including aliases) to the register number.
	regmap map[string]uint
	// Registers declared for this module
	registers []schema.Register
	// Constraints for this module
	constraints []C
	// Assignments for computed registers
	assignments []schema.Assignment[F]
}

// NewModuleBuilder constructs a new builder for a module with the given name.
func NewModuleBuilder[F field.Element[F], C schema.Constraint[F], T Term[F, T]](name string, mid schema.ModuleId,
	multiplier uint, padding bool) *ModuleBuilder[F, C, T] {
	//
	regmap := make(map[string]uint, 0)
	return &ModuleBuilder[F, C, T]{false, name, mid, multiplier, padding, regmap, nil, nil, nil}
}

// NewExternModuleBuilder constructs a new builder suitable for external
// modules.  These are just used for linking purposes.
func NewExternModuleBuilder[F field.Element[F], C schema.Constraint[F], T Term[F, T]](mid schema.ModuleId,
	module schema.Module[F]) *ModuleBuilder[F, C, T] {
	//
	regmap := make(map[string]uint, 0)
	// Initialise register map
	for i, r := range module.Registers() {
		regmap[r.Name] = uint(i)
	}
	// Done
	return &ModuleBuilder[F, C, T]{true, module.Name(), mid, 1, false, regmap, module.Registers(), nil, nil}
}

// AddAssignment adds a new assignment to this module.  Assignments are
// responsible for computing the values of computed columns.
func (p *ModuleBuilder[F, C, T]) AddAssignment(assignment schema.Assignment[F]) {
	if p.extern {
		panic("cannot add assignment to external module")
	}
	//
	p.assignments = append(p.assignments, assignment)
}

// AddConstraint adds a new constraint to this module.
func (p *ModuleBuilder[F, C, T]) AddConstraint(constraint C) {
	if p.extern {
		panic("cannot add constraint to external module")
	}
	//
	p.constraints = append(p.constraints, constraint)
}

// Assignments returns an iterator over the assignments of this schema.
// These are the computations used to assign values to all computed columns
// in this module.
func (p *ModuleBuilder[F, C, T]) Assignments() iter.Iterator[schema.Assignment[F]] {
	return iter.NewArrayIterator(p.assignments)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p *ModuleBuilder[F, C, T]) Consistent(schema schema.AnySchema[F]) []error {
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
func (p *ModuleBuilder[F, C, T]) Constraints() iter.Iterator[schema.Constraint[F]] {
	i := iter.NewArrayIterator(p.constraints)
	return iter.NewCastIterator[C, schema.Constraint[F]](i)
}

// Id returns the module index of this module.
func (p *ModuleBuilder[F, C, T]) Id() uint {
	return p.moduleId
}

// LengthMultiplier identifies the length multiplier for this module.  For every
// trace, the height of the corresponding module must be a multiple of this.
// This is used specifically to support interleaving constraints.
func (p *ModuleBuilder[F, C, T]) LengthMultiplier() uint {
	return p.multiplier
}

// AllowPadding determines the minimum amount of padding requested at the
// beginning of the module.  This is necessary because legacy modules expect an
// initial padding row.
func (p *ModuleBuilder[F, C, T]) AllowPadding() bool {
	return p.padding
}

// Width returns the number of registers in this module.
func (p *ModuleBuilder[F, C, T]) Width() uint {
	return uint(len(p.registers))
}

// HasRegister checks whether a register of the given name exists already and,
// if so, returns its index.
func (p *ModuleBuilder[F, C, T]) HasRegister(name string) (schema.RegisterId, bool) {
	// Lookup register associated with this name
	rid, ok := p.regmap[name]
	//
	return schema.NewRegisterId(rid), ok
}

// Name returns the name of the module being constructed.
func (p *ModuleBuilder[F, C, T]) Name() string {
	return p.name
}

// NewRegister declares a new register within the module being built.  This will
// panic if a register of the same name already exists.
func (p *ModuleBuilder[F, C, T]) NewRegister(register schema.Register) schema.RegisterId {
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
func (p *ModuleBuilder[F, C, T]) NewRegisters(registers ...schema.Register) {
	for _, r := range registers {
		p.NewRegister(r)
	}
}

// Register returns the register details given an appropriate register
// identifier.
func (p *ModuleBuilder[F, C, T]) Register(rid schema.RegisterId) schema.Register {
	return p.registers[rid.Unwrap()]
}

// Registers returns the set of declared registers in the module being
// constructed.
func (p *ModuleBuilder[F, C, T]) Registers() []schema.Register {
	return p.registers
}

// RegisterAccessOf returns a register accessor for the register with the given name.
func (p *ModuleBuilder[F, C, T]) RegisterAccessOf(name string, shift int) *RegisterAccess[F, T] {
	// Lookup register associated with this name
	rid := p.regmap[name]
	//
	return &RegisterAccess[F, T]{
		Register: schema.NewRegisterId(rid),
		Shift:    shift,
	}
}
