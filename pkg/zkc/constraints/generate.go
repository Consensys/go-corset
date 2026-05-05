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
package constraints

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
)

// GenerateAirConstraints is responsible for converting a field machine into a
// corresponding set of AIR constraints.
func GenerateAirConstraints[F field.Element[F]](fm *vm.FieldMachine[F]) air.Schema[F] {
	var (
		modules = make([]air.Module[F], len(fm.Modules()))
	)
	//
	for i, m := range fm.Modules() {
		modules[i] = translateModule[F](uint(i), m)
	}
	//
	return schema.NewUniformSchema(modules)
}

func translateModule[F field.Element[F]](ctx schema.ModuleId, fm vm.Module) air.Module[F] {
	switch fm := fm.(type) {
	case *vm.FieldFunction:
		return translateFunction[F](ctx, *fm)
	default:
		panic(fmt.Sprintf("unknown module \"%s\" encountered", fm.Name()))
	}
}

func translateFunction[F field.Element[F]](ctx schema.ModuleId, fm vm.FieldFunction) air.Module[F] {
	var (
		mod  *schema.Table[F, air.Constraint[F]]
		name = trace.ModuleName{Name: fm.Name(), Multiplier: 1}
	)
	// Initialise module
	mod = mod.Init(name, true, true, false, 0)
	// Add all registers
	mod.AddRegisters(fm.Registers()...)
	// Transle all instructions
	for i, vec := range fm.Code() {
		mod.AddConstraints(translateVectorInstruction[F](ctx, uint(i), vec)...)
	}
	// Done
	return mod
}

func translateVectorInstruction[F field.Element[F]](ctx schema.ModuleId, idx uint, vec vm.Vector[vm.FieldInstruction],
) []air.Constraint[F] {
	//
	var (
		constraints []air.Constraint[F]
		// generate write map
		_ = vec.WriteMap()
		// generate branch table
		_ = vec.BranchTable()
	)
	//
	for i, insn := range vec.Codes {
		//
		switch insn := insn.(type) {
		case *instruction.Debug:
			// no-operation
		case *instruction.FieldAssign[F]:
			constraints = append(constraints, translateFieldAssignment(ctx, idx, uint(i), *insn))
		case *instruction.Return:
			// no-operation (for now)
		default:
			panic("todo")
		}
	}
	//
	return constraints
}

func translateFieldAssignment[F field.Element[F]](ctx schema.ModuleId, macro, micro uint,
	insn instruction.FieldAssign[F]) air.Constraint[F] {
	var (
		handle = fmt.Sprintf("pc%d_%d", macro, micro)
	)
	//
	return air.NewVanishingConstraint(handle, ctx, util.None[int](), insn.Source)
}
