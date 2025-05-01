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
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/insn"
	instruction "github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/asm/macro"
	"github.com/consensys/go-corset/pkg/asm/micro"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/lex"
)

// Parse accepts a given source file representing an assembly language
// program, and assembles it into an instruction sequence which can then the
// executed.
func Parse(srcfile *source.File) ([]MacroFunction, *source.Map[macro.Instruction], []source.SyntaxError) {
	parser := NewParser(srcfile)
	// Parse functions
	return parser.parse()
}

// ============================================================================
// Lexer
// ============================================================================

// END_OF signals "end of file"
const END_OF uint = 0

// WHITESPACE signals whitespace
const WHITESPACE uint = 1

// COMMENT signals ";; ... \n"
const COMMENT uint = 2

// LBRACE signals "("
const LBRACE uint = 3

// RBRACE signals ")"
const RBRACE uint = 4

// LCURLY signals "{"
const LCURLY uint = 5

// RCURLY signals "}"
const RCURLY uint = 6

// COMMA signals ","
const COMMA uint = 7

// COLON signals ":"
const COLON uint = 8

// SEMICOLON signals ":"
const SEMICOLON uint = 9

// NUMBER signals an integer number
const NUMBER uint = 10

// IDENTIFIER signals a column variable.
const IDENTIFIER uint = 11

// RIGHTARROW signals "->"
const RIGHTARROW uint = 12

// EQUALS signals "="
const EQUALS uint = 13

// ADD signals "+"
const ADD uint = 14

// SUB signals "-"
const SUB uint = 15

// MUL signals "*"
const MUL uint = 16

// Rule for describing whitespace
var whitespace lex.Scanner[rune] = lex.Many(lex.Or(lex.Unit(' '), lex.Unit('\t'), lex.Unit('\n')))

// Rule for describing numbers
var number lex.Scanner[rune] = lex.Many(lex.Within('0', '9'))

var identifierStart lex.Scanner[rune] = lex.Or(
	lex.Unit('_'),
	lex.Unit('\''),
	lex.Within('a', 'z'),
	lex.Within('A', 'Z'))

var identifierRest lex.Scanner[rune] = lex.Many(lex.Or(
	lex.Unit('_'),
	lex.Unit('\''),
	lex.Within('0', '9'),
	lex.Within('a', 'z'),
	lex.Within('A', 'Z')))

// Rule for describing identifiers
var identifier lex.Scanner[rune] = lex.And(identifierStart, identifierRest)

// Comments start with ';;'
var commentStart lex.Scanner[rune] = lex.Unit(';', ';')

// Comments continue until a newline or EOF.
var commentRest lex.Scanner[rune] = lex.Until('\n')

var comment lex.Scanner[rune] = lex.And(commentStart, commentRest)

// lexing rules
var rules []lex.LexRule[rune] = []lex.LexRule[rune]{
	lex.Rule(comment, COMMENT),
	lex.Rule(lex.Unit('('), LBRACE),
	lex.Rule(lex.Unit(')'), RBRACE),
	lex.Rule(lex.Unit('{'), LCURLY),
	lex.Rule(lex.Unit('}'), RCURLY),
	lex.Rule(lex.Unit(','), COMMA),
	lex.Rule(lex.Unit(':'), COLON),
	lex.Rule(lex.Unit(';'), SEMICOLON),
	lex.Rule(lex.Unit('-', '>'), RIGHTARROW),
	lex.Rule(lex.Unit('='), EQUALS),
	lex.Rule(lex.Unit('+'), ADD),
	lex.Rule(lex.Unit('-'), SUB),
	lex.Rule(lex.Unit('*'), MUL),
	lex.Rule(whitespace, WHITESPACE),
	lex.Rule(number, NUMBER),
	lex.Rule(identifier, IDENTIFIER),
	lex.Rule(lex.Eof[rune](), END_OF),
}

// ============================================================================
// Assembler
// ============================================================================

// Parser is a parser for assembly language.
type Parser struct {
	srcfile *source.File
	tokens  []lex.Token
	// Source mapping
	srcmap *source.Map[macro.Instruction]
	// Position within the tokens
	index int
}

// NewParser constructs a new parser for a given source file.
func NewParser(srcfile *source.File) *Parser {
	// Construct (initially empty) source mapping
	srcmap := source.NewSourceMap[macro.Instruction](*srcfile)
	//
	return &Parser{srcfile, nil, srcmap, 0}
}

func (p *Parser) parse() ([]MacroFunction, *source.Map[macro.Instruction], []source.SyntaxError) {
	var fns []MacroFunction
	// Initialise tokens array
	if errs := p.lex(); len(errs) > 0 {
		return nil, p.srcmap, errs
	}
	// Continue going until all consumed
	for p.lookahead().Kind != END_OF {
		fn, errs := p.parseFunction()
		//
		if len(errs) > 0 {
			return nil, p.srcmap, errs
		}
		//
		fns = append(fns, fn)
	}
	//
	return fns, p.srcmap, nil
}

