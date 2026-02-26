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
package parser

import (
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/lex"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// UnlinkedSourceFile captures a source file has been successfully parsed but
// which has not yet been linked.   As such, its possible that such a file may
// fail with an error at link time due to an unresolvable reference to an
// external component (e.g. function, RAM, ROM, etc).
type UnlinkedSourceFile struct {
	Includes []*string
	// Components making up this assembly item.
	Components []ast.UnresolvedDeclaration
	// Mapping of instructions back to the source file.
	SourceMap source.Map[any]
}

// Parse accepts a given source file representing an assembly language
// program, and assembles it into an instruction sequence which can then the
// executed.
func Parse(srcfile *source.File) (UnlinkedSourceFile, []source.SyntaxError) {
	parser := NewParser(srcfile)
	// Parse functions
	return parser.Parse()
}

// BINOPS captures the set of binary operations
var BINOPS = []uint{SUB, MUL, ADD}

// ============================================================================
// Assembler
// ============================================================================

// Parser is a parser for assembly language.
type Parser struct {
	srcfile *source.File
	tokens  []lex.Token
	// Source mapping
	srcmap *source.Map[any]
	// Position within the tokens
	index int
}

// NewParser constructs a new parser for a given source file.
func NewParser(srcfile *source.File) *Parser {
	// Construct (initially empty) source mapping
	srcmap := source.NewSourceMap[any](*srcfile)
	//
	return &Parser{srcfile, nil, srcmap, 0}
}

// Parse the given source file into a sequence of zero or more components and/or
// some number of syntax errors.
func (p *Parser) Parse() (UnlinkedSourceFile, []source.SyntaxError) {
	var (
		item      UnlinkedSourceFile
		include   *string
		errors    []source.SyntaxError
		component ast.UnresolvedDeclaration
	)
	// Convert source file into tokens
	if p.tokens, errors = Lex(*p.srcfile); len(errors) > 0 {
		return item, errors
	}
	// Continue going until all consumed
	for p.lookahead().Kind != END_OF {
		lookahead := p.lookahead()
		// Determine type of declaration
		switch lookahead.Kind {
		case KEYWORD_CONST:
			component, errors = p.parseConstant()
		case KEYWORD_INCLUDE:
			include, errors = p.parseInclude()
			if len(errors) == 0 {
				item.Includes = append(item.Includes, include)
			}
			// Avoid appending to components
			continue
		case KEYWORD_FN:
			component, errors = p.parseFunction()
		case KEYWORD_PUBLIC, KEYWORD_PRIVATE:
			component, errors = p.parseInputOutputMemory()
		case KEYWORD_VAR:
			component, errors = p.parseReadWriteMemory()
		case KEYWORD_INPUT, KEYWORD_OUTPUT, KEYWORD_STATIC:
			errors = p.syntaxErrors(lookahead, "requires public or private")
		default:
			errors = p.syntaxErrors(lookahead, "unknown declaration")
		}
		//
		if len(errors) > 0 {
			return item, errors
		}
		//
		item.Components = append(item.Components, component)
	}
	// Copy over source map
	item.SourceMap = *p.srcmap
	//
	return item, nil
}

func (p *Parser) parseConstant() (ast.UnresolvedDeclaration, []source.SyntaxError) {
	var (
		start     = p.index
		errs      []source.SyntaxError
		lookahead lex.Token
		name      string
	)
	// Parse include declaration
	if _, errs := p.expect(KEYWORD_CONST); len(errs) > 0 {
		return nil, errs
	} else if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	} else if _, errs = p.expect(EQUALS); len(errs) > 0 {
		return nil, errs
	} else if lookahead, errs = p.expect(NUMBER); len(errs) > 0 {
		return nil, errs
	}
	// Save for source map
	end := p.index
	// So far, so good.
	val, errs := p.number(lookahead)
	base := p.baserOfNumber(lookahead)
	//
	component := decl.NewConstant[ast.UnresolvedSymbol](name, val, base)
	//
	p.srcmap.Put(component, p.spanOf(start, end-1))
	//
	return component, errs
}

