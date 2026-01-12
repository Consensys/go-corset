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

	"github.com/consensys/go-corset/pkg/ir/term"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ModuleBuilder provides a mechanism to ease the construction of modules for
// use in schemas.  For example, it maintains a mapping from register names to
// their relevant indices.  It also provides a mechanism for constructing a
// register access based on the register name, etc.
//
// NOTE: overall, this interface has got somewhat out-of-hand and it would be
// useful to try and simplify it where possible.
type ModuleBuilder[F field.Element[F], C schema.Constraint[F], T term.Expr[F, T]] interface {
	fmt.Stringer
	schema.ModuleView
	// AddAssignment adds a new assignment to this module.  Assignments are
	// responsible for computing the values of computed columns.
	AddAssignment(assignment schema.Assignment[F])
	// AddConstraint adds a new constraint to this module.
	AddConstraint(constraint C)
	// AllowPadding determines whether the given module allows an initial
	// padding row, or not.
	AllowPadding() bool
	// Assignments returns those assignments added to this module.
	Assignments() []schema.Assignment[F]
	// Constraints returns those constraints added to this module.
	Constraints() []C
	// Id returns the module index of this module.
	Id() uint
	// IsExtern determines whether or not this is an external module or not.
	IsExtern() bool
	// NewRegister declares a new register within the module being built.  This will
	// panic if a register of the same name already exists.
	NewRegister(reg register.Register) register.Id
	// NewRegisters declares zero or more new registers within the module being
	// built.  This will panic if a register of the same name already exists.
	NewRegisters(registers ...register.Register)
	// ZeroRegister returns an ID for the "zero register".  That is, a register
	// which is always zero.  If no such register exists already, one is
	// created.
	ConstRegister(constant uint8) register.Id
}

// ============================================================================
// Internal Module Builder
// ============================================================================

// NewModuleBuilder constructs a new builder for a module with the given name.
func NewModuleBuilder[F field.Element[F], C schema.Constraint[F], T term.Expr[F, T]](name module.Name,
	mid schema.ModuleId, padding, public, synthetic bool) ModuleBuilder[F, C, T] {
	//
	regmap := make(map[string]uint, 0)
	return &internalModuleBuilder[F, C, T]{name, mid, padding, public, synthetic, regmap, nil, nil, nil}
}

type internalModuleBuilder[F field.Element[F], C schema.Constraint[F], T term.Expr[F, T]] struct {
	// Name of the module being constructed
	name module.Name
	// Id of this module
	moduleId schema.ModuleId
	// Indicates whether padding supported for this module
	padding bool
	// Indicates whether externally visible
	public bool
	// Indicates whether this is a synthetic module or not
	synthetic bool
	// Maps register names (including aliases) to the register number.
	regmap map[string]uint
	// Registers declared for this module
	registers []register.Register
	// Constraints for this module
	constraints []C
	// Assignments for computed registers
	assignments []schema.Assignment[F]
}

// AddAssignment implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) AddAssignment(assignment schema.Assignment[F]) {
	p.assignments = append(p.assignments, assignment)
}

// AddConstraint implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) AddConstraint(constraint C) {
	p.constraints = append(p.constraints, constraint)
}

// Assignments implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) Assignments() []schema.Assignment[F] {
	return p.assignments
}

// Constraints implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) Constraints() []C {
	return p.constraints
}

// Id implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) Id() uint {
	return p.moduleId
}

// AllowPadding implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) AllowPadding() bool {
	return p.padding
}

// IsExtern implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) IsExtern() bool {
	return false
}

// IsPublic implementation for schema.ModuleView interface.
func (p *internalModuleBuilder[F, C, T]) IsPublic() bool {
	return p.public
}

// IsSynthetic implementation for schema.ModuleView interface.
func (p *internalModuleBuilder[F, C, T]) IsSynthetic() bool {
	return p.synthetic
}

// Width implementation for schema.ModuleView interface.
func (p *internalModuleBuilder[F, C, T]) Width() uint {
	return uint(len(p.registers))
}

// HasRegister implementation for register.Map interface.
func (p *internalModuleBuilder[F, C, T]) HasRegister(name string) (register.Id, bool) {
	// Lookup register associated with this name
	rid, ok := p.regmap[name]
	//
	return register.NewId(rid), ok
}

// Name implementation for register.Map interface.
func (p *internalModuleBuilder[F, C, T]) Name() module.Name {
	return p.name
}

