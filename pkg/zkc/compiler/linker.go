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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/lval"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
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
	busmap     map[string]symbol.Resolved
	components []ast.UnresolvedDeclaration
	srcmap     source.Maps[any]
	names      map[string]bool
}

// NewLinker constructs a new linker
func NewLinker() *Linker {
	return &Linker{
		srcmap:     *source.NewSourceMaps[any](),
		busmap:     make(map[string]symbol.Resolved),
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
		p.busmap[c.Name()] = symbol.Resolved{Index: uint(len(p.busmap))}
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
	case *ast.UnresolvedConstant:
		return p.linkConstant(*d)
	case *ast.UnresolvedFunction:
		return p.linkFunction(*d)
	case *ast.UnresolvedMemory:
		address, errs1 := p.linkVariableDeclarations(d.Address)
		data, errs2 := p.linkVariableDeclarations(d.Data)
		// nothing to do here
		return decl.NewMemory[symbol.Resolved](d.Name(), d.Kind, address, data, d.Contents), append(errs1, errs2...)
	default:
		panic("unknown declaration")
	}
}

func (p *Linker) linkConstant(fn ast.UnresolvedConstant) (ast.Declaration, []source.SyntaxError) {
	expr, errs1 := p.linkExpr(fn.ConstExpr)
	datatype, errs2 := p.linkType(fn.DataType)
	// FIXME: resolve data type.
	return decl.NewConstant[symbol.Resolved](fn.Name(), datatype, expr), append(errs1, errs2...)
}

func (p *Linker) linkFunction(fn ast.UnresolvedFunction) (ast.Declaration, []source.SyntaxError) {
	var (
		codes = make([]ast.Stmt, len(fn.Code))
		errs1 []source.SyntaxError
	)
	//
	for i, c := range fn.Code {
		var es []source.SyntaxError
		//
		codes[i], es = p.linkInstruction(c)
		//
		errs1 = append(errs1, es...)
	}
	//
	vars, errs2 := p.linkVariableDeclarations(fn.Variables)
	//
	return decl.NewFunction(fn.Name(), vars, codes), append(errs1, errs2...)
}

func (p *Linker) linkVariableDeclarations(decls []variable.UnresolvedDescriptor,
) ([]variable.ResolvedDescriptor, []source.SyntaxError) {
	var (
		ndecls = make([]variable.ResolvedDescriptor, len(decls))
		errors []source.SyntaxError
	)
	//
	for i, d := range decls {
		var datatype, errs = p.linkType(d.DataType)
		//
		ndecls[i] = variable.New(d.Kind, d.Name, datatype)
		//
		errors = append(errors, errs...)
	}
	//
	return ndecls, errors
}

func (p *Linker) linkInstruction(insn ast.UnresolvedInstruction) (ast.Stmt, []source.SyntaxError) {
	var (
		ninsn  ast.Stmt
		errors []source.SyntaxError
	)
	//
	switch insn := insn.(type) {
	case *stmt.Assign[symbol.Unresolved]:
		// Link the left-hand side
		lhs, errs1 := p.linkLVals(insn.Targets)
		// Link the right-hand side
		rhs, errs2 := p.linkExpr(insn.Source)
		//
		ninsn = &stmt.Assign[symbol.Resolved]{Targets: lhs, Source: rhs}
		//
		errors = append(errs1, errs2...)
	case *stmt.Fail[symbol.Unresolved]:
		ninsn = &stmt.Fail[symbol.Resolved]{}
	case *stmt.Goto[symbol.Unresolved]:
		ninsn = &stmt.Goto[symbol.Resolved]{Target: insn.Target}
	case *stmt.IfGoto[symbol.Unresolved]:
		var cond ast.Condition
		// link the condition
		cond, errors = p.linkCondition(insn.Cond)
		//
		ninsn = &stmt.IfGoto[symbol.Resolved]{Cond: cond, Target: insn.Target}
	case *stmt.Return[symbol.Unresolved]:
		ninsn = &stmt.Return[symbol.Resolved]{}
	default:
		panic("unknown instruction encountered")
	}
	//
	if ninsn != nil {
		p.srcmap.Copy(insn, ninsn)
	}
	//
	return ninsn, errors
}

func (p *Linker) linkLVals(lvals []ast.UnresolvedLVal) ([]ast.LVal, []source.SyntaxError) {
	var (
		llvals = make([]ast.LVal, len(lvals))
		errors []source.SyntaxError
	)
	//
	for i, lval := range lvals {
		var errs []source.SyntaxError
		//
		llvals[i], errs = p.linkLVal(lval)
		//
		errors = append(errors, errs...)
	}
	//
	return llvals, errors
}

func (p *Linker) linkLVal(lv ast.UnresolvedLVal) (ast.LVal, []source.SyntaxError) {
	switch lv := lv.(type) {
	case *lval.Variable[symbol.Unresolved]:
		return lval.NewVariable[symbol.Resolved](lv.Id), nil
	case *lval.MemAccess[symbol.Unresolved]:
		// resolve symbols in memory name
		name, errs1 := p.resolve(lv.Name, lv)
		// resolve symbols in index expression
		index, errs2 := p.linkExpr(lv.Index)
		//
		return lval.NewMemAccess(name, index), append(errs1, errs2...)
	default:
		return nil, p.srcmap.SyntaxErrors(lv, "unknown lval encountered")
	}
}