func (p *Parser) parseInclude() (*string, []source.SyntaxError) {
	// Parse include declaration
	if _, errs := p.expect(KEYWORD_INCLUDE); len(errs) > 0 {
		return nil, errs
	}
	//
	tok, errs := p.expect(STRING)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Process string
	str := p.string(tok)
	str = str[1 : len(str)-1]
	pStr := &str
	// Store for error reporting.
	p.srcmap.Put(pStr, tok.Span)
	// Done
	return pStr, errs
}

func (p *Parser) parseFunction() (ast.UnresolvedDeclaration, []source.SyntaxError) {
	var (
		start    = p.index
		env      Environment
		name     string
		code     []ast.UnresolvedInstruction
		errs     []source.SyntaxError
		returned bool
	)
	// Parse function declaration
	if _, errs := p.expect(KEYWORD_FN); len(errs) > 0 {
		return nil, errs
	}
	// Parse function name
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	// Parse inputs
	if errs = p.parseArgsList(variable.PARAMETER, &env); len(errs) > 0 {
		return nil, errs
	}
	// Parse optional '->'
	if p.match(RIGHTARROW) {
		// Parse returns
		if errs = p.parseArgsList(variable.RETURN, &env); len(errs) > 0 {
			return nil, errs
		}
	}
	// Save for source map
	end := p.index
	// Parse start of block
	if returned, code, errs = p.parseStatementBlock(0, &env); len(errs) > 0 {
		return nil, errs
	}
	// Sanity check we parsed something
	if !returned {
		return nil, p.syntaxErrors(p.lookahead(), "missing return")
	}
	// Advance past "}"
	p.match(RCURLY)
	// Construct function
	fn := decl.NewFunction(name, env.variables, code)
	//
	p.srcmap.Put(fn, p.spanOf(start, end-1))
	// Done
	return fn, nil
}

func (p *Parser) parseArgsList(kind variable.Kind, env *Environment) []source.SyntaxError {
	var (
		arg      string
		datatype data.Type
		errs     []source.SyntaxError
		first    = true
	)
	// Parse start of list
	if _, errs = p.expect(LBRACE); len(errs) > 0 {
		return errs
	}
	// Parse entries until end brace
	for p.lookahead().Kind != RBRACE {
		// look for ","
		if !first {
			if _, errs = p.expect(COMMA); len(errs) > 0 {
				return errs
			}
		}
		//
		first = false
		// save lookahead token for syntax errors
		lookahead := p.lookahead()
		// parse name, type & optional padding
		if arg, errs = p.parseIdentifier(); len(errs) > 0 {
			return errs
		} else if datatype, errs = p.parseType(); len(errs) > 0 {
			return errs
		} else if env.IsVariable(arg) {
			return p.syntaxErrors(lookahead, "variable already declared")
		}
		//
		env.DeclareVariable(kind, arg, datatype)
	}
	// Advance past "}"
	p.match(RBRACE)
	//
	return nil
}

func (p *Parser) parseInputOutputMemory() (ast.UnresolvedDeclaration, []source.SyntaxError) {
	var (
		public  bool
		input   bool
		name    string
		errs    []source.SyntaxError
		address data.Type
		data    data.Type
	)
	// Parse pub modifier (if present)
	if p.match(KEYWORD_PUBLIC) {
		public = true
	} else if _, errs := p.expect(KEYWORD_PRIVATE); len(errs) > 0 {
		return nil, errs
	}
	//
	lookahead := p.lookahead()
	// Determine type of declaration
	switch lookahead.Kind {
	case KEYWORD_INPUT:
		p.match(KEYWORD_INPUT)
		//
		input = true
	case KEYWORD_OUTPUT:
		p.match(KEYWORD_OUTPUT)
		//
		input = false
	default:
		return nil, p.syntaxErrors(lookahead, "unknown declaration")
	}
	// Parse memory name
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	// Parse memory type
	if address, data, errs = p.parseMemoryType(false); len(errs) > 0 {
		return nil, errs
	}
	// Done
	if input {
		return decl.NewReadOnlyMemory[ast.UnresolvedSymbol](public, name, address, data), nil
	}
	//
	return decl.NewWriteOnceMemory[ast.UnresolvedSymbol](public, name, address, data), nil
}

