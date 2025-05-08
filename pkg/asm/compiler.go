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
	"github.com/consensys/go-corset/pkg/asm/compiler"
	"github.com/consensys/go-corset/pkg/asm/micro"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/util/source"
)

// CompileAssembly compiles a given set of assembly functions into a binary
// constraint file.
func CompileAssembly(cfg LoweringConfig, assembly ...source.File) (*binfile.BinaryFile, []source.SyntaxError) {
	macroProgram, _, errs := Assemble(assembly...)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Lower macro program into a binary program.
	microProgram := macroProgram.Lower(cfg)
	//
	return Compile(&microProgram), nil
}

// Compile a microprogram into a binary constraint file.
func Compile(microProgram Program[micro.Instruction]) *binfile.BinaryFile {
	compiler := compiler.NewCompiler()
	//
	for i := range microProgram.Functions() {
		fn := microProgram.Function(uint(i))
		compiler.Compile(fn.Name, fn.Registers, fn.Code)
	}

	return binfile.NewBinaryFile(nil, nil, compiler.Schema())
}
