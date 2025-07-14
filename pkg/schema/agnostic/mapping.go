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

	sc "github.com/consensys/go-corset/pkg/schema"
)

// Subdivide a mixed schema of field agnostic modules according to the given
// bandwidth and maximum register width requirements.  See discussion of
// FieldAgnosticModule for more on this process.
func Subdivide[M1 sc.FieldAgnosticModule[M1], M2 sc.FieldAgnosticModule[M2]](
	field sc.FieldConfig, schema sc.MixedSchema[M1, M2]) (sc.MixedSchema[M1, M2], sc.RegisterMap) {
	//
	var (
		left    []M1 = make([]M1, len(schema.LeftModules()))
		right   []M2 = make([]M2, len(schema.RightModules()))
		mapping      = newRegisterMappings(field, schema)
	)
	// Subdivide the left
	for i, m := range schema.LeftModules() {
		left[i] = m.Subdivide(mapping)
	}
	// Subdivide the right
	for i, m := range schema.RightModules() {
		right[i] = m.Subdivide(mapping)
	}
	// Done
	return sc.NewMixedSchema(left, right), mapping
}

// ============================================================================
// SchemaMapping
// ============================================================================

// RegisterMappings provides a straightforward implementation of the
// schema.RegisterMappings interface.
type registerMappings struct {
	field   sc.FieldConfig
	modules []registerMapping
}

// newRegisterMappings constructs a new schema mapping for a given schema and
// parameter combination.  This determines, amongst other things,  the
// composition of limbs for all registers in the schema.
func newRegisterMappings(field sc.FieldConfig, schema sc.AnySchema) sc.RegisterMap {
	var mappings []registerMapping
	//
	for i := range schema.Width() {
		regmap := newRegisterMapping(field, schema.Module(i))
		mappings = append(mappings, regmap)
	}
	//
	return registerMappings{field, mappings}
}

// Field implementation for schema.RegisterMappings interface
func (p registerMappings) Field() sc.FieldConfig {
	return p.field
}

// Module implementation for schema.RegisterMappings interface
func (p registerMappings) Module(mid sc.ModuleId) sc.ModuleRegisterMap {
	return p.modules[mid]
}

// ModuleOf implementation for schema.RegisterMappings interface
func (p registerMappings) ModuleOf(name string) sc.ModuleRegisterMap {
	for _, m := range p.modules {
		if m.name == name {
			return m
		}
	}
	//
	panic(fmt.Sprintf("unknown module \"%s\"", name))
}

// ============================================================================
// RegisterMapping
// ============================================================================

// RegisterMapping provides a mapping from registers from the original schema to
// registers (referred to as limbs) in the split schema.   In some cases, there
// may be only one limb matching the original register above exactly (i.e. when
// the register width was already below the cutoff); in other cases, there can
// be many limbs for a single register above.  It should always be the case that
// the total width of limbs matches that of the original register.  Furthermore,
// if the original register was computed, then the limbs should be also, etc.
type registerMapping struct {
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
func newRegisterMapping(field sc.FieldConfig, module sc.Module) registerMapping {
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
	return registerMapping{
		module.Name(),
		field,
		regs,
		limbs,
		mapping,
	}
}

// Field implementation for schema.RegisterMappings interface
func (p registerMapping) Field() sc.FieldConfig {
	return p.field
}

// Limbs implementation for the schema.RegisterMapping interface
func (p registerMapping) LimbIds(reg sc.RegisterId) []sc.LimbId {
	return p.mapping[reg.Unwrap()]
}

// Limb implementation for the schema.RegisterMapping interface
func (p registerMapping) Limb(reg sc.LimbId) sc.Limb {
	return p.limbs[reg.Unwrap()]
}

// Limbs implementation for the schema.RegisterMapping interface
func (p registerMapping) Limbs() []sc.Limb {
	return p.limbs
}

// RegisterOf determines a register's ID based on its name.
func (p registerMapping) RegisterOf(name string) sc.RegisterId {
	for i, reg := range p.registers {
		if reg.Name == name {
			return sc.NewRegisterId(uint(i))
		}
	}
	//
	panic(fmt.Sprintf("unknown register \"%s\"", name))
}