// NewRegister implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) NewRegister(reg register.Register) register.Id {
	// Determine identifier
	id := uint(len(p.registers))
	// Sanity check
	if _, ok := p.regmap[reg.Name]; ok {
		panic(fmt.Sprintf("register \"%s\" already declared", reg.Name))
	}
	//
	p.registers = append(p.registers, reg)
	p.regmap[reg.Name] = id
	//
	return register.NewId(id)
}

// NewRegisters implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) NewRegisters(registers ...register.Register) {
	for _, r := range registers {
		p.NewRegister(r)
	}
}

// Register implementation for register.Map interface.
func (p *internalModuleBuilder[F, C, T]) Register(rid register.Id) register.Register {
	return p.registers[rid.Unwrap()]
}

// Registers implementation for register.Map interface.
func (p *internalModuleBuilder[F, C, T]) Registers() []register.Register {
	return p.registers
}

// RegisterAccessOf implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) RegisterAccessOf(name string, shift int) *term.RegisterAccess[F, T] {
	// Lookup register associated with this name
	var (
		rid = register.NewId(p.regmap[name])
		reg = p.Register(rid)
	)
	//
	return term.RawRegisterAccess[F, T](rid, reg.Width, shift)
}

func (p *internalModuleBuilder[F, C, T]) String() string {
	return register.MapToString(p)
}

// ZeroRegister implementation for ModuleBuilder interface.
func (p *internalModuleBuilder[F, C, T]) ConstRegister(constant uint8) register.Id {
	var name = fmt.Sprintf("%d", constant)
	// Check whether register already exists
	if rid, ok := p.HasRegister(name); ok {
		return rid
	}
	// If not, create a new one.
	return p.NewRegister(register.NewConst(constant))
}

// ============================================================================
// External Module Builder
// ============================================================================

// NewExternModuleBuilder constructs a new builder suitable for external
// modules.  These are just used for linking purposes.
func NewExternModuleBuilder[F field.Element[F], C schema.Constraint[F], T term.Expr[F, T]](mid schema.ModuleId,
	module register.ConstMap) ModuleBuilder[F, C, T] {
	return &externalModuleBuilder[F, C, T]{mid, module}
}

// externalModuleBuilder essentially provides a wrapper for an externally
// defined module to allow it to be accessed as though it were an internal
// module.
type externalModuleBuilder[F field.Element[F], C schema.Constraint[F], T term.Expr[F, T]] struct {
	// Id of this module
	moduleId schema.ModuleId
	// External source
	module register.ConstMap
}

// AddAssignment implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) AddAssignment(assignment schema.Assignment[F]) {
	panic("cannot add assignment to external module")
}

// AddConstraint implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) AddConstraint(constraint C) {
	panic("cannot add constraint to external module")
}

// Assignments implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) Assignments() []schema.Assignment[F] {
	return nil
}

// Constraints implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) Constraints() []C {
	return nil
}

// Id implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) Id() uint {
	return p.moduleId
}

// AllowPadding implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) AllowPadding() bool {
	return false
}

// IsExtern implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) IsExtern() bool {
	return true
}

// IsPublic implementation for schema.ModuleView interface.
func (p *externalModuleBuilder[F, C, T]) IsPublic() bool {
	return false
}

// IsSynthetic implementation for schema.ModuleView interface.
func (p *externalModuleBuilder[F, C, T]) IsSynthetic() bool {
	return false
}

// Width implementation for schema.ModuleView interface.
func (p *externalModuleBuilder[F, C, T]) Width() uint {
	return uint(len(p.module.Registers()))
}

// HasRegister implementation for register.Map interface.
func (p *externalModuleBuilder[F, C, T]) HasRegister(name string) (register.Id, bool) {
	return p.module.HasRegister(name)
}

// Name implementation for register.Map interface.
func (p *externalModuleBuilder[F, C, T]) Name() module.Name {
	return p.module.Name()
}

// NewRegister implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) NewRegister(reg register.Register) register.Id {
	panic("cannot add register to external module")
}

// NewRegisters implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) NewRegisters(registers ...register.Register) {
	for _, r := range registers {
		p.NewRegister(r)
	}
}

// Register implementation for register.Map interface.
func (p *externalModuleBuilder[F, C, T]) Register(rid register.Id) register.Register {
	return p.module.Register(rid)
}

// Registers implementation for register.Map interface.
func (p *externalModuleBuilder[F, C, T]) Registers() []register.Register {
	return p.module.Registers()
}

func (p *externalModuleBuilder[F, C, T]) String() string {
	return register.MapToString(p)
}

// ZeroRegister implementation for ModuleBuilder interface.
func (p *externalModuleBuilder[F, C, T]) ConstRegister(constant uint8) register.Id {
	return p.module.ConstRegister(constant)
}