func (p *Parser) parseReadWriteMemory() (ast.UnresolvedDeclaration, []source.SyntaxError) {
	var (
		name    string
		errs    []source.SyntaxError
		address data.Type
		data    data.Type
	)
	//
	if _, errs := p.expect(KEYWORD_VAR); len(errs) > 0 {
		return nil, errs
	}
	// Parse memory name
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	// Parse memory type
	if address, data, errs = p.parseMemoryType(true); len(errs) > 0 {
		return nil, errs
	}
	// Done
	return decl.NewRandomAccessMemory[ast.UnresolvedSymbol](name, address, data), nil
}

func (p *Parser) parseMemoryType(ram bool) (data.Type, data.Type, []source.SyntaxError) {
	var (
		addressBus data.Type
		dataBus    data.Type
		errs       []source.SyntaxError
	)
	// Parse entries until end brace
	if ram {
		if addressBus, errs = p.parseSeparatedTypeList(LSQUARE, RSQUARE); len(errs) > 0 {
			return nil, nil, errs
		}
	} else if addressBus, errs = p.parseTypeList(LSQUARE, RSQUARE); len(errs) > 0 {
		return nil, nil, errs
	}
	// Check for type list or not
	if p.lookahead().Kind == LBRACE {
		if dataBus, errs = p.parseTypeList(LBRACE, RBRACE); len(errs) > 0 {
			return nil, nil, errs
		}
	} else {
		if dataBus, errs = p.parseType(); len(errs) > 0 {
			return nil, nil, errs
		}
	}
	//
	return addressBus, dataBus, nil
}

func (p *Parser) parseSeparatedTypeList(lBrace, rBrace uint) (data.Type, []source.SyntaxError) {
	var (
		types []data.Type
		errs  []source.SyntaxError
	)
	// Parse start of list
	if _, errs = p.expect(lBrace); len(errs) > 0 {
		return nil, errs
	}
	// Keep going until end of list
	// Parse entries until end brace
	for p.lookahead().Kind != rBrace {
		var next data.Type
		// look for ","
		switch {
		case len(types) == 1:
			if _, errs = p.expect(SEMICOLON); len(errs) > 0 {
				return nil, errs
			}
		case len(types) > 1:
			if _, errs = p.expect(COMMA); len(errs) > 0 {
				return nil, errs
			}
		}
		//
		if next, errs = p.parseType(); len(errs) > 0 {
			return nil, errs
		}
		//
		types = append(types, next)
	}
	//
	p.match(rBrace)
	//
	if len(types) == 1 {
		return types[0], nil
	}
	//
	return data.NewTuple(types...), nil
}

func (p *Parser) parseTypeList(lBrace, rBrace uint) (data.Type, []source.SyntaxError) {
	var (
		types []data.Type
		errs  []source.SyntaxError
	)
	// Parse start of list
	if _, errs = p.expect(lBrace); len(errs) > 0 {
		return nil, errs
	}
	// Keep going until end of list
	// Parse entries until end brace
	for p.lookahead().Kind != rBrace {
		var next data.Type
		// look for ","
		if len(types) > 0 {
			if _, errs = p.expect(COMMA); len(errs) > 0 {
				return nil, errs
			}
		}
		//
		if next, errs = p.parseType(); len(errs) > 0 {
			return nil, errs
		}
		//
		types = append(types, next)
	}
	//
	p.match(rBrace)
	//
	if len(types) == 1 {
		return types[0], nil
	}
	//
	return data.NewTuple(types...), nil
}

func (p *Parser) parseType() (data.Type, []source.SyntaxError) {
	var (
		lookahead = p.lookahead()
		name      string
		errs      []source.SyntaxError
	)
	//
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	//
	switch {
	case strings.HasPrefix(name, "u"):
		// Parse bitwidth
		bw, err := strconv.Atoi(name[1:])
		//
		if err != nil {
			return nil, p.syntaxErrors(lookahead, err.Error())
		}
		//
		return data.NewUnsignedInt(uint(bw)), nil
	default:
		return nil, p.syntaxErrors(lookahead, "unknown type")
	}
}

