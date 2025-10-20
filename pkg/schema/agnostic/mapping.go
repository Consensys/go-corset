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
package agnostic

import (
	"fmt"
	"strings"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// NewModuleMap constructs a new module map
func NewModuleMap[T sc.RegisterMap](field sc.FieldConfig, modules []T) sc.ModuleMap[T] {
	return limbsMap[T]{field, modules}
}

// ConvertModuleMap converts a module map of one kind into a module map of another kind.
func ConvertModuleMap[S, T sc.RegisterMap](mapping sc.ModuleMap[S], fn func(S) T) sc.ModuleMap[T] {
	var (
		mods = make([]T, mapping.Width())
	)
	//
	for i := range mapping.Width() {
		mods[i] = fn(mapping.Module(i))
	}
	//
	return NewModuleMap(mapping.Field(), mods)
}

// NewLimbsMap constructs a new schema mapping for a given schema and
// parameter combination.  This determines, amongst other things,  the
// composition of limbs for all registers in the schema.
func NewLimbsMap[F any, M sc.Module[F]](field sc.FieldConfig, modules ...M) sc.LimbsMap {
	var mappings []sc.RegisterLimbsMap
	//
	for _, m := range modules {
		regmap := newRegisterMapping(field, m)
		mappings = append(mappings, regmap)
	}
	//
	return limbsMap[sc.RegisterLimbsMap]{field, mappings}
}

// ============================================================================
// LimbMap
// ============================================================================

// limbsMap provides a straightforward implementation of the schema.LimbMap
// interface.
type limbsMap[T sc.RegisterMap] struct {
	field   sc.FieldConfig
	modules []T
}

// Field implementation for schema.LimbMap interface
func (p limbsMap[T]) Field() sc.FieldConfig {
	return p.field
}

// Module implementation for schema.RegisterMappings interface
func (p limbsMap[T]) Module(mid sc.ModuleId) T {
	return p.modules[mid]
}

// ModuleOf implementation for schema.RegisterMappings interface
func (p limbsMap[T]) ModuleOf(name string) T {
	for _, m := range p.modules {
		if m.Name() == name {
			return m
		}
	}
	//
	panic(fmt.Sprintf("unknown module \"%s\"", name))
}

// Width returns the number of modules in this map
func (p limbsMap[T]) Width() uint {
	return uint(len(p.modules))
}

func (p limbsMap[T]) String() string {
	var builder strings.Builder
	//
	builder.WriteString("[")
	builder.WriteString(p.field.Name)
	builder.WriteString(":")
	//
	for i, m := range p.modules {
		if i != 0 {
			builder.WriteString(";")
		}
		//
		builder.WriteString(m.String())
	}

	builder.WriteString("]")

	return builder.String()
}

// ============================================================================
// RegisterLimbMap
// ============================================================================

// registerLimbsMap provides a mapping from registers from the original schema to
// registers (referred to as limbs) in the split schema.   In some cases, there
// may be only one limb matching the original register above exactly (i.e. when
// the register width was already below the cutoff); in other cases, there can
// be many limbs for a single register above.  It should always be the case that
// the total width of limbs matches that of the original register.  Furthermore,
// if the original register was computed, then the limbs should be also, etc.
type registerLimbsMap struct {
	// Name of the module to which this mapping corresponds
	name string
	// Field configuration in play
	field sc.FieldConfig
	// Set of registers in the original schema (i.e. as they were before the
	// split)
	registers []sc.Register
	// Set of "limbs" (i.e registers) in the split schema.
	limbs []sc.Register
	// Mapping for each register above to its corresponding set of limbs.
	mapping [][]sc.LimbId
}

// newRegisterMapping constructs an appropriate register map for a given module
// and parameter combination.
func newRegisterMapping[F any](field sc.FieldConfig, module sc.Module[F]) registerLimbsMap {
	var (
		regs    = module.Registers()
		limbs   []sc.Register
		mapping = make([][]sc.LimbId, len(regs))
		limbId  = uint(0)
	)
	// Split up limbs
	for i, r := range regs {
		ls := SplitIntoLimbs(field.RegisterWidth, r)
		limbs = append(limbs, ls...)
		// build mapping
		m := make([]sc.RegisterId, len(ls))
		//
		for i := 0; i != len(m); i++ {
			m[i] = sc.NewRegisterId(limbId)
			limbId++
		}
		// Assign mapping
		mapping[i] = m
	}
	// Done
	return registerLimbsMap{
		module.Name(),
		field,
		regs,
		limbs,
		mapping,
	}
}

// Field implementation for schema.RegisterMappings interface
func (p registerLimbsMap) Field() sc.FieldConfig {
	return p.field
}

// Limbs implementation for the schema.RegisterMapping interface
func (p registerLimbsMap) LimbIds(reg sc.RegisterId) []sc.LimbId {
	return p.mapping[reg.Unwrap()]
}

// Limb implementation for the schema.RegisterMapping interface
func (p registerLimbsMap) Limb(reg sc.LimbId) sc.Limb {
	return p.limbs[reg.Unwrap()]
}

// Limbs implementation for the schema.RegisterMapping interface
func (p registerLimbsMap) Limbs() []sc.Limb {
	return p.limbs
}

// LimbsMap implementation for the schema.RegisterMapping interface
func (p registerLimbsMap) LimbsMap() sc.RegisterMap {
	return registerLimbsMap{
		p.name, p.field, p.limbs, nil, nil,
	}
}

// Name implementation for schema.RegisterMapping interface
func (p registerLimbsMap) Name() string {
	return p.name
}

// RegisterOf determines a register's ID based on its name.
func (p registerLimbsMap) RegisterOf(name string) sc.RegisterId {
	for i, reg := range p.registers {
		if reg.Name == name {
			return sc.NewRegisterId(uint(i))
		}
	}
	//
	panic(fmt.Sprintf("unknown register \"%s\"", name))
}

// HasRegister implementation for RegisterMap interface.
func (p registerLimbsMap) HasRegister(name string) (sc.RegisterId, bool) {
	for i, reg := range p.registers {
		if reg.Name == name {
			return sc.NewRegisterId(uint(i)), true
		}
	}
	//
	return sc.NewUnusedRegisterId(), false
}

// Register implementation for RegisterMap interface.
func (p registerLimbsMap) Register(rid sc.RegisterId) sc.Register {
	return p.registers[rid.Unwrap()]
}

// Registers implementation for RegisterMap interface.
func (p registerLimbsMap) Registers() []sc.Register {
	return p.registers
}

func (p registerLimbsMap) String() string {
	return sc.RegisterLimbsMapToString(p)
}
