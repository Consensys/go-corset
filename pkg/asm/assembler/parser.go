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
package assembler

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/lex"
)

// MacroFunction is a function whose instructions are themselves macro
// instructions.  A macro function must be compiled down into a micro function
// before we can generate constraints.
type MacroFunction = io.Function[macro.Instruction]

// MicroFunction is a function whose instructions are themselves micro
// instructions.  A micro function represents the lowest representation of a
// function, where each instruction is made up of microcodes.
type MicroFunction = io.Function[micro.Instruction]

// Parse accepts a given source file representing an assembly language
// program, and assembles it into an instruction sequence which can then the
// executed.
func Parse(srcfile *source.File) (AssemblyItem, []source.SyntaxError) {
	parser := NewParser(srcfile)
	// Parse functions
	return parser.Parse()
}

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
func (p *Parser) Parse() (AssemblyItem, []source.SyntaxError) {
	var (
		item    AssemblyItem
		include *string
		errors  []source.SyntaxError
		fn      MacroFunction
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
		case KEYWORD_INCLUDE:
			include, errors = p.parseInclude()
			if len(errors) == 0 {
				item.Includes = append(item.Includes, include)
			}
			// Avoid appending to components
			continue
		case KEYWORD_FN:
			fn, errors = p.parseFunction()
		default:
			errors = p.syntaxErrors(lookahead, "unknown declaration")
		}
		//
		if len(errors) > 0 {
			return item, errors
		}
		//
		item.Components = append(item.Components, fn)
	}
	// Copy over source map
	item.SourceMap = *p.srcmap
	//
	return item, nil
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

func (p *Parser) parseFunction() (MacroFunction, []source.SyntaxError) {
	var (
		env             Environment
		inst            macro.Instruction
		name            string
		inputs, outputs []io.Register
		code            []macro.Instruction
		errs            []source.SyntaxError
		pc              uint
	)
	// Parse function declaration
	if _, errs := p.expect(KEYWORD_FN); len(errs) > 0 {
		return MacroFunction{}, errs
	}
	// Parse function name
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return MacroFunction{}, errs
	}
	// Parse inputs
	if inputs, errs = p.parseArgsList(schema.INPUT_REGISTER); len(errs) > 0 {
		return MacroFunction{}, errs
	}
	// Parse optional '->'
	if p.match(RIGHTARROW) {
		// Parse returns
		if outputs, errs = p.parseArgsList(schema.OUTPUT_REGISTER); len(errs) > 0 {
			return MacroFunction{}, errs
		}
	}
	// Update register list with inputs/outputs
	env.registers = append(env.registers, inputs...)
	env.registers = append(env.registers, outputs...)
	// Parse start of block
	if _, errs = p.expect(LCURLY); len(errs) > 0 {
		return MacroFunction{}, errs
	}
	// Parse instructions until end of block
	for p.lookahead().Kind != RCURLY {
		if inst, errs = p.parseMacroInstruction(pc, &env); len(errs) > 0 {
			return MacroFunction{}, errs
		}
		//
		if inst != nil {
			code = append(code, inst)
			// inc pc only for real instructions.
			pc = pc + 1
		}
	}
	// Advance past "}"
	p.match(RCURLY)
	// Finalise labels
	env.BindLabels(code)
	// Done
	return io.NewFunction(name, env.registers, env.buses, code), nil
}

func (p *Parser) parseArgsList(kind schema.RegisterType) ([]io.Register, []source.SyntaxError) {
	var (
		arg     string
		width   uint
		errs    []source.SyntaxError
		regs    []io.Register
		padding big.Int
	)
	// Parse start of list
	if _, errs = p.expect(LBRACE); len(errs) > 0 {
		return nil, errs
	}
	// Parse entries until end brace
	for p.lookahead().Kind != RBRACE {
		// look for ","
		if len(regs) != 0 {
			if _, errs = p.expect(COMMA); len(errs) > 0 {
				return nil, errs
			}
		}
		// parse name, type & optional padding
		if arg, errs = p.parseIdentifier(); len(errs) > 0 {
			return nil, errs
		} else if padding, errs = p.parseOptionalPadding(); len(errs) > 0 {
			return nil, errs
		} else if width, errs = p.parseType(); len(errs) > 0 {
			return nil, errs
		}
		//
		regs = append(regs, schema.NewRegister(kind, arg, width, padding))
	}
	// Advance past "}"
	p.match(RBRACE)
	//
	return regs, nil
}