func (p *Parser) parseStatementBlock(pc uint, env *Environment,
) (bool, []ast.UnresolvedInstruction, []source.SyntaxError) {
	//
	var (
		errs     []source.SyntaxError
		insns    []ast.UnresolvedInstruction
		returned bool
	)
	// Parse start of block
	if _, errs = p.expect(LCURLY); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse instructions until end of block
	for p.lookahead().Kind != RCURLY {
		var (
			ith []ast.UnresolvedInstruction
			ret bool
		)
		//
		if ret, ith, errs = p.parseStatement(pc, env); len(errs) > 0 {
			return false, nil, errs
		}
		//
		returned = returned || ret
		//
		insns = append(insns, ith...)
		// increment pc
		pc = pc + uint(len(ith))
	}
	// Advance past "}"
	p.match(RCURLY)
	//
	return returned, insns, errs
}

func (p *Parser) parseStatement(pc uint, env *Environment) (bool, []ast.UnresolvedInstruction, []source.SyntaxError) {
	var (
		// Save current position for backtracking
		start    = p.index
		errs     []source.SyntaxError
		insns    []ast.UnresolvedInstruction
		insn     ast.UnresolvedInstruction
		returned bool
	)
	//
	lookahead := p.lookahead()
	//
	switch lookahead.Kind {
	case KEYWORD_FAIL:
		returned, insn, errs = p.parseFail(env)
	case KEYWORD_IF:
		returned, insns, errs = p.parseIfElse(pc, env)
	case KEYWORD_FOR:
		returned, insns, errs = p.parseFor(pc, env)
	case KEYWORD_WHILE:
		returned, insns, errs = p.parseWhile(pc, env)
	case KEYWORD_RETURN:
		returned, insn, errs = p.parseReturn(env)
	case KEYWORD_VAR:
		insns, errs = p.parseVar(env)
	default:
		// parse assignment
		insn, errs = p.parseAssignment(env)
	}
	// Include unit instruction (if applicable)
	if insn != nil {
		insns = append(insns, insn)
	}
	// Record source mapping
	for _, insn := range insns {
		// Check whether instruction already added to source map.  This can
		// arise with recursive calls to parseStatement() (e.g. for blocks).
		if !p.srcmap.Has(insn) {
			p.srcmap.Put(insn, p.spanOf(start, p.index-1))
		}
	}
	//
	return returned, insns, errs
}

func (p *Parser) parseAssignment(env *Environment) (ast.UnresolvedInstruction, []source.SyntaxError) {
	var (
		lhs  []variable.Id
		rhs  expr.Expr
		errs []source.SyntaxError
	)
	// parse left-hand side
	if lhs, errs = p.parseVariableList(env); len(errs) > 0 {
		return nil, errs
	}
	// Reverse items so that least significant comes first.
	lhs = array.Reverse(lhs)
	// Parse '='
	if _, errs = p.expect(EQUALS); len(errs) > 0 {
		return nil, errs
	}
	// Parse right-hand side
	if rhs, errs = p.parseExpr(env); len(errs) > 0 {
		return nil, errs
	}
	// Done
	return &stmt.Assign[ast.UnresolvedSymbol]{Targets: lhs, Source: rhs}, nil
}