func (p *Linker) linkCondition(cond ast.UnresolvedCondition) (ast.Condition, []source.SyntaxError) {
	switch e := cond.(type) {
	case *expr.Cmp[symbol.Unresolved]:
		lhs, lerrs := p.linkExpr(e.Left)
		rhs, rerrs := p.linkExpr(e.Right)
		//
		return expr.NewCmp(e.Operator, lhs, rhs), append(lerrs, rerrs...)
	default:
		panic("unknown condition encountered")
	}
}

func (p *Linker) linkExpr(e ast.UnresolvedExpr) (ast.Expr, []source.SyntaxError) {
	var (
		args   []ast.Expr
		errors []source.SyntaxError
		nexpr  ast.Expr
	)
	//
	switch e := e.(type) {
	case *expr.Add[symbol.Unresolved]:
		args, errors = p.linkExprs(e.Exprs...)
		nexpr = expr.NewAdd[symbol.Resolved](args...)
	case *expr.Const[symbol.Unresolved]:
		nexpr = expr.NewConstant[symbol.Resolved](e.Constant, e.Base)
	case *expr.ExternAccess[symbol.Unresolved]:
		// resolve arguments
		args, errors = p.linkExprs(e.Args...)
		// attempt to resolve this non-local access
		sym, errs := p.resolve(e.Name, e)
		// combine errors
		errors = append(errors, errs...)
		//
		nexpr = expr.NewExternAccess(sym, args...)
	case *expr.Mul[symbol.Unresolved]:
		args, errors = p.linkExprs(e.Exprs...)
		nexpr = expr.NewMul[symbol.Resolved](args...)
	case *expr.LocalAccess[symbol.Unresolved]:
		nexpr = expr.NewLocalAccess[symbol.Resolved](e.Variable)
	case *expr.Sub[symbol.Unresolved]:
		args, errors = p.linkExprs(e.Exprs...)
		nexpr = expr.NewSub[symbol.Resolved](args...)
	default:
		panic("unknown expression encountered")
	}
	//
	if nexpr != nil {
		p.srcmap.Copy(e, nexpr)
	}
	//
	return nexpr, errors
}

func (p *Linker) linkExprs(exprs ...ast.UnresolvedExpr) ([]ast.Expr, []source.SyntaxError) {
	var (
		nexprs []ast.Expr = make([]ast.Expr, len(exprs))
		errors []source.SyntaxError
	)
	//
	for i, e := range exprs {
		ne, errs := p.linkExpr(e)
		nexprs[i] = ne
		//
		errors = append(errors, errs...)
	}
	//
	return nexprs, errors
}

func (p *Linker) linkType(datatype data.UnresolvedType) (data.ResolvedType, []source.SyntaxError) {
	switch t := datatype.(type) {
	case *data.UnsignedInt[symbol.Unresolved]:
		return data.NewUnsignedInt[symbol.Resolved](t.Width(), t.IsOpen()), nil
	default:
		return nil, p.srcmap.SyntaxErrors(datatype, "unknown type encountered")
	}
}

// Resolve the symbol referred to by an external access into a resolved symbol,
// or return an error if there is some issue matching the symbol.
func (p *Linker) resolve(name symbol.Unresolved, node any) (symbol.Resolved, []source.SyntaxError) {
	var sym symbol.Resolved
	//
	for i, c := range p.components {
		nIns, _ := c.Arity()
		// first, check whether name matches
		if c.Name() == name.Name {
			// now, check arity
			if nIns != name.Inputs {
				return sym, p.srcmap.SyntaxErrors(node, fmt.Sprintf("incorrect number of arguments (expected %d)", nIns))
			} else if msg, err := checkSymbolKind(c, name); err {
				return sym, p.srcmap.SyntaxErrors(node, msg)
			}
			// hit
			return symbol.NewResolved(c.Name(), name.Kind, uint(i)), nil
		}
	}
	// fail
	return sym, p.srcmap.SyntaxErrors(node, "unknown symbol")
}

// Attempt to determine whether or not the given symbol kind matches the
// declaration.
func checkSymbolKind(decl ast.UnresolvedDeclaration, sym symbol.Unresolved) (msg string, err bool) {
	var nIns, _ = decl.Arity()
	//
	switch sym.Kind {
	case symbol.READABLE_MEMORY:
		if mem, ok := decl.(*ast.UnresolvedMemory); ok && mem.IsReadable() {
			return "", false
		}
		//
		return "invalid memory read", true
	case symbol.WRITEABLE_MEMORY:
		if mem, ok := decl.(*ast.UnresolvedMemory); ok && mem.IsWriteable() {
			return "", false
		}
		//
		return "invalid memory write", true
	case symbol.FUNCTION:
	case symbol.CONSTANT:
	}
	// Final arity check
	if nIns != sym.Inputs {
		return fmt.Sprintf("incorrect number of arguments (expected %d)", nIns), true
	}
	// No mismatch
	return "", false
}
