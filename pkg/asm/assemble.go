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
package asm

import (
	"math"

	"github.com/consensys/go-corset/pkg/asm/assembler"
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source"
)

// Register describes a single register within a function.
type Register = io.Register

// MacroFunction is a function whose instructions are themselves macro
// instructions.  A macro function must be compiled down into a micro function
// before we can generate constraints.
type MacroFunction[F field.Element[F]] = io.Function[F, macro.Instruction]

// MicroFunction is a function whose instructions are themselves micro
// instructions.  A micro function represents the lowest representation of a
// function, where each instruction is made up of microcodes.
type MicroFunction[F field.Element[F]] = io.Function[F, micro.Instruction]

// MacroProgram represents a set of components at the macro level.
type MacroProgram[F field.Element[F]] = io.Program[F, macro.Instruction]

// MicroProgram represents a set of components at the micro level.
type MicroProgram[F field.Element[F]] = io.Program[F, micro.Instruction]

// MixedMacroProgram is a schema comprised of both macro assembly components and
// MIR components (hence the term mixed).
type MixedMacroProgram[F field.Element[F]] = schema.MixedSchema[F, *MacroFunction[F], mir.Module[F]]

// MixedMicroProgram is a schema comprised of both micro assembly components and
// MIR components (hence the term mixed).
type MixedMicroProgram[F field.Element[F]] = schema.MixedSchema[F, *MicroFunction[F], mir.Module[F]]

// Assemble takes a given set of assembly files, and parses them into a given
// set of functions.  This includes performing various checks on the files, such
// as type checking, etc.
func Assemble[F field.Element[F]](assembly ...source.File) (
	MacroProgram[F], source.Maps[any], []source.SyntaxError) {
	//
	var (
		items   []assembler.AssemblyItem[F]
		errors  []source.SyntaxError
		program MacroProgram[F]
		srcmaps source.Maps[any]
	)
	// Parse each file in turn.
	for _, asm := range assembly {
		// Parse source file
		cs, errs := assembler.Parse[F](&asm)
		if len(errs) == 0 {
			items = append(items, cs)
		}
		//
		errors = append(errors, errs...)
	}
	// Link assembly
	if len(errors) != 0 {
		return program, srcmaps, errors
	}
	// Link assembly and resolve buses
	program, srcmaps = assembler.Link[F](items...)
	// Well-formedness checks (assuming unlimited field width).
	errors = assembler.Validate(math.MaxUint, program, srcmaps)
	// Done
	return program, srcmaps, errors
}

// LowerMixedMacroProgram a mixed macro program (i.e. schema) into a mixed micro program, using
// vectorisation if desired.  Specifically, any macro modules within the schema
// are lowered into "micro" modules (i.e. those using only micro instructions).
// This does not impact any externally defined (e.g. MIR) modules in the schema.
func LowerMixedMacroProgram[F field.Element[F]](vectorize bool, p MixedMacroProgram[F]) MixedMicroProgram[F] {
	functions := make([]*MicroFunction[F], len(p.LeftModules()))
	//
	for i, f := range p.LeftModules() {
		nf := lowerFunction(vectorize, *f)
		functions[i] = &nf
	}
	// Construct program for validation
	program := io.NewProgram(functions...)
	// Validate generated program.  Whilst not strictly necessary, it is useful
	// from a debugging perspective.
	assembler.ValidateMicro(math.MaxUint, program)
	// Construct mixed micro schema
	return schema.NewMixedSchema(functions, p.RightModules())
}