// Initialise lexer and lex contents
func (p *Parser) lex() []source.SyntaxError {
	var (
		lexer = lex.NewLexer(p.srcfile.Contents(), rules...)
		// Lex as many tokens as possible
		tokens = lexer.Collect()
	)
	// Check whether anything was left (if so this is an error)
	if lexer.Remaining() != 0 {
		start, end := lexer.Index(), lexer.Index()+lexer.Remaining()
		err := p.srcfile.SyntaxError(source.NewSpan(int(start), int(end)), "unknown text encountered")
		// errors
		return []source.SyntaxError{*err}
	}
	// Remove any whitespace
	tokens = util.RemoveMatching(tokens, func(t lex.Token) bool { return t.Kind == WHITESPACE })
	// Remove any comments
	p.tokens = util.RemoveMatching(tokens, func(t lex.Token) bool { return t.Kind == COMMENT })
	//
	return nil
}

func (p *Parser) parseFunction() (MacroFunction, []source.SyntaxError) {
	var (
		fn              MacroFunction
		env             Environment
		inst            macro.Instruction
		inputs, outputs []Register
		errs            []source.SyntaxError
		pc              uint
	)
	// Parse function declaration
	if fn.Name, errs = p.parseIdentifier(); len(errs) > 0 || fn.Name != "fn" {
		return fn, errs
	}
	// Parse function name
	if fn.Name, errs = p.parseIdentifier(); len(errs) > 0 {
		return fn, errs
	}
	// Parse inputs
	if inputs, errs = p.parseArgsList(insn.INPUT_REGISTER); len(errs) > 0 {
		return fn, errs
	}
	// Parse '->'
	if _, errs = p.expect(RIGHTARROW); len(errs) > 0 {
		return fn, errs
	}
	// Parse outputs
	if outputs, errs = p.parseArgsList(insn.OUTPUT_REGISTER); len(errs) > 0 {
		return fn, errs
	}
	// Initialise register list from inputs/outputs
	env.registers = append(inputs, outputs...)
	// Parse start of block
	if _, errs = p.expect(LCURLY); len(errs) > 0 {
		return fn, errs
	}
	// Parse instructions until end of block
	for p.lookahead().Kind != RCURLY {
		if inst, errs = p.parseMacroInstruction(pc, &env); len(errs) > 0 {
			return fn, errs
		}
		//
		if inst != nil {
			fn.Code = append(fn.Code, inst)
			// inc pc only for real instructions.
			pc = pc + 1
		}
	}
	// Advance past "}"
	p.match(RCURLY)
	// Finalise labels
	env.BindLabels(fn.Code)
	// Assign registers
	fn.Registers = env.registers
	//
	return fn, nil
}

func (p *Parser) parseArgsList(kind uint8) ([]Register, []source.SyntaxError) {
	var (
		arg   string
		width uint
		errs  []source.SyntaxError
		regs  []Register
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
		// parse name & type
		if arg, errs = p.parseIdentifier(); len(errs) > 0 {
			return nil, errs
		} else if width, errs = p.parseType(); len(errs) > 0 {
			return nil, errs
		}
		//
		regs = append(regs, insn.NewRegister(kind, arg, width))
	}
	// Advance past "}"
	p.match(RBRACE)
	//
	return regs, nil
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

/*
	func (p *Parser) parseVectorInstruction(pc uint, env *Environment) (Instruction, []source.SyntaxError) {
		var (
			insns  []macro.Instruction = make([]macro.Instruction, 1)
			errors []source.SyntaxError
		)
		// parse first instruction
		return insn, errors := p.parseMacroInstruction(pc, env)
		// check real instruction parsed
		if insns[0] == nil {
			return Instruction{Instructions: nil}, errors
		}
		//
		for len(errors) == 0 && p.match(SEMICOLON) {
			i, errs := p.parsemacro.Instruction(pc, env)
			insns = append(insns, i)
			errors = append(errors, errs...)
		}
		//
		return Instruction{Instructions: insns}, errors
	}
*/
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
	case "var":
		return nil, p.parseVar(env)
	case "jz":
		insn, errs = p.parseJznz(env, true)
	case "jnz":
		insn, errs = p.parseJznz(env, false)
	case "jmp":
		insn, errs = p.parseJmp(env)
	case "ret":
		insn, errs = &micro.Ret{}, nil
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

func (p *Parser) parseJmp(env *Environment) (macro.Instruction, []source.SyntaxError) {
	lab, errs := p.parseIdentifier()
	//
	if len(errs) > 0 {
		return nil, errs
	}
	//
	return &micro.Jmp{
		Target: env.Bind(lab)}, nil
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
		env.DeclareRegister(insn.TEMP_REGISTER, name, width)
	}
	//
	return nil
}

