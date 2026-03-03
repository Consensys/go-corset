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
package codegen

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Condition is a convenient alias
type Condition = expr.Condition[symbol.Resolved]

// Expr is a convenient alias
type Expr = expr.Expr[symbol.Resolved]

// Compiler provides a working environment for compiling individual statements
// within a given function.  For example, it provides the ability to allocate
// new temporary registers as required.
type Compiler struct {
	components []ast.Declaration
	variables  []variable.Descriptor
	registers  []register.Register
}

func (p *Compiler) lookup(id symbol.Resolved) ast.Expr {
	// Expecting this to be a constant
	if c, ok := p.components[id.Index].(*ast.Constant); ok {
		return c.ConstExpr
	}
	//
	panic(fmt.Sprintf("unknown constant %s", id.Name))
}

func (p *Compiler) compileStatement(pc uint, s ast.Stmt) Instruction {
	var insns []MicroInstruction
	//
	switch s := s.(type) {
	case *stmt.Assign[symbol.Resolved]:
		targets := mapRegisters(s.Targets)
		insns = p.compileExpr(s.Source, targets...)
	case *stmt.IfGoto[symbol.Resolved]:
		return p.compileCondition(pc, s.Cond, s.Target)
	case *stmt.Goto[symbol.Resolved]:
		return &instruction.Jmp{Target: s.Target}
	case *stmt.Fail[symbol.Resolved]:
		return &instruction.Fail{}
	case *stmt.Return[symbol.Resolved]:
		return &instruction.Return{}
	default:
		panic("unknown statement encountered")
	}
	//
	return instruction.NewVector[word.Uint](insns...)
}

func (p *Compiler) compileCondition(pc uint, e Condition, target uint) Instruction {
	var (
		insns []MicroInstruction
		args  []register.Id
	)
	//
	switch e := e.(type) {
	case *expr.Cmp[symbol.Resolved]:
		args, insns = p.compileArgs(e.Left, e.Right)
		insns = append(insns, instruction.NewSkipIf(instruction.Condition(e.Operator), args[0], args[1], 1))
		insns = append(insns, instruction.NewJmp(pc+1))
		insns = append(insns, instruction.NewJmp(target))
	default:
		panic("unknown condition encountered")
	}
	//
	return instruction.NewVector[word.Uint](insns...)
}

func (p *Compiler) compileExpr(e Expr, targets ...register.Id) []MicroInstruction {
	var (
		zero  word.Uint
		insns []MicroInstruction
		insn  MicroInstruction
	)
	//
	switch e := e.(type) {
	case *expr.Add[symbol.Resolved]:
		insns, insn = p.compileAdd(e.Exprs, targets)
	case *expr.Const[symbol.Resolved]:
		var c word.Uint
		//
		insn = instruction.NewAdd[word.Uint](targets, nil, c.SetBigInt(&e.Constant))
	case *expr.LocalAccess[symbol.Resolved]:
		var reg = []register.Id{register.NewId(e.Variable)}
		//
		insn = instruction.NewAdd[word.Uint](targets, reg, zero)
	case *expr.Mul[symbol.Resolved]:
		insns, insn = p.compileMul(e.Exprs, targets)
	case *expr.NonLocalAccess[symbol.Resolved]:
		insn = instruction.NewAdd[word.Uint](targets, nil, p.evalConstant(e))
	default:
		panic("unknown expression encountered")
	}
	//
	return append(insns, insn)
}

func (p *Compiler) compileAdd(args []Expr, targets []register.Id,
) ([]MicroInstruction, MicroInstruction) {
	//
	var (
		constant word.Uint
		nargs    []Expr
		w        word.Uint
	)
	//
	for _, e := range args {
		if c, ok := e.(*expr.Const[symbol.Resolved]); ok {
			constant = constant.Add(w.SetBigInt(&c.Constant))
		} else if _, ok := e.(*expr.NonLocalAccess[symbol.Resolved]); ok {
			constant = constant.Add(p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(nargs...)
	// Done
	return insns, instruction.NewAdd[word.Uint](targets, sources, constant)
}

func (p *Compiler) compileMul(args []Expr, targets []register.Id,
) ([]MicroInstruction, MicroInstruction) {
	//
	var (
		constant word.Uint = word.Uint64[word.Uint](1)
		nargs    []Expr
		w        word.Uint
	)
	//
	for _, e := range args {
		if c, ok := e.(*expr.Const[symbol.Resolved]); ok {
			constant = constant.Mul(w.SetBigInt(&c.Constant))
		} else if _, ok := e.(*expr.NonLocalAccess[symbol.Resolved]); ok {
			constant = constant.Mul(p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(nargs...)
	// Done
	return insns, instruction.NewMul[word.Uint](targets, sources, constant)
}

func (p *Compiler) compileArgs(exprs ...Expr) ([]register.Id, []MicroInstruction) {
	var (
		insns   []MicroInstruction
		targets = make([]register.Id, len(exprs))
	)
	//
	for i, e := range exprs {
		// Determine width of expression
		var bitwidth = e.BitWidth()
		//
		if r, ok := e.(*expr.LocalAccess[symbol.Resolved]); ok {
			targets[i] = register.NewId(r.Variable)
		} else {
			// Allocate temporary variable
			targets[i] = p.allocate(bitwidth)
			// Compile expression, storing result in temporary
			insns = append(insns, p.compileExpr(e, targets[i])...)
		}
	}
	//
	return targets, insns
}

func (p *Compiler) evalConstant(e ast.Expr) word.Uint {
	switch e := e.(type) {
	case *expr.Add[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		return word.Sum(args...)
	case *expr.Const[symbol.Resolved]:
		var c word.Uint
		//
		return c.SetBigInt(&e.Constant)
	case *expr.Mul[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		return word.Product(args...)
	case *expr.NonLocalAccess[symbol.Resolved]:
		return p.evalConstant(p.lookup(e.Name))
	default:
		panic("unknown expression encountered")
	}
}

func (p *Compiler) evalConstants(es []ast.Expr) []word.Uint {
	var words = make([]word.Uint, len(es))
	//
	for i, e := range es {
		words[i] = p.evalConstant(e)
	}
	//
	return words
}

func (p *Compiler) allocate(bitwidth uint) register.Id {
	var (
		padding big.Int
		n       = len(p.registers)
		name    = fmt.Sprintf("$%d", n)
	)
	//
	p.registers = append(p.registers, register.NewComputed(name, bitwidth, padding))
	//
	return register.NewId(uint(n))
}

func mapRegisters(variables []variable.Id) []register.Id {
	var regs = make([]register.Id, len(variables))
	//
	for i, v := range variables {
		regs[i] = register.NewId(v)
	}
	//
	return regs
}