func (p *Parser) parseOptionalPadding() (big.Int, []source.SyntaxError) {
	var (
		padding   big.Int
		errs      []source.SyntaxError
		lookahead lex.Token
	)
	//
	if !p.match(EQUALS) {
		// no optional padding provided
		return padding, nil
	} else if lookahead, errs = p.expect(NUMBER); len(errs) > 0 {
		return padding, errs
	}
	// Yes, optional padding provided
	return p.number(lookahead), nil
}

func (p *Parser) parseType() (uint, []source.SyntaxError) {
	var (
		lookahead = p.lookahead()
		name      string
		errs      []source.SyntaxError
	)
	//
	if name, errs = p.parseIdentifier(); len(errs) > 0 {
		return 0, errs
	}
	//
	switch {
	case strings.HasPrefix(name, "u"):
		// Parse bitwidth
		bw, err := strconv.Atoi(name[1:])
		//
		if err != nil {
			return 0, p.syntaxErrors(lookahead, err.Error())
		}
		//
		return uint(bw), nil
	default:
		return 0, p.syntaxErrors(lookahead, "unknown type")
	}
}

func (p *Parser) parseMacroInstruction(pc uint, env *Environment) (macro.Instruction, []source.SyntaxError) {
	var (
		// Save current position for backtracking
		start = p.index
		errs  []source.SyntaxError
		first string
		insn  macro.Instruction
	)
	//
	if first, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	//
	switch first {
	case "fail":
		insn, errs = &macro.Fail{}, nil
	case "goto":
		insn, errs = p.parseGoto(env)
	case "if":
		insn, errs = p.parseIfGoto(env)
	case "return":
		insn, errs = &macro.Return{}, nil
	case "var":
		return nil, p.parseVar(env)
	default:
		isLabel := p.lookahead().Kind == COLON
		// Backtrack
		p.index = start
		// Distinguish label from assignment
		if isLabel {
			return p.parseLabel(pc, env)
		} else {
			insn, errs = p.parseAssignment(env)
		}
	}
	// Record source mapping
	if insn != nil {
		p.srcmap.Put(insn, p.spanOf(start, p.index))
	}
	//
	return insn, errs
}

func (p *Parser) parseGoto(env *Environment) (macro.Instruction, []source.SyntaxError) {
	lab, errs := p.parseIdentifier()
	//
	if len(errs) > 0 {
		return nil, errs
	}
	//
	return &macro.Goto{
		Target: env.BindLabel(lab)}, nil
}

func (p *Parser) parseLabel(pc uint, env *Environment) (macro.Instruction, []source.SyntaxError) {
	// Observe, following cannot fail
	tok, _ := p.expect(IDENTIFIER)
	// Likewise, this cannot fail
	p.expect(COLON)
	//
	lab := p.string(tok)
	//
	if env.IsBoundLabel(lab) {
		return nil, p.syntaxErrors(tok, "label already declared")
	}
	// Declare label at given pc
	env.DeclareLabel(lab, pc)
	// Done
	return nil, nil
}

func (p *Parser) parseVar(env *Environment) []source.SyntaxError {
	var (
		errs  []source.SyntaxError
		names []string
		width uint
	)
	// Parse name(s)
	for len(names) == 0 || p.match(COMMA) {
		// Store lookahead for error reporting
		lookahead := p.lookahead()
		//
		name, errs := p.parseIdentifier()
		//
		if len(errs) > 0 {
			return errs
		} else if env.IsRegister(name) {
			return p.syntaxErrors(lookahead, "variable already declared")
		}
		//
		names = append(names, name)
	}
	// parse bitwidth
	if width, errs = p.parseType(); len(errs) > 0 {
		return errs
	}
	//
	for _, name := range names {
		env.DeclareRegister(schema.COMPUTED_REGISTER, name, width)
	}
	//
	return nil
}

