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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Compiler provides a working environment for compiling individual statements
// within a given function.  For example, it provides the ability to allocate
// new temporary registers as required.
type Compiler struct {
	variables []variable.Descriptor
	registers []register.Register
}

func (p *Compiler) compileStatement(pc uint, s ast.Instruction) Instruction {
	var insns []MicroInstruction
	//
	switch s := s.(type) {
	case *stmt.Assign[ast.ResolvedSymbol]:
		targets := mapRegisters(s.Targets)
		insns = p.compileExpr(s.Source, targets...)
	case *stmt.IfGoto[ast.ResolvedSymbol]:
		return p.compileCondition(pc, s.Cond, s.Target)
	case *stmt.Goto[ast.ResolvedSymbol]:
		return &instruction.Jmp{Target: s.Target}
	case *stmt.Fail[ast.ResolvedSymbol]:
		return &instruction.Fail{}
	case *stmt.Return[ast.ResolvedSymbol]:
		return &instruction.Return{}
	default:
		panic("unknown statement encountered")
	}
	//
	return instruction.NewVector[word.Uint](insns...)
}

func (p *Compiler) compileCondition(pc uint, e expr.Condition, target uint) Instruction {
	var (
		insns []MicroInstruction
		args  []register.Id
	)
	//
	switch e := e.(type) {
	case *expr.Cmp:
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

func (p *Compiler) compileExpr(e expr.Expr, targets ...register.Id) []MicroInstruction {
	var (
		zero  word.Uint
		insns []MicroInstruction
		insn  MicroInstruction
	)
	//
	switch e := e.(type) {
	case *expr.Add:
		insns, insn = p.compileAdd(e.Exprs, targets)
	case *expr.Const:
		var c word.Uint
		//
		insn = instruction.NewAdd[word.Uint](targets, nil, c.SetBigInt(&e.Constant))
	case *expr.Mul:
		insns, insn = p.compileMul(e.Exprs, targets)
	case *expr.VarAccess:
		var reg = []register.Id{register.NewId(e.Variable)}
		//
		insn = instruction.NewAdd[word.Uint](targets, reg, zero)
	}
	//
	return append(insns, insn)
}

func (p *Compiler) compileAdd(args []expr.Expr, targets []register.Id,
) ([]MicroInstruction, MicroInstruction) {
	//
	var (
		constant word.Uint
		nargs    []expr.Expr
		w        word.Uint
	)
	//
	for _, e := range args {
		if c, ok := e.(*expr.Const); ok {
			constant = constant.Add(w.SetBigInt(&c.Constant))
		} else {
			nargs = append(nargs, e)
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(nargs...)
	// Done
	return insns, instruction.NewAdd[word.Uint](targets, sources, constant)
}

func (p *Compiler) compileMul(args []expr.Expr, targets []register.Id,
) ([]MicroInstruction, MicroInstruction) {
	//
	var (
		constant word.Uint = word.Uint64[word.Uint](1)
		nargs    []expr.Expr
		w        word.Uint
	)
	//
	for _, e := range args {
		if c, ok := e.(*expr.Const); ok {
			constant = constant.Mul(w.SetBigInt(&c.Constant))
		} else {
			nargs = append(nargs, e)
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(nargs...)
	// Done
	return insns, instruction.NewMul[word.Uint](targets, sources, constant)
}
func (p *Compiler) compileArgs(exprs ...expr.Expr) ([]register.Id, []MicroInstruction) {
	var (
		insns   []MicroInstruction
		targets = make([]register.Id, len(exprs))
	)
	//
	for i, e := range exprs {
		// Determine width of expression
		var bitwidth, signed = expr.BitWidth(e, p.variableMap())
		//
		if signed {
			panic("handle signed expressions")
		} else if r, ok := e.(*expr.VarAccess); ok {
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

func (p *Compiler) variableMap() variable.Map {
	return variable.ArrayMap(p.variables...)
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
