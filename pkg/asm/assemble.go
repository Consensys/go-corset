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
	"path/filepath"

	"github.com/consensys/go-corset/pkg/asm/assembler"
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/asm/program"
	"github.com/consensys/go-corset/pkg/ir/hir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/word"
)

// Register describes a single register within a function.
type Register = io.Register

// MacroFunction is a function whose instructions are themselves macro
// instructions.  A macro function must be compiled down into a micro function
// before we can generate constraints.
type MacroFunction = io.Function[macro.Instruction]

// MicroFunction is a function whose instructions are themselves micro
// instructions.  A micro function represents the lowest representation of a
// function, where each instruction is made up of microcodes.
type MicroFunction = io.Function[micro.Instruction]

// MacroProgram represents a set of components at the macro level.
type MacroProgram = io.Program[macro.Instruction]

// MicroProgram represents a set of components at the micro level.
type MicroProgram = io.Program[micro.Instruction]

// MacroHirProgram represents a mixed assembly and legacy program, where
// assembly functions are composed from macro instructions.
type MacroHirProgram = MixedProgram[word.BigEndian, macro.Instruction, hir.Module]

// MicroHirProgram represents a mixed assembly and legacy program, where
// assembly functions are composed from micro instructions.
type MicroHirProgram = MixedProgram[word.BigEndian, micro.Instruction, hir.Module]

// MicroMirProgram represents a mixed assembly and legacy program, where
// assembly functions are composed from micro instructions.
type MicroMirProgram[F field.Element[F]] = MixedProgram[F, micro.Instruction, mir.Module[F]]

// MacroModule is an instance of schema.Module which encapsulates a MacroFunction[F].
type MacroModule[F field.Element[F]] = program.Module[F, macro.Instruction]

// MicroModule is an instance of schema.Module which encapsulates a MicroFunction[F].
type MicroModule[F field.Element[F]] = program.Module[F, micro.Instruction]

// Assemble takes a given set of assembly files, and parses them into a given
// set of functions.  This includes performing various checks on the files, such
// as type checking, etc.
func Assemble(files ...source.File) (
	MacroProgram, source.Maps[any], []source.SyntaxError) {
	//
	var (
		items      []assembler.AssemblyItem
		errors     []source.SyntaxError
		components []*MacroFunction
		srcmaps    source.Maps[any]
		visited    map[string]bool = make(map[string]bool)
	)
	// Parse each file in turn.
	for len(files) > 0 {
		var (
			asm      = files[0]
			errs     []source.SyntaxError
			included []source.File
			cs       assembler.AssemblyItem
		)
		//
		files = files[1:]
		// Parse source file
		if cs, errs = assembler.Parse(&asm); len(errs) == 0 {
			items = append(items, cs)
			// Process included source files
			included, errs = readIncludedFiles(asm, cs, visited)
			// Append any new files for processing
			files = append(files, included...)
		}
		// Include all errors
		errors = append(errors, errs...)
	}
	// Link assembly
	if len(errors) != 0 {
		return MacroProgram{}, srcmaps, errors
	}
	// Link assembly and resolve buses
	components, srcmaps, errors = assembler.Link(items...)
	// Error check
	if len(errors) != 0 {
		return MacroProgram{}, srcmaps, errors
	}
	// Well-formedness checks (assuming unlimited field width).
	errors = assembler.Validate(math.MaxUint, components, srcmaps)
	// Done
	return io.NewProgram(components), srcmaps, errors
}

func readIncludedFiles(file source.File, item assembler.AssemblyItem,
	visited map[string]bool) ([]source.File, []source.SyntaxError) {
	//
	var (
		dir    = filepath.Dir(file.Filename())
		files  []source.File
		errors []source.SyntaxError
	)
	//
	for _, include := range item.Includes {
		filename := filepath.Join(dir, *include)
		// Check filename not already parsed
		if seen, ok := visited[filename]; seen && ok {
			// file already loaded, therefore ignore.
		} else if fs, err := source.ReadFiles(filename); err == nil {
			files = append(files, fs...)
		} else {
			errors = append(errors, *item.SourceMap.SyntaxError(include, err.Error()))
		}
		// Record that we've seen this file now.
		visited[filename] = true
	}
	//
	return files, errors
}

// LowerMixedMacroProgram a mixed macro program (i.e. schema) into a mixed micro program, using
// vectorisation if desired.  Specifically, any macro modules within the schema
// are lowered into "micro" modules (i.e. those using only micro instructions).
// This does not impact any externally defined (e.g. MIR) modules in the schema.
func LowerMixedMacroProgram(vectorize bool, program MacroHirProgram) MicroHirProgram {
	var microProgram = lowerMacroProgram(vectorize, program.program)
	// Done
	return NewMixedProgram(microProgram, program.externs...)
}

func lowerMacroProgram(vectorize bool, p MacroProgram) MicroProgram {
	functions := make([]*MicroFunction, len(p.Functions()))
	//
	for i, f := range p.Functions() {
		nf := lowerFunction(vectorize, *f)
		functions[i] = &nf
	}
	// Validate generated program.  Whilst not strictly necessary, it is useful
	// from a debugging perspective.
	assembler.ValidateMicro(math.MaxUint, functions)
	//
	return io.NewProgram(functions)
}
