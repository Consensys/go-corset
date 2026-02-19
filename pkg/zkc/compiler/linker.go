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
	"fmt"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/parser"
)

// Link a set of one or more source files together to produce a complete program
// (or one or more errors).  Linking is the process of resolving external
// identifiers used within a source file, or generateing errors when this fails.
// For example, if a function in one source file calls another function in a
// different source file, then this linkage needs to be resolved (i.e. checked).
// This can fail for various reasons: for example, if no function of the given
// name can be found in any source file; or, if a function with the correct name
// but incorrect arity (i.e. number of parameters/returns) is found.
func Link(files ...parser.UnlinkedSourceFile) (ast.Program, source.Maps[any], []source.SyntaxError) {
	var (
		program ast.Program
		linker  = NewLinker()
		errors  []source.SyntaxError
	)
	// Constuct bus and source mappings
	for _, item := range files {
		linker.Join(item.SourceMap)
		//
		for _, declaration := range item.Components {
			// Check whether component of same name already exists.
			if linker.Exists(declaration.Name()) {
				// Indicates component of same name already exists.  It would be
				// good to report a source error here, but the problem is that
				// our source map doesn't contain the right information.
				msg := fmt.Sprintf("duplicate declaration %s", declaration.Name())
				errors = append(errors, *linker.srcmap.SyntaxError(declaration, msg))
			} else {
				linker.Register(declaration)
			}
		}
	}
	// Link all assembly items
	if len(errors) == 0 {
		program, errors = linker.Link()
	}
	//
	return program, linker.srcmap, errors
}

// Linker packages together the various bits of information required for linking
// the assembly files.
type Linker struct {
	busmap     map[string]ast.ResolvedSymbol
	components []ast.UnresolvedDeclaration
	srcmap     source.Maps[any]
	names      map[string]bool
}

// NewLinker constructs a new linker
func NewLinker() *Linker {
	return &Linker{
		srcmap:     *source.NewSourceMaps[any](),
		busmap:     make(map[string]ast.ResolvedSymbol),
		components: nil,
		names:      make(map[string]bool),
	}
}

// Exists checks whether or not a component of the given name already exists.
func (p *Linker) Exists(name string) bool {
	_, ok := p.names[name]
	//
	return ok
}

// Join a source map into this linker
func (p *Linker) Join(srcmap source.Map[any]) {
	p.srcmap.Join(&srcmap)
}

// Register a new components with this linker.
func (p *Linker) Register(component ast.UnresolvedDeclaration) {
	// First, record name
	p.names[component.Name()] = true
	// Second, act on component type
	switch c := component.(type) {
	case ast.UnresolvedDeclaration:
		// Allocate bus entry
		p.busmap[c.Name()] = ast.ResolvedSymbol{Index: uint(len(p.busmap))}
		//
		p.components = append(p.components, c)
	default:
		// Should be unreachable
		panic(fmt.Sprintf("unknown component %s", component.Name()))
	}
}

// Link all components register with this linker
func (p *Linker) Link() (ast.Program, []source.SyntaxError) {
	var (
		errors []source.SyntaxError
		decls  []ast.Declaration
	)
	//
	for index := range p.components {
		decl, errs := p.linkDeclaration(uint(index))
		if len(errs) == 0 {
			decls = append(decls, decl)
		}
		//
		errors = append(errors, errs...)
		//
		p.srcmap.Copy(p.components[index], decl)
	}
	//
	return ast.NewProgram(decls), errors
}

// Link all buses used within this function to their intended targets.  This
// means, for every bus used locally, settings the global bus identifier and
// also allocated regisers for the address/data lines.
func (p *Linker) linkDeclaration(index uint) (ast.Declaration, []source.SyntaxError) {
	switch d := p.components[index].(type) {
	case *ast.UnresolvedFunction:
		return p.linkFunction(*d)
	case *ast.UnresolvedMemory:
		// nothing to do here
		return decl.NewMemory[ast.ResolvedSymbol](d.Name(), d.Kind, d.Address, d.Data, d.Contents), nil
	default:
		panic("unknown declaration")
	}
}

func (p *Linker) linkFunction(fn ast.UnresolvedFunction) (ast.Declaration, []source.SyntaxError) {
	var (
		codes = make([]ast.Instruction, len(fn.Code))
		errs  []source.SyntaxError
	)
	//
	for i, c := range fn.Code {
		var es []source.SyntaxError
		//
		codes[i], es = p.linkInstruction(c)
		//
		errs = append(errs, es...)
	}
	//
	return decl.NewFunction(fn.Name(), fn.Variables, codes), errs
}

func (p *Linker) linkInstruction(insn ast.UnresolvedInstruction) (ast.Instruction, []source.SyntaxError) {
	var ninsn ast.Instruction
	//
	switch insn := insn.(type) {
	case *stmt.Assign[ast.UnresolvedSymbol]:
		// FIXME: need to link LHS & RHS
		ninsn = &stmt.Assign[ast.ResolvedSymbol]{Targets: insn.Targets, Source: insn.Source}
	case *stmt.Fail[ast.UnresolvedSymbol]:
		ninsn = &stmt.Fail[ast.ResolvedSymbol]{}
	case *stmt.Goto[ast.UnresolvedSymbol]:
		ninsn = &stmt.Goto[ast.ResolvedSymbol]{Target: insn.Target}
	case *stmt.IfGoto[ast.UnresolvedSymbol]:
		// FIXME: need to link condition
		ninsn = &stmt.IfGoto[ast.ResolvedSymbol]{Cond: insn.Cond, Target: insn.Target}
	case *stmt.Return[ast.UnresolvedSymbol]:
		ninsn = &stmt.Return[ast.ResolvedSymbol]{}
	default:
		panic("unknown instruction encountered")
	}
	//
	if ninsn != nil {
		p.srcmap.Copy(insn, ninsn)
	}
	//
	return ninsn, nil
}