func (p *Parser) parseIfElse(pc uint, env *Environment) (bool, []ast.UnresolvedInstruction, []source.SyntaxError) {
	var (
		errs              []source.SyntaxError
		cond              expr.Condition
		insns             = []ast.UnresolvedInstruction{nil}
		trueBranch        []ast.UnresolvedInstruction
		falseBranch       []ast.UnresolvedInstruction
		trueRet, falseRet bool
	)
	// Match if
	if _, errs := p.expect(KEYWORD_IF); len(errs) > 0 {
		return false, nil, errs
	}
	// save lookahead for error reporting
	if cond, errs = p.parseCondition(env); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse true branch
	if trueRet, trueBranch, errs = p.parseStatementBlock(pc+1, env); len(errs) > 0 {
		return false, nil, errs
	}
	// falseTarget for if-goto
	falseTarget := pc + 1 + uint(len(trueBranch))
	// Include the true branch
	insns = append(insns, trueBranch...)
	// Check for "else"
	if p.lookahead().Kind == KEYWORD_ELSE {
		// Skip over if
		_, _ = p.expect(KEYWORD_ELSE)
		// add branch bypass (if needed)
		if !trueRet {
			// update targets
			falseTarget++
			endTarget := falseTarget + uint(len(falseBranch))
			insns = append(insns, &stmt.Goto[ast.UnresolvedSymbol]{Target: endTarget})
		}
		// parse false branch
		if falseRet, falseBranch, errs = p.parseStatementBlock(falseTarget, env); len(errs) > 0 {
			return false, nil, errs
		}
		// add false branch (if applicable)
		//
		insns = append(insns, falseBranch...)
	}
	// Configure initial if-goto
	insns[0] = &stmt.IfGoto[ast.UnresolvedSymbol]{
		Cond: cond.Negate(), Target: falseTarget}

	// Done
	return trueRet && falseRet, insns, nil
}

func (p *Parser) parseWhile(pc uint, env *Environment) (bool, []ast.UnresolvedInstruction, []source.SyntaxError) {
	var (
		errs  []source.SyntaxError
		cond  expr.Condition
		insns = []ast.UnresolvedInstruction{nil}
		body  []ast.UnresolvedInstruction
	)
	// Match while
	if _, errs = p.expect(KEYWORD_WHILE); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse condition
	if cond, errs = p.parseCondition(env); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse body block; body starts at pc+1
	if _, body, errs = p.parseStatementBlock(pc+1, env); len(errs) > 0 {
		return false, nil, errs
	}
	// Back-goto jumps to the if-goto at pc
	insns = append(insns, body...)
	insns = append(insns, &stmt.Goto[ast.UnresolvedSymbol]{Target: pc})
	// The conditional skip jumps past the back-goto to the instruction after the loop
	exitTarget := pc + uint(len(insns))
	insns[0] = &stmt.IfGoto[ast.UnresolvedSymbol]{Cond: cond.Negate(), Target: exitTarget}
	// A while loop never guarantees a return
	return false, insns, nil
}

func (p *Parser) parseFor(pc uint, env *Environment) (bool, []ast.UnresolvedInstruction, []source.SyntaxError) {
	var (
		errs []source.SyntaxError
		init ast.UnresolvedInstruction
		cond expr.Condition
		post ast.UnresolvedInstruction
		body []ast.UnresolvedInstruction
	)
	// Match 'for'
	if _, errs = p.expect(KEYWORD_FOR); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse init assignment
	if init, errs = p.parseAssignment(env); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse ';'
	if _, errs = p.expect(SEMICOLON); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse condition
	if cond, errs = p.parseCondition(env); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse ';'
	if _, errs = p.expect(SEMICOLON); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse post assignment
	if post, errs = p.parseAssignment(env); len(errs) > 0 {
		return false, nil, errs
	}
	// Layout:
	//   pc+0:                   init
	//   pc+1:                   if !cond goto exit  (placeholder)
	//   pc+2 .. pc+1+|body|:    body
	//   pc+2+|body|:            post
	//   pc+3+|body|:            goto pc+1
	//   pc+4+|body| (exit):     ...
	condPC := pc + 1
	// Parse body; starts at condPC+1 = pc+2
	if _, body, errs = p.parseStatementBlock(condPC+1, env); len(errs) > 0 {
		return false, nil, errs
	}
	// Build the instruction sequence
	insns := make([]ast.UnresolvedInstruction, 0, len(body)+4)
	insns = append(insns, init)
	insns = append(insns, nil) // placeholder for if-goto at condPC
	insns = append(insns, body...)
	insns = append(insns, post)
	insns = append(insns, &stmt.Goto[ast.UnresolvedSymbol]{Target: condPC})
	// Fill in the conditional check: exit is the instruction after the back-goto
	exitTarget := pc + uint(len(insns))
	insns[1] = &stmt.IfGoto[ast.UnresolvedSymbol]{Cond: cond.Negate(), Target: exitTarget}
	// A for loop never guarantees a return
	return false, insns, nil
}

