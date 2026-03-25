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
package compiler

import (
	"path/filepath"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/parser"
	"github.com/consensys/go-corset/pkg/zkc/compiler/validate"
)

// Compile takes a given set of source files, and parses them into a given set
// of (linked) declarations.  This includes performing various checks on the
// files, such as type checking, etc.
func Compile(files ...source.File) (ast.Program, source.Maps[any], []source.SyntaxError) {
	//
	var (
		items   []parser.UnlinkedSourceFile
		errors  []source.SyntaxError
		program ast.Program
		srcmaps source.Maps[any]
		visited map[string]bool = make(map[string]bool)
	)
	// Initialise visited map with all top-level files
	for _, sf := range files {
		visited[sf.Filename()] = true
	}
	// Parse each file in turn.
	for len(files) > 0 {
		var (
			asm      = files[0]
			errs     []source.SyntaxError
			included []source.File
			cs       parser.UnlinkedSourceFile
		)
		//
		files = files[1:]
		// Parse source file
		if cs, errs = parser.Parse(&asm); len(errs) == 0 {
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
		return ast.Program{}, srcmaps, errors
	}
	// Link assembly and resolve buses
	program, srcmaps, errors = Link(items...)
	// Error check
	if len(errors) != 0 {
		return ast.Program{}, srcmaps, errors
	}
	// Well-formedness checks (assuming unlimited field width).
	errors = validateProgram(program, srcmaps)
	// Done
	return program, srcmaps, errors
}

func readIncludedFiles(file source.File, item parser.UnlinkedSourceFile,
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

// Validate checks that a given program is well-formed.  For example, an
// assignment "x,y = z" must be balanced (i.e. number of bits on lhs must match
// number on rhs).  Likewise, variables cannot be used before they are defined,
// and all control-flow paths must reach a "return" instruction, etc. Finally,
// we cannot assign to an input register under the current calling convention.
func validateProgram(program ast.Program, srcmaps source.Maps[any]) []source.SyntaxError {
	var (
		errors []source.SyntaxError
	)

	// Check for cyclic definitions (constants and type aliases)
	errors = append(errors, validate.CycleDetection(program, srcmaps)...)
	// If a cycle is detected, we skip the typing and control flow phase
	if len(errors) > 0 {
		return errors
	}
	// Apply various checks
	errors = append(errors, validate.Typing(program, srcmaps)...)
	errors = append(errors, validate.ControlFlow(program, srcmaps)...)
	// Done
	return errors
}