func (p *Parser) parseIfGoto(env *Environment) (macro.Instruction, []source.SyntaxError) {
	var (
		errs     []source.SyntaxError
		lhs, rhs io.RegisterId
		constant big.Int
		label    string
		cond     uint8
	)
	// Parse left hand side
	if lhs, errs = p.parseRegister(env); len(errs) > 0 {
		return nil, errs
	}
	// save lookahead for error reporting
	if cond, errs = p.parseComparator(); len(errs) > 0 {
		return nil, errs
	}
	// Parse right hand side
	if rhs, constant, errs = p.parseRegisterOrConstant(env); len(errs) > 0 {
		return nil, errs
	}
	// Parse "goto"
	if errs = p.parseKeyword("goto"); len(errs) > 0 {
		return nil, errs
	}
	// Parse target label
	if label, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	//
	return &macro.IfGoto{
		Cond:     cond,
		Left:     lhs,
		Right:    rhs,
		Constant: constant,
		Target:   env.BindLabel(label),
	}, nil
}

func (p *Parser) parseAssignment(env *Environment) (macro.Instruction, []source.SyntaxError) {
	var (
		lhs      []io.RegisterId
		rhs      []io.RegisterId
		constant big.Int
		errs     []source.SyntaxError
		kind     uint
		insn     macro.Instruction
	)
	// parse left-hand side
	if lhs, errs = p.parseAssignmentLhs(env); len(errs) > 0 {
		return nil, errs
	}
	// Parse '='
	if _, errs = p.expect(EQUALS); len(errs) > 0 {
		return nil, errs
	}
	// Check what we've got
	if p.follows(IDENTIFIER, LBRACE) {
		// function call
		return p.parseCallRhs(lhs, env)
	} else {
		// Reverse items so that least significant comes first.  NOTE:
		// eventually should be updated to retain the given order.
		lhs = array.Reverse(lhs)
		// Parse right-hand side
		if kind, rhs, constant, errs = p.parseAssignmentRhs(env); len(errs) > 0 {
			return nil, errs
		}
		//
		switch kind {
		case ADD:
			insn = &macro.Add{Targets: lhs, Sources: rhs, Constant: constant}
		case SUB:
			insn = &macro.Sub{Targets: lhs, Sources: rhs, Constant: constant}
		case MUL:
			insn = &macro.Mul{Targets: lhs, Sources: rhs, Constant: constant}
		default:
			panic("unreachable")
		}
		// Done
		return insn, nil
	}
}

func (p *Parser) parseAssignmentLhs(env *Environment) ([]io.RegisterId, []source.SyntaxError) {
	lhs, errs := p.parseRegisterList(env)
	//
	return lhs, errs
}

func (p *Parser) parseAssignmentRhs(env *Environment) (uint, []io.RegisterId, big.Int, []source.SyntaxError) {
	var (
		constant big.Int
		rhs      []io.RegisterId
		kind     uint = ADD
		reg      io.RegisterId
		errs     []source.SyntaxError
	)
	//
	for p.lookahead().Kind == IDENTIFIER {
		if reg, errs = p.parseRegister(env); len(errs) > 0 {
			return 0, nil, constant, errs
		}
		// Append reg
		rhs = append(rhs, reg)
		// Parse trailing + / - / *
		if tok, ok := p.parseAssignmentOp(); !ok {
			// Special case for multiply!
			if kind == MUL {
				constant = *big.NewInt(1)
			}
			//
			return kind, rhs, constant, nil
		} else if len(rhs) == 1 {
			// first time around
			kind = tok.Kind
		} else if kind != tok.Kind {
			// subsequent times around
			return 0, nil, constant, p.syntaxErrors(tok, "inconsistent operation")
		}
	}
	// If we get here, we are expecting a constant.
	lookahead, errs := p.expect(NUMBER)
	//
	if len(errs) > 0 {
		return 0, nil, constant, errs
	}
	//
	return kind, rhs, p.number(lookahead), errs
}

func (p *Parser) parseAssignmentOp() (lex.Token, bool) {
	lookahead := p.lookahead()
	//
	switch lookahead.Kind {
	case ADD, SUB, MUL:
		// Match the token
		p.match(lookahead.Kind)
		//
		return lookahead, true
	default:
		return lex.Token{}, false
	}
}