func (p *Parser) parseReturn(env *Environment) (bool, ast.UnresolvedInstruction, []source.SyntaxError) {
	if _, errs := p.expect(KEYWORD_RETURN); len(errs) > 0 {
		return true, nil, errs
	}
	//
	return true, &stmt.Return[ast.UnresolvedSymbol]{}, nil
}

func (p *Parser) parseFail(env *Environment) (bool, ast.UnresolvedInstruction, []source.SyntaxError) {
	if _, errs := p.expect(KEYWORD_FAIL); len(errs) > 0 {
		return true, nil, errs
	}
	//
	return true, &stmt.Fail[ast.UnresolvedSymbol]{}, nil
}
func (p *Parser) parseVar(env *Environment) ([]ast.UnresolvedInstruction, []source.SyntaxError) {
	var (
		errs     []source.SyntaxError
		names    []string
		datatype data.Type
	)
	//
	if _, errs = p.expect(KEYWORD_VAR); len(errs) > 0 {
		return nil, errs
	}
	// Parse name(s)
	for len(names) == 0 || p.match(COMMA) {
		// Store lookahead for error reporting
		lookahead := p.lookahead()
		//
		name, errs := p.parseIdentifier()
		//
		if len(errs) > 0 {
			return nil, errs
		} else if env.IsVariable(name) {
			return nil, p.syntaxErrors(lookahead, "variable already declared")
		}
		//
		names = append(names, name)
	}
	// parse type
	if datatype, errs = p.parseType(); len(errs) > 0 {
		return nil, errs
	}
	// Declare all variables before parsing any initialiser, so the
	// initialiser expression can reference other already-declared variables.
	for _, name := range names {
		env.DeclareVariable(variable.LOCAL, name, datatype)
	}
	// Check for optional initialiser
	if !p.match(EQUALS) {
		return nil, nil
	}
	// Initialisers are only supported for single-variable declarations.
	if len(names) > 1 {
		return nil, p.syntaxErrors(p.lookahead(), "initialiser requires single variable declaration")
	}
	// Parse the initialiser expression
	rhs, errs := p.parseExpr(env)
	if len(errs) > 0 {
		return nil, errs
	}
	// Build the assignment instruction
	target := env.LookupVariable(names[0])
	insn := &stmt.Assign[ast.UnresolvedSymbol]{
		Targets: []variable.Id{target},
		Source:  rhs,
	}
	//
	return []ast.UnresolvedInstruction{insn}, nil
}

func (p *Parser) parseCondition(env *Environment) (expr.Condition, []source.SyntaxError) {
	var (
		errs     []source.SyntaxError
		lhs, rhs expr.Expr
		op       expr.CmpOp
	)
	// Parse left hand side
	if lhs, errs = p.parseExpr(env); len(errs) > 0 {
		return nil, errs
	}
	// save lookahead for error reporting
	if op, errs = p.parseComparator(); len(errs) > 0 {
		return nil, errs
	}
	// Parse right hand side
	if rhs, errs = p.parseExpr(env); len(errs) > 0 {
		return nil, errs
	}
	//
	return &expr.Cmp{Operator: op, Left: lhs, Right: rhs}, nil
}

func (p *Parser) parseExpr(env *Environment) (expr.Expr, []source.SyntaxError) {
	var (
		start     = p.index
		arg, errs = p.parseUnitExpr(env)
		args      = []expr.Expr{arg}
		tmp       expr.Expr
	)
	// initialise lookahead
	kind := p.lookahead().Kind
	//
	for len(errs) == 0 && p.follows(BINOPS...) {
		// Sanity check
		if !p.follows(kind) {
			return tmp, p.syntaxErrors(p.lookahead(), "braces required")
		}
		// Consume connective
		p.expect(p.lookahead().Kind)
		//
		tmp, errs = p.parseUnitExpr(env)
		// Accumulate arguments
		args = append(args, tmp)
	}
	//
	switch {
	case len(errs) != 0:
		return arg, errs
	case len(args) == 1:
		return arg, nil
	case kind == ADD:
		arg = expr.NewAdd(args...)
	case kind == MUL:
		arg = expr.NewMul(args...)
	case kind == SUB:
		arg = expr.NewSub(args...)
	}
	//
	p.srcmap.Put(arg, p.spanOf(start, p.index-1))
	//
	return arg, nil
}