func (p *Parser) parseJznz(env *Environment, sign bool) (macro.Instruction, []source.SyntaxError) {
	var (
		errs     []source.SyntaxError
		register string
		label    string
	)
	// save lookahead for error reporting
	lookahead := p.lookahead()
	// Parse register name
	if register, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	} else if !env.IsRegister(register) {
		return nil, p.syntaxErrors(lookahead, "unknown register")
	}
	// Parse target label
	if label, errs = p.parseIdentifier(); len(errs) > 0 {
		return nil, errs
	}
	//
	return &macro.Jznz{
		Sign:   sign,
		Source: env.LookupRegister(register),
		Target: env.Bind(label),
	}, nil
}

func (p *Parser) parseAssignment(env *Environment) (macro.Instruction, []source.SyntaxError) {
	var (
		lhs      []uint
		rhs      []uint
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
	// Parse right-hand side
	if kind, rhs, constant, errs = p.parseAssignmentRhs(env); len(errs) > 0 {
		return nil, errs
	}
	//
	switch kind {
	case ADD:
		insn = &micro.Add{Targets: lhs, Sources: rhs, Constant: constant}
	case SUB:
		insn = &micro.Sub{Targets: lhs, Sources: rhs, Constant: constant}
	case MUL:
		insn = &micro.Mul{Targets: lhs, Sources: rhs, Constant: constant}
	default:
		panic("unreachable")
	}
	// Done
	return insn, nil
}

func (p *Parser) parseAssignmentLhs(env *Environment) ([]uint, []source.SyntaxError) {
	var (
		lhs  []uint = make([]uint, 1)
		errs []source.SyntaxError
		reg  uint
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
	// Reverse items so that least significant comes first.
	lhs = util.Reverse(lhs)
	//
	return lhs, nil
}

func (p *Parser) parseAssignmentRhs(env *Environment) (uint, []uint, big.Int, []source.SyntaxError) {
	var (
		constant big.Int
		rhs      []uint
		kind     uint = ADD
		reg      uint
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

func (p *Parser) parseRegister(env *Environment) (uint, []source.SyntaxError) {
	lookahead := p.lookahead()
	reg, errs := p.parseIdentifier()
	//
	if len(errs) > 0 {
		return 0, errs
	} else if !env.IsRegister(reg) {
		return 0, p.syntaxErrors(lookahead, "unknown register")
	}
	// Done
	return env.LookupRegister(reg), nil
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

// Expect panics if the next token is not what was expected.
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

// Environment captures useful information used during the assembling process.
type Environment struct {
	// Labels identifies branch targets.
	labels []Label
	// Registers identifies set of declared registers.
	registers []Register
}

// Bind associates a label with a given index which can subsequently be used to
// determine a concrete program counter value.
func (p *Environment) Bind(name string) uint {
	// Check whether label already declared.
	for i, lab := range p.labels {
		if lab.name == name {
			return uint(i)
		}
	}
	// Determine index for new label
	index := uint(len(p.labels))
	// Create new label
	p.labels = append(p.labels, UnboundLabel(name))
	// Done
	return index
}

// DeclareLabel declares a given label at a given program counter position.  If
// a label with the same name already exists, this will panic.
func (p *Environment) DeclareLabel(name string, pc uint) {
	// First, check whether the label already exists
	for i, lab := range p.labels {
		if lab.name == name {
			if lab.pc == math.MaxUint {
				p.labels[i].pc = pc
				return
			}
			//
			panic("label already bound")
		}
	}
	// Create new label
	p.labels = append(p.labels, BoundLabel(name, pc))
}

// DeclareRegister declares a new register with the given name and bitwidth.  If
// a register with the same name already exists, this panics.
func (p *Environment) DeclareRegister(kind uint8, name string, width uint) {
	if p.IsRegister(name) {
		panic(fmt.Sprintf("register %s already declared", name))
	}
	//
	p.registers = append(p.registers, instruction.NewRegister(kind, name, width))
}

// IsRegister checks whether or not a given name is already declared as a
// register.
func (p *Environment) IsRegister(name string) bool {
	for _, reg := range p.registers {
		if reg.Name == name {
			return true
		}
	}
	//
	return false
}

// IsBoundLabel checks whether or not a given label has already been bound to a
// given PC.
func (p *Environment) IsBoundLabel(name string) bool {
	for _, l := range p.labels {
		if l.name == name && l.pc != math.MaxUint {
			return true
		}
	}
	//
	return false
}

// LookupRegister looks up the index for a given register.
func (p *Environment) LookupRegister(name string) uint {
	for i, reg := range p.registers {
		if reg.Name == name {
			return uint(i)
		}
	}
	//
	panic(fmt.Sprintf("unknown register %s", name))
}

// BindLabels processes a given set of instructions by mapping their label
// indexes to concrete program counter locations.
func (p *Environment) BindLabels(insns []macro.Instruction) {
	labels := make([]uint, len(p.labels))
	// Initial the label map
	for i := range labels {
		labels[i] = p.labels[i].pc
		// sanity check
		if labels[i] == math.MaxUint {
			panic(fmt.Sprintf("unbound label \"%s\"", p.labels[i].name))
		}
	}
	// Bind labels using the map
	for _, insn := range insns {
		insn.Bind(labels)
	}
}
