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
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/lex"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/lval"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	zkc_util "github.com/consensys/go-corset/pkg/zkc/util"
)

// Condition is a convenient alias
type Condition = expr.Condition[symbol.Unresolved]

// Expr is a convenient alias
type Expr = expr.Expr[symbol.Unresolved]

// LVal is a convenient alias
type LVal = lval.LVal[symbol.Unresolved]

// Type is a convenient alias
type Type = data.Type[symbol.Unresolved]

// VariableDescriptor is a convenient alias
type VariableDescriptor = variable.Descriptor[symbol.Unresolved]

// IfGoto is a convenient alias
type IfGoto = stmt.IfGoto[symbol.Unresolved]

// Goto is a convenient alias
type Goto = stmt.Goto[symbol.Unresolved]

// UnlinkedSourceFile captures a source file has been successfully parsed but
// which has not yet been linked.   As such, its possible that such a file may
// fail with an error at link time due to an unresolvable reference to an
// external component (e.g. function, RAM, ROM, etc).
type UnlinkedSourceFile struct {
	Includes []*string
	// Components making up this assembly item.
	Components []decl.Unresolved
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
var BINOPS = []uint{
	SUB, MUL, ADD, DIV, REM, BITWISE_AND, BITWISE_OR, BITWISE_XOR, BITWISE_SHL,
	BITWISE_SHR, EQUALS_EQUALS, NOT_EQUALS,
	LESS_THAN, LESS_THAN_EQUALS, GREATER_THAN, GREATER_THAN_EQUALS}

// LOGICAL_BINOPS captures the set of logical binary operations
var LOGICAL_BINOPS = []uint{LOGICAL_AND, LOGICAL_OR}

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
		component decl.Unresolved
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
			var consts []decl.Unresolved

			consts, errors = p.parseConstant()
			if len(errors) > 0 {
				return item, errors
			}

			item.Components = append(item.Components, consts...)
			// Avoid appending to components below
			continue
		case KEYWORD_INCLUDE:
			include, errors = p.parseInclude()
			if len(errors) == 0 {
				item.Includes = append(item.Includes, include)
			}
			// Avoid appending to components
			continue
		case KEYWORD_FN:
			component, errors = p.parseFunction()
		case KEYWORD_PUB, KEYWORD_INPUT, KEYWORD_OUTPUT, KEYWORD_STATIC:
			component, errors = p.parseInputOutputMemory()
		case KEYWORD_MEMORY:
			component, errors = p.parseReadWriteMemory()
		case KEYWORD_TYPE:
			component, errors = p.parseTypeAlias()
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

func (p *Parser) parseConstant() ([]decl.Unresolved, []source.SyntaxError) {
	var (
		start    = p.index
		errs     []source.SyntaxError
		datatype Type
		name     string
		env      = EmptyEnvironment()
		consts   []decl.Unresolved
	)
	// Parse const declaration
	if _, errs := p.expect(KEYWORD_CONST); len(errs) > 0 {
		return nil, errs
	} else if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	} else if _, errs = p.expect(COLON); len(errs) > 0 {
		return nil, errs
	} else if datatype, errs = p.parseType(); len(errs) > 0 {
		return nil, errs
	} else if _, errs = p.expect(EQUALS); len(errs) > 0 {
		return nil, errs
	}
	// Save for source map
	end := p.index
	// So far, so good.
	constExpr, errs := p.parseTernaryOrExpr(env)
	if len(errs) > 0 {
		return nil, errs
	}
	//
	component := decl.NewConstant[symbol.Unresolved](name, datatype, constExpr)
	p.srcmap.Put(component, p.spanOf(start, end-1))
	consts = append(consts, component)
	// Parse additional comma-separated constants on the same line.
	for p.match(COMMA) {
		start = p.index
		if name, errs = p.parseIdentifier(); len(errs) > 0 {
			return nil, errs
		} else if _, errs = p.expect(COLON); len(errs) > 0 {
			return nil, errs
		} else if datatype, errs = p.parseType(); len(errs) > 0 {
			return nil, errs
		} else if _, errs = p.expect(EQUALS); len(errs) > 0 {
			return nil, errs
		}

		end = p.index

		constExpr, errs = p.parseTernaryOrExpr(env)
		if len(errs) > 0 {
			return nil, errs
		}

		component = decl.NewConstant[symbol.Unresolved](name, datatype, constExpr)
		p.srcmap.Put(component, p.spanOf(start, end-1))
		consts = append(consts, component)
	}
	//
	return consts, nil
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

func (p *Parser) parseFunction() (decl.Unresolved, []source.SyntaxError) {
	var (
		start    = p.index
		env      = EmptyEnvironment()
		name     string
		code     []stmt.Unresolved
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
	// Check for any effects
	if p.match(LESS_THAN) {
		if errs := p.parseMemoryEffects(env); len(errs) > 0 {
			return nil, errs
		}
	}
	// Parse inputs
	if errs = p.parseArgsList(variable.PARAMETER, env); len(errs) > 0 {
		return nil, errs
	}
	// Parse optional '->'
	if p.match(RIGHTARROW) {
		// Parse returns
		if errs = p.parseArgsList(variable.RETURN, env); len(errs) > 0 {
			return nil, errs
		}
	}
	// Save for source map
	end := p.index
	// Parse start of block
	if returned, code, errs = p.parseStatementBlock(0, env, util.None[uint](), util.None[uint]()); len(errs) > 0 {
		return nil, errs
	}
	// Sanity check for implicit or explicit return
	if !returned {
		stmt := &stmt.Return[symbol.Unresolved]{}
		// Implicit return
		code = append(code, stmt)
		// Associate return with span of final curly brace.
		p.srcmap.Put(stmt, p.previousToken().Span)
	}
	// Construct function
	fn := decl.NewFunction(name, env.Effects(), env.Variables(), code)
	//
	p.srcmap.Put(fn, p.spanOf(start, end-1))
	// Done
	return fn, nil
}

func (p *Parser) parseMemoryEffects(env Environment) []source.SyntaxError {
	var (
		arg   string
		errs  []source.SyntaxError
		first = true
	)
	// Parse entries until end brace
	for p.lookahead().Kind != GREATER_THAN {
		// look for ","
		if !first {
			if _, errs = p.expect(COMMA); len(errs) > 0 {
				return errs
			}
		}
		//
		first = false
		// save lookahead here so errors point at the name token
		lookahead := p.lookahead()
		// parse name first (new syntax: name:type)
		if arg, errs = p.parseIdentifier(); len(errs) > 0 {
			return errs
		} else if env.IsDeclared(arg) {
			return p.syntaxErrors(lookahead, "effect already declared")
		}
		// construct effect
		sym := symbol.NewUnresolved(arg, symbol.MEMORY_EFFECT, symbol.ANY_INPUTS)
		effect := &sym
		//
		p.srcmap.Put(effect, lookahead.Span)
		//
		env.DeclareEffect(effect)
	}
	// Advance past "}"
	p.match(GREATER_THAN)
	//
	return nil
}
func (p *Parser) parseArgsList(kind variable.Kind, env Environment) []source.SyntaxError {
	var (
		arg      string
		datatype Type
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
		// save lookahead here so errors point at the name token
		lookahead := p.lookahead()
		// parse name first (new syntax: name:type)
		if arg, errs = p.parseIdentifier(); len(errs) > 0 {
			return errs
		} else if env.IsDeclared(arg) {
			return p.syntaxErrors(lookahead, "variable already declared")
		}
		// parse ':'
		if _, errs = p.expect(COLON); len(errs) > 0 {
			return errs
		}
		// parse type
		if datatype, errs = p.parseType(); len(errs) > 0 {
			return errs
		}
		//
		env.DeclareVariable(kind, arg, datatype)
	}
	// Advance past "}"
	p.match(RBRACE)
	//
	return nil
}

func (p *Parser) parseInputOutputMemory() (decl.Unresolved, []source.SyntaxError) {
	var (
		public  bool
		name    string
		errs    []source.SyntaxError
		address []VariableDescriptor
		data    []VariableDescriptor
	)
	// Parse optional pub modifier; private by default
	if p.match(KEYWORD_PUB) {
		public = true
	}
	//
	lookahead := p.lookahead()
	// Validate keyword before consuming it
	switch lookahead.Kind {
	case KEYWORD_INPUT, KEYWORD_OUTPUT, KEYWORD_STATIC:
		p.index++
	default:
		return nil, p.syntaxErrors(lookahead, "unknown declaration")
	}
	// Parse the shared header: name(address) -> (data)
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	//
	if address, errs = p.parseMemoryArgsList(variable.PARAMETER); len(errs) > 0 {
		return nil, errs
	}
	//
	if _, errs = p.expect(RIGHTARROW); len(errs) > 0 {
		return nil, errs
	}
	//
	if data, errs = p.parseMemoryArgsList(variable.RETURN); len(errs) > 0 {
		return nil, errs
	}
	// Construct the appropriate memory declaration
	switch lookahead.Kind {
	case KEYWORD_INPUT:
		return decl.NewReadOnlyMemory[symbol.Unresolved](public, name, address, data), nil
	case KEYWORD_OUTPUT:
		return decl.NewWriteOnceMemory[symbol.Unresolved](public, name, address, data), nil
	default: // KEYWORD_STATIC
		contents, errs := p.parseStaticInitialiser()
		if len(errs) > 0 {
			return nil, errs
		}

		return decl.NewStaticMemory[symbol.Unresolved](public, name, address, data, contents), nil
	}
}

// parseStaticInitialiser parses a brace-enclosed comma-separated list of
// compile-time constant expressions: { expr, expr, ... }
func (p *Parser) parseStaticInitialiser() ([]expr.Unresolved, []source.SyntaxError) {
	var (
		contents []expr.Unresolved
		errs     []source.SyntaxError
		env      = EmptyEnvironment()
	)
	//
	if _, errs = p.expect(LCURLY); len(errs) > 0 {
		return nil, errs
	}
	//
	for p.lookahead().Kind != RCURLY {
		e, errs := p.parseTernaryOrExpr(env)
		if len(errs) > 0 {
			return nil, errs
		}

		contents = append(contents, e)
		// Consume comma separator; stop if next token is '}'
		if !p.match(COMMA) {
			break
		}
	}
	//
	if _, errs = p.expect(RCURLY); len(errs) > 0 {
		return nil, errs
	}
	//
	return contents, nil
}

func (p *Parser) parseReadWriteMemory() (decl.Unresolved, []source.SyntaxError) {
	var (
		name    string
		errs    []source.SyntaxError
		address []VariableDescriptor
		data    []VariableDescriptor
	)
	//
	if _, errs := p.expect(KEYWORD_MEMORY); len(errs) > 0 {
		return nil, errs
	}
	// Parse memory name first (function-style)
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	// Parse address args: (type param, ...)
	if address, errs = p.parseMemoryArgsList(variable.PARAMETER); len(errs) > 0 {
		return nil, errs
	}
	// Parse ->
	if _, errs = p.expect(RIGHTARROW); len(errs) > 0 {
		return nil, errs
	}
	// Parse data args: (type result, ...)
	if data, errs = p.parseMemoryArgsList(variable.RETURN); len(errs) > 0 {
		return nil, errs
	}
	// Done
	mem := decl.NewRandomAccessMemory[symbol.Unresolved](name, address, data)

	return mem, nil
}

func (p *Parser) parseTypeAlias() (decl.Unresolved, []source.SyntaxError) {
	var (
		start    = p.index
		errs     []source.SyntaxError
		datatype Type
		name     string
	)
	// Parse type declaration
	if _, errs := p.expect(KEYWORD_TYPE); len(errs) > 0 {
		return nil, errs
	} else if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	} else if _, errs = p.expect(EQUALS); len(errs) > 0 {
		return nil, errs
	} else if datatype, errs = p.parseType(); len(errs) > 0 {
		return nil, errs
	}
	// Save for source map
	end := p.index
	component := decl.NewTypeAlias[symbol.Unresolved](name, datatype)
	//
	p.srcmap.Put(component, p.spanOf(start, end-1))
	//
	return component, errs
}

// parseMemoryArgsList parses a function-style typed parameter list for memory
// declarations: (type name, type name, ...).  Returns both the combined type
// (for the address/data bus) and the individual named descriptors.
func (p *Parser) parseMemoryArgsList(kind variable.Kind) ([]VariableDescriptor, []source.SyntaxError) {
	var (
		params []VariableDescriptor
		errs   []source.SyntaxError
	)
	if _, errs = p.expect(LBRACE); len(errs) > 0 {
		return nil, errs
	}

	for p.lookahead().Kind != RBRACE {
		var t Type

		if len(params) > 0 {
			if _, errs = p.expect(COMMA); len(errs) > 0 {
				return nil, errs
			}
		}

		var pname string
		if pname, errs = p.parseIdentifier(); len(errs) > 0 {
			return nil, errs
		}

		if _, errs = p.expect(COLON); len(errs) > 0 {
			return nil, errs
		}

		if t, errs = p.parseType(); len(errs) > 0 {
			return nil, errs
		}

		params = append(params, variable.New(kind, pname, t))
	}

	p.match(RBRACE)

	return params, nil
}

func (p *Parser) parseType() (Type, []source.SyntaxError) {
	var (
		name  string
		errs  []source.SyntaxError
		start = p.index
	)
	//
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	// Parse to check if bitwidth is present
	bw, err := strconv.Atoi(name[1:])
	//
	switch {
	case strings.HasPrefix(name, "u") && err == nil:
		//
		return data.NewUnsignedInt[symbol.Unresolved](uint(bw), false), nil
	// we assume that if not a fundamental type, it is an alias
	default:
		alias := data.NewAlias[symbol.Unresolved](symbol.NewUnresolved(name, symbol.TYPE_ALIAS, 0))
		//
		p.srcmap.Put(alias, p.spanOf(start, p.index-1))
		//
		return alias, nil
	}
}

func (p *Parser) parseStatementBlock(pc uint, env Environment, breakLab, contLab util.Option[uint],
) (bool, []stmt.Unresolved, []source.SyntaxError) {
	//
	var (
		errs     []source.SyntaxError
		insns    []stmt.Unresolved
		returned bool
	)
	// Clone environment.  This is to ensure that variables declared in this
	// block do not clash with those of the same name declared elsewhere.
	env = env.Clone(breakLab, contLab)
	// Parse start of block
	if _, errs = p.expect(LCURLY); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse instructions until end of block
	for p.lookahead().Kind != RCURLY {
		var (
			ith []stmt.Unresolved
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

func (p *Parser) parseStatement(pc uint, env Environment,
) (bool, []stmt.Unresolved, []source.SyntaxError) {
	//
	var (
		// Save current position for backtracking
		start    = p.index
		errs     []source.SyntaxError
		insns    []stmt.Unresolved
		insn     stmt.Unresolved
		returned bool
	)
	//
	lookahead := p.lookahead()
	//
	switch lookahead.Kind {
	case KEYWORD_BREAK:
		returned, insn, errs = p.parseBreak(env)
	case KEYWORD_CONTINUE:
		returned, insn, errs = p.parseContinue(env)
	case KEYWORD_FAIL:
		returned, insn, errs = p.parseFail()
	case KEYWORD_FOR:
		returned, insns, errs = p.parseFor(pc, env)
	case KEYWORD_IF:
		returned, insns, errs = p.parseIfElse(pc, env)
	case KEYWORD_PRINTF:
		returned, insn, errs = p.parsePrintf(env)
	case KEYWORD_RETURN:
		returned, insn, errs = p.parseReturn()
	case KEYWORD_WHILE:
		returned, insns, errs = p.parseWhile(pc, env)
	case KEYWORD_VAR:
		insns, errs = p.parseVar(env)
	case IDENTIFIER:
		// Detect a bare function call statement: name(...) with no assignment.
		if p.index+1 < len(p.tokens) && p.tokens[p.index+1].Kind == LBRACE {
			insn, errs = p.parseCallStatement(env)
		} else {
			insn, errs = p.parseAssignment(env)
		}
	default:
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

func (p *Parser) parseAssignment(env Environment) (stmt.Unresolved, []source.SyntaxError) {
	var (
		lhs  []LVal
		rhs  Expr
		errs []source.SyntaxError
	)
	// parse left-hand side
	if lhs, errs = p.parseLVals(env); len(errs) > 0 {
		return nil, errs
	}
	// Reverse items so that least significant comes first.
	lhs = array.Reverse(lhs)
	// Parse '='
	if _, errs = p.expect(EQUALS); len(errs) > 0 {
		return nil, errs
	}
	// Parse right-hand side
	if rhs, errs = p.parseTernaryOrExpr(env); len(errs) > 0 {
		return nil, errs
	}
	// Done
	return &stmt.Assign[symbol.Unresolved]{Targets: lhs, Source: rhs}, nil
}

func (p *Parser) parseCallStatement(env Environment) (stmt.Unresolved, []source.SyntaxError) {
	// Parse call as a general expression, since this ensures source mapping is
	// handled.  This means, however, that we need to check afterwards that we
	// actually got a call expression rather than a general expression.
	call, errs := p.parseExpr(env)
	//
	if len(errs) > 0 {
		return nil, errs
	} else if ea, ok := call.(*expr.ExternAccess[symbol.Unresolved]); ok && ea.Name.Kind == symbol.FUNCTION {
		// Yes, its a function call
		return &stmt.Assign[symbol.Unresolved]{Targets: nil, Source: call}, nil
	}
	// No, its some other kind of expression.
	return nil, p.srcmap.SyntaxErrors(call, "expression unused")
}

func (p *Parser) parseIfElse(pc uint, env Environment) (bool, []stmt.Unresolved, []source.SyntaxError) {
	var (
		errs              []source.SyntaxError
		trueBranch        []stmt.Unresolved
		falseBranch       []stmt.Unresolved
		falseLabel        = env.FreshLabel()
		insns             []stmt.Unresolved
		trueRet, falseRet bool
	)
	// Match if
	if _, errs := p.expect(KEYWORD_IF); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse condition
	if insns, errs = p.parseCondition(pc, false, falseLabel, env); len(errs) > 0 {
		return false, nil, errs
	}
	//
	n := uint(len(insns))
	// Parse true branch
	if trueRet, trueBranch, errs = p.parseStatementBlock(pc+n, env, env.BreakLabel(), env.ContinueLabel()); len(errs) > 0 {
		return false, nil, errs
	}
	// falseTarget for if-goto
	falseTarget := pc + n + uint(len(trueBranch))
	// Include the true branch
	insns = append(insns, trueBranch...)
	// Check for "else"
	if p.lookahead().Kind == KEYWORD_ELSE {
		// Skip over else
		_, _ = p.expect(KEYWORD_ELSE)
		// add branch bypass (if needed)
		if !trueRet {
			falseTarget++
		}
		// parse false branch (either a block or an else-if chain)
		if p.lookahead().Kind == KEYWORD_IF {
			falseRet, falseBranch, errs = p.parseIfElse(falseTarget, env)
		} else {
			falseRet, falseBranch, errs = p.parseStatementBlock(falseTarget, env, env.BreakLabel(), env.ContinueLabel())
		}
		// Sanity check errors
		if len(errs) > 0 {
			return false, nil, errs
		}
		// add bypass (if applicable)
		if !trueRet {
			endTarget := falseTarget + uint(len(falseBranch))
			insns = append(insns, &stmt.Goto[symbol.Unresolved]{Target: endTarget})
		}
		// add false branch (if applicable)
		insns = append(insns, falseBranch...)
	}
	// Patch false target
	patchBranches(falseLabel, insns, falseTarget)
	// Done
	return trueRet && falseRet, insns, nil
}

func (p *Parser) parseWhile(pc uint, env Environment) (bool, []stmt.Unresolved, []source.SyntaxError) {
	var (
		errs           []source.SyntaxError
		insns, body    []stmt.Unresolved
		breakLabel     = env.FreshLabel()
		continueLabel  = env.FreshLabel()
		conditionLabel = env.FreshLabel()
	)
	// Match while
	if _, errs = p.expect(KEYWORD_WHILE); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse condition
	if insns, errs = p.parseCondition(pc, false, conditionLabel, env); len(errs) > 0 {
		return false, nil, errs
	}
	//
	n := uint(len(insns))
	// Parse body block; body starts at pc+1
	if _, body, errs = p.parseStatementBlock(pc+n, env, util.Some(breakLabel), util.Some(continueLabel)); len(errs) > 0 {
		return false, nil, errs
	}
	//
	insns = append(insns, body...)
	insns = append(insns, &stmt.Goto[symbol.Unresolved]{Target: pc})
	// The conditional skip jumps past the back-goto to the instruction after the loop
	exitTarget := pc + uint(len(insns))
	// Patch any condition labels to jump to exit
	patchBranches(conditionLabel, insns, exitTarget)
	// Patch any break labels to jump to exit
	patchBranches(breakLabel, insns, exitTarget)
	// Patch any continue labels to jump back to condition check
	patchBranches(continueLabel, insns, pc)
	// A while loop never guarantees a return
	return false, insns, nil
}

func (p *Parser) parseFor(pc uint, env Environment) (bool, []stmt.Unresolved, []source.SyntaxError) {
	var (
		errs              []source.SyntaxError
		init, post        stmt.Unresolved
		insns, body, cond []stmt.Unresolved
		breakLabel        = env.FreshLabel()
		continueLabel     = env.FreshLabel()
		conditionLabel    = env.FreshLabel()
	)
	// Match 'for'
	if _, errs = p.expect(KEYWORD_FOR); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse init: either an inline variable declaration (name:type = expr) or a plain assignment
	if init, errs = p.parseForInit(env); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse ';'
	if _, errs = p.expect(SEMICOLON); len(errs) > 0 {
		return false, nil, errs
	}
	// Parse condition
	if cond, errs = p.parseCondition(pc+1, false, conditionLabel, env); len(errs) > 0 {
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
	//
	n := uint(len(cond))
	// Parse body; starts at condPC+1 = pc+2
	if _, body, errs = p.parseStatementBlock(pc+1+n, env, util.Some(breakLabel), util.Some(continueLabel)); len(errs) > 0 {
		return false, nil, errs
	}
	// Build the instruction sequence
	insns = append(insns, init)
	insns = append(insns, cond...)
	insns = append(insns, body...)
	insns = append(insns, post)
	insns = append(insns, &stmt.Goto[symbol.Unresolved]{Target: pc + 1})
	// Fill in the conditional check: exit is the instruction after the back-goto
	exitTarget := pc + uint(len(insns))
	// Patch any condition labels to jump to exit
	patchBranches(conditionLabel, insns, exitTarget)
	// Patch any break labels to jump to exit
	patchBranches(breakLabel, insns, exitTarget)
	// Patch any continue labels to jump to the post instruction
	patchBranches(continueLabel, insns, pc+1+n+uint(len(body)))
	// A for loop never guarantees a return
	return false, insns, nil
}

// parseForInit parses the initialiser of a for loop.  It accepts either an
// inline variable declaration of the form "name:type = expr" or a plain
// assignment to an already-declared variable.
func (p *Parser) parseForInit(env Environment) (stmt.Unresolved, []source.SyntaxError) {
	// Detect "name:type = expr" by peeking one token ahead.
	if p.index+1 < len(p.tokens) &&
		p.tokens[p.index].Kind == IDENTIFIER &&
		p.tokens[p.index+1].Kind == COLON {
		// Inline variable declaration
		lookahead := p.lookahead()

		name, errs := p.parseIdentifier()
		if len(errs) > 0 {
			return nil, errs
		} else if env.IsDeclared(name) {
			return nil, p.syntaxErrors(lookahead, "variable already declared")
		}

		if _, errs = p.expect(COLON); len(errs) > 0 {
			return nil, errs
		}

		dt, errs := p.parseType()
		if len(errs) > 0 {
			return nil, errs
		}

		env.DeclareVariable(variable.LOCAL, name, dt)

		if _, errs = p.expect(EQUALS); len(errs) > 0 {
			return nil, errs
		}

		rhs, errs := p.parseTernaryOrExpr(env)
		if len(errs) > 0 {
			return nil, errs
		}

		target := lval.NewVariable[symbol.Unresolved](env.LookupVariable(name))

		return &stmt.Assign[symbol.Unresolved]{Targets: []LVal{target}, Source: rhs}, nil
	}
	// Fall back to a plain assignment to an already-declared variable.
	return p.parseAssignment(env)
}

func (p *Parser) parseReturn() (bool, stmt.Unresolved, []source.SyntaxError) {
	if _, errs := p.expect(KEYWORD_RETURN); len(errs) > 0 {
		return true, nil, errs
	}
	//
	return true, &stmt.Return[symbol.Unresolved]{}, nil
}

func (p *Parser) parseFail() (bool, stmt.Unresolved, []source.SyntaxError) {
	if _, errs := p.expect(KEYWORD_FAIL); len(errs) > 0 {
		return true, nil, errs
	}
	//
	return true, &stmt.Fail[symbol.Unresolved]{}, nil
}

func (p *Parser) parsePrintf(env Environment) (bool, stmt.Unresolved, []source.SyntaxError) {
	var (
		token  lex.Token
		chunks []stmt.FormattedChunk
	)
	//
	if _, errs := p.expect(KEYWORD_PRINTF); len(errs) > 0 {
		return false, nil, errs
	} else if token, errs = p.expect(STRING); len(errs) > 0 {
		return false, nil, errs
	}
	// Process string
	var (
		str    = p.string(token)
		runes  = []rune(str[1 : len(str)-1])
		nrunes []rune
		args   []Expr
		nargs  int
	)
	//
	for i := 0; i < len(runes); {
		switch runes[i] {
		case '%':
			var (
				f  zkc_util.Format
				ok bool
			)
			if i, f, ok = parseFormatting(i+1, runes); !ok {
				return false, nil, p.syntaxErrors(token, "invalid formatting")
			}
			//
			nargs++
			// append new chunkformat
			chunks = append(chunks, stmt.FormattedChunk{Text: string(nrunes), Format: f})
			// reset to build next chunk
			nrunes = nil
		case '\\':
			if i+1 == len(runes) {
				return false, nil, p.syntaxErrors(token, "invalid escape")
			}
			// Attempt to parse escape character
			c, ok := escapeCharacter(runes[i+1])
			//
			if !ok {
				return false, nil, p.syntaxErrors(token, "invalid escape")
			}
			// Continue
			nrunes = append(nrunes, c)
			i = i + 2
		default:
			nrunes = append(nrunes, runes[i])
			i = i + 1
		}
	}
	//
	if len(nrunes) > 0 {
		chunks = append(chunks, stmt.FormattedChunk{Text: string(nrunes), Format: zkc_util.EMPTY_FORMAT})
	}
	// parse expression arguments
	for p.match(COMMA) {
		arg, errs := p.parseExpr(env)
		//
		if len(errs) > 0 {
			return false, nil, errs
		}
		//
		args = append(args, arg)
	}
	// Sanity check for matching arguments / chunks
	if len(args) > nargs {
		return false, nil, p.srcmap.SyntaxErrors(args[nargs], "too many arguments")
	} else if len(args) > 0 && len(args) < nargs {
		n := len(args) - 1
		//
		return false, nil, p.srcmap.SyntaxErrors(args[n], "insufficient arguments")
	} else if len(args) < nargs {
		return false, nil, p.syntaxErrors(token, "insufficient arguments")
	}
	//
	return false, &stmt.Printf[symbol.Unresolved]{Chunks: chunks, Arguments: args}, nil
}

func parseFormatting(index int, runes []rune) (int, zkc_util.Format, bool) {
	if index >= len(runes) {
		return 0, zkc_util.EMPTY_FORMAT, false
	}
	//
	switch runes[index] {
	case 'd':
		return index + 1, zkc_util.DecimalFormat(), true
	case 'x':
		return index + 1, zkc_util.HexFormat(), true
	default:
		return 0, zkc_util.EMPTY_FORMAT, false
	}
}

func escapeCharacter(ch rune) (rune, bool) {
	switch ch {
	case 'n':
		return '\n', true
	case 'r':
		return '\r', true
	case 't':
		return '\t', true
	case '\\':
		return '\\', true
	}
	//
	return ' ', false
}

func (p *Parser) parseBreak(env Environment) (bool, stmt.Unresolved, []source.SyntaxError) {
	var (
		label     = env.BreakLabel()
		tok, errs = p.expect(KEYWORD_BREAK)
	)
	//
	if len(errs) > 0 {
		return true, nil, errs
	}

	if !label.HasValue() {
		return true, nil, p.syntaxErrors(tok, "break outside loop")
	}

	return true, &stmt.Goto[symbol.Unresolved]{Target: label.Unwrap()}, nil
}

func (p *Parser) parseContinue(env Environment) (bool, stmt.Unresolved, []source.SyntaxError) {
	var (
		label     = env.ContinueLabel()
		tok, errs = p.expect(KEYWORD_CONTINUE)
	)
	if len(errs) > 0 {
		return true, nil, errs
	}

	if !label.HasValue() {
		return true, nil, p.syntaxErrors(tok, "continue outside loop")
	}

	return true, &stmt.Goto[symbol.Unresolved]{Target: label.Unwrap()}, nil
}

func (p *Parser) parseVar(env Environment) ([]stmt.Unresolved, []source.SyntaxError) {
	var (
		errs  []source.SyntaxError
		names []string
		types []Type
	)
	// Consume 'var' keyword
	if _, errs = p.expect(KEYWORD_VAR); len(errs) > 0 {
		return nil, errs
	}
	// Parse one or more name:type pairs (comma-separated)
	for len(names) == 0 || p.match(COMMA) {
		// Store lookahead for error reporting
		lookahead := p.lookahead()
		//
		name, errs := p.parseIdentifier()
		//
		if len(errs) > 0 {
			return nil, errs
		} else if env.IsDeclared(name) {
			return nil, p.syntaxErrors(lookahead, "variable already declared")
		}
		// Parse ':'
		if _, errs = p.expect(COLON); len(errs) > 0 {
			return nil, errs
		}
		// Parse type
		dt, errs := p.parseType()
		if len(errs) > 0 {
			return nil, errs
		}
		//
		names = append(names, name)
		types = append(types, dt)
	}
	// Declare all variables before parsing any initialiser, so the
	// initialiser expression can reference other already-declared variables.
	for i, name := range names {
		env.DeclareVariable(variable.LOCAL, name, types[i])
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
	rhs, errs := p.parseTernaryOrExpr(env)
	if len(errs) > 0 {
		return nil, errs
	}
	// Build the assignment instruction
	target := env.LookupVariable(names[0])
	insn := &stmt.Assign[symbol.Unresolved]{
		Targets: []LVal{lval.NewVariable[symbol.Unresolved](target)},
		Source:  rhs,
	}
	//
	return []stmt.Unresolved{insn}, nil
}

func (p *Parser) parseCondition(pc uint, sign bool, target uint, env Environment,
) ([]stmt.Unresolved, []source.SyntaxError) {
	//
	var (
		errs []source.SyntaxError
		expr Expr
	)
	// Parse condition
	if expr, errs = p.parseExpr(env); len(errs) > 0 {
		return nil, errs
	}
	// Back-goto jumps to the if-goto at pc
	return p.flatternCondition(expr, pc, sign, target, env)
}

// parseTernaryOrExpr parses either a ternary expression (cond ? ifTrue : ifFalse)
// where cond is a comparison expression, or a plain arithmetic expression. It is
// the top-level expression parser for all value positions (assignments, var
// declarations, function arguments, etc.).
func (p *Parser) parseTernaryOrExpr(env Environment) (Expr, []source.SyntaxError) {
	start := p.index

	// First parse a full expression, which may already be a comparison expression.
	ex, errs := p.parseExpr(env)
	if len(errs) > 0 {
		return nil, errs
	}

	// If the next token is not '?', this is just a plain expression.
	if !p.follows(QMARK) {
		return ex, nil
	}

	// We have a '?' — this is a ternary candidate. Require the condition to be a
	// comparison expression; otherwise, leave parsing of '?' to higher-level logic.
	cond, ok := ex.(*expr.Cmp[symbol.Unresolved])
	if !ok {
		return ex, nil
	}

	if _, errs := p.expect(QMARK); len(errs) > 0 {
		return nil, errs
	}

	ifTrue, errs := p.parseTernaryOrExpr(env)
	if len(errs) > 0 {
		return nil, errs
	}

	if _, errs := p.expect(COLON); len(errs) > 0 {
		return nil, errs
	}

	ifFalse, errs := p.parseTernaryOrExpr(env)
	if len(errs) > 0 {
		return nil, errs
	}

	result := expr.NewTernary[symbol.Unresolved](cond, ifTrue, ifFalse)
	p.srcmap.Put(result, p.spanOf(start, p.index-1))

	return result, nil
}

func (p *Parser) parseExpr(env Environment) (Expr, []source.SyntaxError) {
	var (
		start     = p.index
		arg, errs = p.parseArithExpr(env)
		args      = []Expr{arg}
		tmp       Expr
	)
	// initialise lookahead
	kind := p.lookahead().Kind
	//
	for len(errs) == 0 && p.follows(LOGICAL_BINOPS...) {
		// Sanity check
		if !p.follows(kind) {
			return tmp, p.syntaxErrors(p.lookahead(), "braces required")
		}
		// Consume connective
		p.expect(p.lookahead().Kind)
		//
		tmp, errs = p.parseArithExpr(env)
		// Accumulate arguments
		args = append(args, tmp)
	}
	//
	switch {
	case len(errs) != 0:
		return nil, errs
	case len(args) == 1:
		return arg, nil
	case kind == LOGICAL_AND:
		arg = expr.NewLogicalAnd(args...)
	case kind == LOGICAL_OR:
		arg = expr.NewLogicalOr(args...)
	}
	//
	p.srcmap.Put(arg, p.spanOf(start, p.index-1))
	//
	return arg, errs
}

func (p *Parser) parseArithExpr(env Environment) (Expr, []source.SyntaxError) {
	var (
		start     = p.index
		arg, errs = p.parseConcatExpr(env)
		args      = []Expr{arg}
		tmp       Expr
		binary    bool
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
		tmp, errs = p.parseConcatExpr(env)
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
	case kind == BITWISE_AND:
		arg = expr.NewBitwiseAnd(args...)
	case kind == BITWISE_OR:
		arg = expr.NewBitwiseOr(args...)
	case kind == BITWISE_XOR:
		arg = expr.NewXor(args...)
	case kind == BITWISE_SHL:
		arg = expr.NewShl(args...)
	case kind == BITWISE_SHR:
		arg = expr.NewShr(args...)
	case kind == MUL:
		arg = expr.NewMul(args...)
	case kind == DIV:
		arg = expr.NewDiv(args...)
	case kind == REM:
		arg = expr.NewRem(args...)
	case kind == SUB:
		arg = expr.NewSub(args...)
	case kind == EQUALS_EQUALS:
		binary = true
		arg = expr.NewCmp(expr.EQ, args[0], args[1])
	case kind == NOT_EQUALS:
		binary = true
		arg = expr.NewCmp(expr.NEQ, args[0], args[1])
	case kind == LESS_THAN:
		binary = true
		arg = expr.NewCmp(expr.LT, args[0], args[1])
	case kind == LESS_THAN_EQUALS:
		binary = true
		arg = expr.NewCmp(expr.LTEQ, args[0], args[1])
	case kind == GREATER_THAN_EQUALS:
		binary = true
		arg = expr.NewCmp(expr.GTEQ, args[0], args[1])
	case kind == GREATER_THAN:
		binary = true
		arg = expr.NewCmp(expr.GT, args[0], args[1])
	}
	//
	p.srcmap.Put(arg, p.spanOf(start, p.index-1))
	//
	if binary && len(args) != 2 {
		return nil, p.srcmap.SyntaxErrors(arg, "invalid binary expression")
	}
	//
	return arg, nil
}

func (p *Parser) parseConcatExpr(env Environment) (Expr, []source.SyntaxError) {
	var (
		exprs = make([]Expr, 1)
		errs  []source.SyntaxError
		start = p.index
	)
	// Parse initial expression
	exprs[0], errs = p.parseUnitExpr(env)
	// Check for trailing concatenation
	for len(errs) == 0 && p.match(COLONCOLON) {
		var expr Expr
		//
		expr, errs = p.parseUnitExpr(env)
		exprs = append(exprs, expr)
	}
	//
	if len(errs) > 0 {
		return nil, errs
	} else if len(exprs) == 1 {
		return exprs[0], nil
	}
	// Bitwise concatenation
	arg := expr.NewConcat(exprs...)
	// Record span for this new expression
	p.srcmap.Put(arg, p.spanOf(start, p.index-1))
	//
	return arg, nil
}

func (p *Parser) parseUnitExpr(env Environment) (Expr, []source.SyntaxError) {
	var (
		lookahead = p.lookahead()
		errors    []source.SyntaxError
		start     = p.index
		nexpr     Expr
	)

	switch lookahead.Kind {
	case IDENTIFIER:
		nexpr, errors = p.parseAccessExpr(env)
	case NUMBER:
		var val big.Int
		//
		p.match(NUMBER)
		//
		val, errors = p.number(lookahead)
		base := p.baserOfNumber(lookahead)
		//
		nexpr = expr.NewConstant[symbol.Unresolved](val, base)
	case BITWISE_NOT:
		p.match(BITWISE_NOT)

		var operand Expr

		operand, errors = p.parseUnitExpr(env)
		if len(errors) == 0 {
			nexpr = expr.NewBitwiseNot(operand)
		}
	case LOGICAL_NOT:
		p.match(LOGICAL_NOT)

		var operand Expr

		operand, errors = p.parseUnitExpr(env)
		if len(errors) == 0 {
			nexpr = expr.NewLogicalNot(operand)
		}
	case LBRACE:
		p.match(LBRACE)
		nexpr, errors = p.parseTernaryOrExpr(env)
		//
		if len(errors) == 0 && !p.match(RBRACE) {
			return nil, p.syntaxErrors(lookahead, "expected )")
		}
		// Fall through to check for trailing `as` cast.
	default:
		return nil, p.syntaxErrors(lookahead, "unexpected token")
	}
	//
	if len(errors) == 0 && !p.srcmap.Has(nexpr) {
		p.srcmap.Put(nexpr, p.spanOf(start, p.index-1))
	}
	//
	if len(errors) == 0 && p.match(KEYWORD_AS) {
		var castType Type
		//
		if castType, errors = p.parseType(); len(errors) == 0 {
			cast := expr.NewCast(nexpr, castType)
			p.srcmap.Put(cast, p.spanOf(start, p.index-1))
			nexpr = cast
		}
	}
	//
	return nexpr, errors
}

func (p *Parser) parseAccessExpr(env Environment) (Expr, []source.SyntaxError) {
	var (
		nexpr Expr
		errs  []source.SyntaxError
		name  string
	)
	//
	name, errs = p.parseIdentifier()
	// now, check for function call or memory access
	if len(errs) == 0 && p.match(LSQUARE) {
		var args []Expr
		//
		args, errs = p.parseExprList(RSQUARE, env)
		//
		nexpr = expr.NewExternAccess(symbol.NewUnresolved(name, symbol.READABLE_MEMORY, uint(len(args))), args...)
	} else if len(errs) == 0 && p.match(LBRACE) {
		var args []Expr
		//
		args, errs = p.parseExprList(RBRACE, env)
		//
		nexpr = expr.NewExternAccess(symbol.NewUnresolved(name, symbol.FUNCTION, uint(len(args))), args...)
	} else if !env.IsDeclaredVariable(name) {
		// Constant access
		nexpr = expr.NewExternAccess(symbol.NewUnresolved(name, symbol.CONSTANT, 0))
	} else {
		// Register access
		rid := env.LookupVariable(name)
		// Done
		nexpr = expr.NewLocalAccess[symbol.Unresolved](rid)
	}
	//
	return nexpr, errs
}

// Parse sequence of one or more expressions separated by a comma.
// nolint
func (p *Parser) parseExprList(terminator uint, env Environment) ([]Expr, []source.SyntaxError) {
	var (
		lhs  = make([]Expr, 0)
		errs []source.SyntaxError
		expr Expr
	)
	// lhs may have additional registers
	for !p.match(terminator) {
		lookahead := p.lookahead()
		// match ",""
		if len(lhs) != 0 && !p.match(COMMA) {
			return nil, p.syntaxErrors(lookahead, "expected ,")
		}
		//
		if expr, errs = p.parseTernaryOrExpr(env); len(errs) > 0 {
			return nil, errs
		}
		// Add register to lhs
		lhs = append(lhs, expr)
	}
	//
	return lhs, nil
}

func (p *Parser) parseLVals(env Environment) ([]LVal, []source.SyntaxError) {
	var (
		lhs  []LVal = make([]LVal, 1)
		errs []source.SyntaxError
		reg  LVal
	)
	// lhs always starts with a register
	if lhs[0], errs = p.parseLVal(env); len(errs) > 0 {
		return nil, errs
	}
	// lhs may have additional registers
	for p.match(COMMA) {
		if reg, errs = p.parseLVal(env); len(errs) > 0 {
			return nil, errs
		}
		// Add register to lhs
		lhs = append(lhs, reg)
	}
	//
	return lhs, nil
}

func (p *Parser) parseLVal(env Environment) (LVal, []source.SyntaxError) {
	var (
		lv        LVal
		start     = p.index
		lookahead = p.lookahead()
		reg, errs = p.parseIdentifier()
		index     []Expr
	)
	//
	if len(errs) > 0 {
		return lv, errs
	} else if env.IsDeclaredVariable(reg) {
		var vars = []variable.Id{env.LookupVariable(reg)}
		// Look for destructuring lvals
		for p.match(COLONCOLON) {
			// save lookahead for error reporting
			lookahead = p.lookahead()
			//
			if reg, errs = p.parseIdentifier(); len(errs) > 0 {
				return lv, errs
			} else if !env.IsDeclaredVariable(reg) {
				return lv, p.syntaxErrors(lookahead, "unknown variable")
			}
			//
			vars = append(vars, env.LookupVariable(reg))
		}
		//
		lv = lval.NewVariable[symbol.Unresolved](vars...)
	} else if !p.match(LSQUARE) {
		return lv, p.syntaxErrors(lookahead, "unknown variable")
	} else if index, errs = p.parseExprList(RSQUARE, env); len(errs) > 0 {
		return lv, errs
	} else {
		// construct name symbol
		var name = symbol.NewUnresolved(reg, symbol.WRITEABLE_MEMORY, 1)
		// Done
		lv = lval.NewMemAccess(name, index)
	}
	// update source mapping
	p.srcmap.Put(lv, p.spanOf(start, p.index-1))
	//
	return lv, nil
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

// Lookahead returns the next token.  This must exist because EOF is always
// appended at the end of the token stream.
func (p *Parser) previousToken() lex.Token {
	if p.index == 0 {
		return p.tokens[p.index]
	}
	//
	return p.tokens[p.index-1]
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

// Flattern a given condition starting at a given program counter position into
// a sequence of one or more instructions.  The intention is: (1) for positive
// sign, branch to target label if condition holds (otherwise fall through); (2)
// for negative sign, branch to target label if condition does not hold
// (otherwise fall through).
func (p *Parser) flatternCondition(cond Expr, pc uint, sign bool, target uint, env Environment,
) ([]stmt.Unresolved, []source.SyntaxError) {
	switch c := cond.(type) {
	case *expr.Cmp[symbol.Unresolved]:
		return p.flatternComparison(c, sign, target)
	case *expr.LogicalAnd[symbol.Unresolved]:
		if sign {
			return p.flatternLogicalAnd(c.Exprs, pc, true, target, env)
		} else {
			return p.flatternLogicalOr(c.Exprs, pc, false, target, env)
		}
	case *expr.LogicalOr[symbol.Unresolved]:
		if sign {
			return p.flatternLogicalOr(c.Exprs, pc, true, target, env)
		} else {
			return p.flatternLogicalAnd(c.Exprs, pc, false, target, env)
		}
	case *expr.LogicalNot[symbol.Unresolved]:
		return p.flatternCondition(c.Expr, pc, !sign, target, env)
	default:
		return nil, p.srcmap.SyntaxErrors(cond, "invalid condition")
	}
}

func (p *Parser) flatternLogicalAnd(args []Expr, pc uint, sign bool, target uint, env Environment,
) ([]stmt.Unresolved, []source.SyntaxError) {
	var (
		label = env.FreshLabel()
		stmts []stmt.Unresolved
	)
	//
	for _, arg := range args {
		ss, errs := p.flatternCondition(arg, pc+uint(len(stmts)), !sign, label, env)
		//
		if len(errs) > 0 {
			return nil, errs
		}
		//
		stmts = append(stmts, ss...)
	}
	//
	stmts = append(stmts, &Goto{Target: target})
	// Patch labels
	patchBranches(label, stmts, pc+uint(len(stmts)))
	//
	return stmts, nil
}

func (p *Parser) flatternLogicalOr(args []Expr, pc uint, sign bool, target uint, env Environment,
) ([]stmt.Unresolved, []source.SyntaxError) {
	var stmts []stmt.Unresolved
	//
	for _, arg := range args {
		ss, errs := p.flatternCondition(arg, pc+uint(len(stmts)), sign, target, env)
		//
		if len(errs) > 0 {
			return nil, errs
		}
		//
		stmts = append(stmts, ss...)
	}
	//
	return stmts, nil
}

func (p *Parser) flatternComparison(cond *expr.Cmp[symbol.Unresolved], sign bool, target uint,
) ([]stmt.Unresolved, []source.SyntaxError) {
	var stmts []stmt.Unresolved
	//
	if sign {
		return append(stmts, &IfGoto{Cond: cond, Target: target}), nil
	} else {
		return append(stmts, &IfGoto{Cond: cond.Negate(), Target: target}), nil
	}
}

// patchBranches replaces all occurrences of the given marker in a branching
// instruction (i.e. goto or if-goto) with the given target.
func patchBranches(label uint, insns []stmt.Unresolved, target uint) {
	for _, insn := range insns {
		if g, ok := insn.(*stmt.Goto[symbol.Unresolved]); ok && g.Target == label {
			g.Target = target
		} else if g, ok := insn.(*stmt.IfGoto[symbol.Unresolved]); ok && g.Target == label {
			g.Target = target
		}
	}
}