func (p *Parser) parseUnitExpr(env *Environment) (expr.Expr, []source.SyntaxError) {
	var lookahead = p.lookahead()

	switch lookahead.Kind {
	case IDENTIFIER, NUMBER:
		return p.parseAtomicExpr(env)
	case LBRACE:
		p.match(LBRACE)
		expr, errs := p.parseExpr(env)
		//
		if len(errs) > 0 {
			return nil, errs
		} else if _, errs = p.expect(RBRACE); len(errs) > 0 {
			return nil, errs
		}
		// Don't add to source map, since it will already have been added.
		return expr, nil
	default:
		return nil, p.syntaxErrors(lookahead, "unexpected token")
	}
}

func (p *Parser) parseAtomicExpr(env *Environment) (expr.Expr, []source.SyntaxError) {
	var (
		start     = p.index
		lookahead = p.lookahead()
		atom      expr.Expr
		errs      []source.SyntaxError
	)

	switch lookahead.Kind {
	case IDENTIFIER:
		var reg string
		//
		reg, errs = p.parseIdentifier()
		//
		if len(errs) > 0 {
			return nil, errs
		} else if !env.IsVariable(reg) {
			atom = expr.NewConstantAccess(reg)
		} else {
			// Register access
			rid := env.LookupVariable(reg)
			// Done
			atom = expr.NewVarAccess(rid)
		}
	case NUMBER:
		var val big.Int
		//
		p.match(NUMBER)
		//
		val, errs = p.number(lookahead)
		base := p.baserOfNumber(lookahead)
		//
		atom = expr.NewConstant(val, base)
	default:
		return nil, p.syntaxErrors(lookahead, "expected register or constant")
	}
	//
	p.srcmap.Put(atom, p.spanOf(start, p.index-1))
	//
	return atom, errs
}

// Parse sequence of one or more expressions separated by a comma.
// nolint
func (p *Parser) parseExprList(env *Environment) ([]expr.Expr, []source.SyntaxError) {
	var (
		lhs  = make([]expr.Expr, 1)
		errs []source.SyntaxError
		expr expr.Expr
	)
	// lhs always starts with a register
	if lhs[0], errs = p.parseExpr(env); len(errs) > 0 {
		return nil, errs
	}
	// lhs may have additional registers
	for p.match(COMMA) {
		if expr, errs = p.parseExpr(env); len(errs) > 0 {
			return nil, errs
		}
		// Add register to lhs
		lhs = append(lhs, expr)
	}
	//
	return lhs, nil
}

// Parse sequence of one or more registers separated by a comma.
func (p *Parser) parseVariableList(env *Environment) ([]variable.Id, []source.SyntaxError) {
	var (
		lhs  []variable.Id = make([]variable.Id, 1)
		errs []source.SyntaxError
		reg  variable.Id
	)
	// lhs always starts with a register
	if lhs[0], errs = p.parseVariable(env); len(errs) > 0 {
		return nil, errs
	}
	// lhs may have additional registers
	for p.match(COMMA) {
		if reg, errs = p.parseVariable(env); len(errs) > 0 {
			return nil, errs
		}
		// Add register to lhs
		lhs = append(lhs, reg)
	}
	//
	return lhs, nil
}

func (p *Parser) parseVariable(env *Environment) (variable.Id, []source.SyntaxError) {
	var (
		empty     uint
		lookahead = p.lookahead()
		reg, errs = p.parseIdentifier()
	)
	//
	if len(errs) > 0 {
		return empty, errs
	} else if !env.IsVariable(reg) {
		return empty, p.syntaxErrors(lookahead, "unknown register")
	}
	// Done
	return env.LookupVariable(reg), nil
}