func (p *Parser) parseCallRhs(lhs []io.RegisterId, env *Environment) (macro.Instruction, []source.SyntaxError) {
	var (
		errs []source.SyntaxError
		rhs  []io.RegisterId
		fn   string
	)
	//
	if fn, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	} else if _, errs = p.expect(LBRACE); len(errs) > 0 {
		return nil, errs
	} else if rhs, errs = p.parseRegisterList(env); len(errs) > 0 {
		return nil, errs
	} else if _, errs = p.expect(RBRACE); len(errs) > 0 {
		return nil, errs
	}
	// Generate temporary bus identifier
	bus := env.BindBus(fn)
	// Done
	return macro.NewCall(bus, lhs, rhs), nil
}

// Parse sequence of one or more registers separated by a comma.
func (p *Parser) parseRegisterList(env *Environment) ([]io.RegisterId, []source.SyntaxError) {
	var (
		lhs  []io.RegisterId = make([]io.RegisterId, 1)
		errs []source.SyntaxError
		reg  io.RegisterId
	)
	// lhs always starts with a register
	if lhs[0], errs = p.parseRegister(env); len(errs) > 0 {
		return nil, errs
	}
	// lhs may have additional registers
	for p.match(COMMA) {
		if reg, errs = p.parseRegister(env); len(errs) > 0 {
			return nil, errs
		}
		// Add register to lhs
		lhs = append(lhs, reg)
	}
	//
	return lhs, nil
}

func (p *Parser) parseRegisterOrConstant(env *Environment) (io.RegisterId, big.Int, []source.SyntaxError) {
	var (
		reg      io.RegisterId
		constant big.Int
		errs     []source.SyntaxError
	)
	//
	lookahead := p.lookahead()
	//
	switch lookahead.Kind {
	case IDENTIFIER:
		reg, errs = p.parseRegister(env)
	case NUMBER:
		p.match(NUMBER)
		//
		reg = schema.NewUnusedRegisterId()
		constant = p.number(lookahead)
	default:
		errs = p.syntaxErrors(lookahead, "expecting register or constant")
	}
	//
	return reg, constant, errs
}

func (p *Parser) parseRegister(env *Environment) (io.RegisterId, []source.SyntaxError) {
	lookahead := p.lookahead()
	reg, errs := p.parseIdentifier()
	//
	if len(errs) > 0 {
		return io.RegisterId{}, errs
	} else if !env.IsRegister(reg) {
		return io.RegisterId{}, p.syntaxErrors(lookahead, "unknown register")
	}
	// Done
	return env.LookupRegister(reg), nil
}

func (p *Parser) parseKeyword(keyword string) []source.SyntaxError {
	tok, errs := p.expect(IDENTIFIER)
	//
	if len(errs) > 0 {
		return errs
	} else if p.string(tok) != keyword {
		return p.syntaxErrors(tok, fmt.Sprintf("expected \"%s\"", keyword))
	}
	//
	return nil
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

func (p *Parser) parseComparator() (uint8, []source.SyntaxError) {
	var (
		lookahead = p.lookahead()
		op        uint8
	)
	// Parse operation
	switch lookahead.Kind {
	case EQUALS_EQUALS:
		op = macro.EQ
	case NOT_EQUALS:
		op = macro.NEQ
	case LESS_THAN:
		op = macro.LT
	case LESS_THAN_EQUALS:
		op = macro.LTEQ
	case GREATER_THAN:
		op = macro.GT
	case GREATER_THAN_EQUALS:
		op = macro.GTEQ
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
func (p *Parser) number(token lex.Token) big.Int {
	var number big.Int
	//
	number.SetString(p.string(token), 0)
	//
	return number
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

// Follows attempts to check what follows the current position.
func (p *Parser) follows(kinds ...uint) bool {
	for i, kind := range kinds {
		n := i + p.index
		if n >= len(p.tokens) {
			return false
		} else if p.tokens[n].Kind != kind {
			return false
		}
	}
	//
	return true
}
func (p *Parser) spanOf(firstToken, lastToken int) source.Span {
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