func (p *Parser) parseIdentifier() (string, []source.SyntaxError) {
	tok, errs := p.expect(IDENTIFIER)
	//
	if len(errs) > 0 {
		return "", errs
	}
	//
	return p.string(tok), nil
}

func (p *Parser) parseComparator() (expr.CmpOp, []source.SyntaxError) {
	var (
		lookahead = p.lookahead()
		op        expr.CmpOp
	)
	// Parse operation
	switch lookahead.Kind {
	case EQUALS_EQUALS:
		op = expr.EQ
	case NOT_EQUALS:
		op = expr.NEQ
	case LESS_THAN:
		op = expr.LT
	case LESS_THAN_EQUALS:
		op = expr.LTEQ
	case GREATER_THAN:
		op = expr.GT
	case GREATER_THAN_EQUALS:
		op = expr.GTEQ
	default:
		return math.MaxUint8, p.syntaxErrors(lookahead, "unknown comparator")
	}
	//
	p.match(lookahead.Kind)
	//
	return op, nil
}

// Get the text representing the given token as a string.
func (p *Parser) string(token lex.Token) string {
	start, end := token.Span.Start(), token.Span.End()
	return string(p.srcfile.Contents()[start:end])
}

// Get the text representing the given token as a string.
func (p *Parser) number(token lex.Token) (big.Int, []source.SyntaxError) {
	var (
		number, exponent big.Int
		ok               bool
		numstr           = p.string(token)
		splits           = strings.Split(numstr, "^")
	)
	//
	if len(splits) == 0 || len(splits) > 2 {
		ok = false
	} else if len(splits) == 1 {
		// non-exponent case
		_, ok = number.SetString(numstr, 0)
	} else if len(splits[0]) == 0 || len(splits[1]) == 0 {
		ok = false
	} else {
		_, ok = number.SetString(splits[0], 0)
		exponent.SetString(splits[1], 0)
		number.Exp(&number, &exponent, nil)
	}
	//
	if !ok {
		return number, p.syntaxErrors(token, "malformed numeric literal")
	}
	//
	return number, nil
}

// Get the text representing the given token as a string.
func (p *Parser) baserOfNumber(token lex.Token) uint {
	var str = p.string(token)
	//
	if strings.HasPrefix(str, "0x") {
		return 16
	} else if strings.HasPrefix(str, "0b") {
		return 2
	}
	//
	return 10
}

// Lookahead returns the next token.  This must exist because EOF is always
// appended at the end of the token stream.
func (p *Parser) lookahead() lex.Token {
	return p.tokens[p.index]
}

// Expect reurns an arror if the next token is not what was expected.
func (p *Parser) expect(kind uint) (lex.Token, []source.SyntaxError) {
	lookahead := p.lookahead()
	//
	if lookahead.Kind != kind {
		errs := p.syntaxErrors(lookahead, "unexpected token")
		return lookahead, errs
	}
	//
	p.index++
	//
	return lookahead, nil
}

// Match attempts to match the given token.
func (p *Parser) match(kind uint) bool {
	if p.lookahead().Kind == kind {
		p.index++
		return true
	}
	//
	return false
}

// Follows checks whether one of the given token kinds is next.
func (p *Parser) follows(options ...uint) bool {
	return slices.Contains(options, p.lookahead().Kind)
}

func (p *Parser) spanOf(firstToken, lastToken int) source.Span {
	//
	start := p.tokens[firstToken].Span.Start()
	end := p.tokens[lastToken].Span.End()
	//
	return source.NewSpan(start, end)
}

func (p *Parser) syntaxErrors(token lex.Token, msg string) []source.SyntaxError {
	return []source.SyntaxError{*p.srcfile.SyntaxError(token.Span, msg)}
}

// Label represents a potentially unresolved label within an assembly.
type Label struct {
	// Name of the label
	name string
	// PC position the label represents.  This will be math.MaxUint until the
	// label is officially declared.
	pc uint
}

// UnboundLabel constructs a label whose PC location is (as yet) unknown.
func UnboundLabel(name string) Label {
	return Label{name, math.MaxUint}
}

// BoundLabel constructs a label whose PC location is known.
func BoundLabel(name string, pc uint) Label {
	return Label{name, pc}
}
