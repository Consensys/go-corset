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
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/lval"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/util"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Condition is a convenient alias
type Condition = expr.Condition[symbol.Resolved]

// Expr is a convenient alias
type Expr = expr.Expr[symbol.Resolved]

// LVal is a convenient alias
type LVal = lval.LVal[symbol.Resolved]

// Compiler provides a working environment for compiling individual statements
// within a given function.  For example, it provides the ability to allocate
// new temporary registers as required.
type Compiler struct {
	components  []Declaration
	variables   []VariableDescriptor
	registers   []register.Register
	environment data.Environment[symbol.Resolved]
	srcmaps     source.Maps[any]
	errors      []source.SyntaxError
}

func (p *Compiler) compileStatement(pc uint, mapping []uint, s Stmt) Instruction {
	var insns []MicroInstruction
	//
	switch s := s.(type) {
	case *stmt.Assign[symbol.Resolved]:
		targets, pre, post := p.mapLVals(mapping, s.Targets)
		//
		insns = p.compileExpr(s.Source, mapping, targets...)
		// Configure pre/post instructions
		insns = append(pre, insns...)
		insns = append(insns, post...)
	case *stmt.IfGoto[symbol.Resolved]:
		return p.compileCondition(pc, s.Cond, mapping, s.Target)
	case *stmt.Goto[symbol.Resolved]:
		return &instruction.Jmp[word.Uint]{Target: s.Target}
	case *stmt.Fail[symbol.Resolved]:
		return &instruction.Fail[word.Uint]{}
	case *stmt.Printf[symbol.Resolved]:
		return p.compilePrintf(mapping, s.Chunks, s.Arguments)
	case *stmt.Return[symbol.Resolved]:
		return &instruction.Return[word.Uint]{}
	default:
		panic("unknown statement encountered")
	}
	//
	return instruction.NewVector[word.Uint](insns...)
}

func (p *Compiler) mapLVals(mapping []uint, lvals []LVal) ([]register.Id, []MicroInstruction, []MicroInstruction) {
	var (
		regs                = make([]register.Id, len(lvals))
		preInsns, postInsns []MicroInstruction
	)
	//
	for i, lv := range lvals {
		switch lv := lv.(type) {
		case *lval.Variable[symbol.Resolved]:
			regs[i] = register.NewId(lv.Id)
		case *lval.MemAccess[symbol.Resolved]:
			var (
				ext = p.components[lv.Name.Index].(*Memory)
				// Determine vm module identifier
				id = mapping[lv.Name.Index]
			)
			if !ext.IsWriteable() {
				panic(fmt.Sprintf("unreadable memory \"%s\" encountered", ext.Name()))
			}
			//
			sources := make([]register.Id, len(ext.Data))
			targets, pre := p.compileArgs(mapping, lv.Args...)
			// Sanity check (for now)
			if len(ext.Data) > 1 {
				panic("multiple data lines not (currently) supported")
			}
			// Allocate data lines as needed
			for j, t := range ext.Data {
				bitwidth := data.BitWidthOf(t.DataType, p.environment)
				sources[j] = p.allocate(bitwidth)
				// FIXME: broken when len(ext.Data) > 1
				regs[i+j] = sources[j]
			}
			//
			preInsns = append(preInsns, pre...)
			postInsns = append(postInsns, instruction.NewMemWrite[word.Uint](id, targets, sources))
		}
	}
	//
	return regs, preInsns, postInsns
}

func (p *Compiler) compilePrintf(mapping []uint, chunks []stmt.FormattedChunk, args []Expr) Instruction {
	var (
		nchunks     []instruction.FormattedChunk
		regs, insns = p.compileArgs(mapping, args...)
		index       uint
	)
	//

	// Manage all chunks
	for _, chunk := range chunks {
		if chunk.Format.HasFormat() {
			nchunks = append(nchunks, instruction.FormattedChunk{
				Text: chunk.Text, Format: chunk.Format, Argument: regs[index],
			})
			//
			index++
		} else {
			nchunks = append(nchunks, instruction.FormattedChunk{
				Text: chunk.Text, Format: util.EMPTY_FORMAT, Argument: register.UnusedId(),
			})
		}
	}
	//
	insns = append(insns, &instruction.Debug[word.Uint]{Chunks: nchunks})
	//
	return instruction.NewVector[word.Uint](insns...)
}

func (p *Compiler) compileCondition(pc uint, e Condition, mapping []uint, target uint) Instruction {
	var (
		insns []MicroInstruction
		args  []register.Id
	)
	//
	switch e := e.(type) {
	case *expr.Cmp[symbol.Resolved]:
		args, insns = p.compileArgs(mapping, e.Left, e.Right)
		insns = append(insns, instruction.NewSkipIf[word.Uint](instruction.Condition(e.Operator), args[0], args[1], 1))
		insns = append(insns, instruction.NewJmp[word.Uint](pc+1))
		insns = append(insns, instruction.NewJmp[word.Uint](target))
	default:
		panic("unknown condition encountered")
	}
	//
	return instruction.NewVector[word.Uint](insns...)
}

func (p *Compiler) compileExpr(e Expr, mapping []uint, targets ...register.Id) []MicroInstruction {
	var (
		zero     word.Uint
		insns    []MicroInstruction
		insn     MicroInstruction
		unitExpr = false
	)
	//
	switch e := e.(type) {
	case *expr.Cast[symbol.Resolved]:
		insns, insn = p.compileCast(e, mapping, targets[0])
		unitExpr = true
	case *expr.Add[symbol.Resolved]:
		insns, insn = p.compileAdd(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.BitwiseAnd[symbol.Resolved]:
		insns, insn = p.compileAnd(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Const[symbol.Resolved]:
		var c word.Uint
		//
		insn = instruction.NewAdd[word.Uint](targets[0], nil, c.SetBigInt(&e.Constant))
		unitExpr = true
	case *expr.ExternAccess[symbol.Resolved]:
		//
		switch ext := p.components[e.Name.Index].(type) {
		case *Constant:
			insn = instruction.NewAdd[word.Uint](targets[0], nil, p.evalConstant(e))
			unitExpr = true
		case *Memory:
			if !ext.IsReadable() {
				panic(fmt.Sprintf("unreadable memory \"%s\" encountered", e.Name.String()))
			}
			//
			insns, insn = p.compileMemoryRead(e, ext, mapping, targets...)
		case *Function:
			insns, insn = p.compileFunctionCall(e, ext, mapping, targets...)
		default:
			panic(fmt.Sprintf("unknown symbol \"%s\" encountered", e.Name.String()))
		}
	case *expr.LocalAccess[symbol.Resolved]:
		var reg = []register.Id{register.NewId(e.Variable)}
		//
		insn = instruction.NewAdd[word.Uint](targets[0], reg, zero)
		unitExpr = true
	case *expr.Mul[symbol.Resolved]:
		insns, insn = p.compileMul(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.BitwiseNot[symbol.Resolved]:
		insns, insn = p.compileNot(e, mapping, targets[0])
		unitExpr = true
	case *expr.BitwiseOr[symbol.Resolved]:
		insns, insn = p.compileOr(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Div[symbol.Resolved]:
		insns, insn = p.compileDiv(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Rem[symbol.Resolved]:
		insns, insn = p.compileRem(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Shl[symbol.Resolved]:
		insns, insn = p.compileShl(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Shr[symbol.Resolved]:
		insns, insn = p.compileShr(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Sub[symbol.Resolved]:
		insns, insn = p.compileSub(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Xor[symbol.Resolved]:
		insns, insn = p.compileXor(e.Exprs, mapping, targets[0])
		unitExpr = true
	case *expr.Ternary[symbol.Resolved]:
		insns, insn = p.compileTernary(e, mapping, targets[0])
		unitExpr = true
	default:
		panic("unknown expression encountered")
	}
	//
	if unitExpr && len(targets) > 1 {
		panic("incorrect arity for unit expression")
	}
	//
	return append(insns, insn)
}

func (p *Compiler) compileTernary(e *expr.Ternary[symbol.Resolved], mapping []uint, target register.Id,
) ([]MicroInstruction, MicroInstruction) {
	cmp := e.Cond.(*expr.Cmp[symbol.Resolved])
	// Eagerly evaluate both branches into temporaries.
	trueRegs, trueInsns := p.compileArgs(mapping, e.IfTrue)
	falseRegs, falseInsns := p.compileArgs(mapping, e.IfFalse)
	// Evaluate condition operands.
	condRegs, condInsns := p.compileArgs(mapping, cmp.Left, cmp.Right)
	// Selection sequence:
	//   skip_if(cond, lhs, rhs, 2)      if TRUE skip false-copy + skip(1)
	//   add(target, [falseReg], 0)       false branch (skipped when TRUE)
	//   skip(1)                          skip over true branch
	//   add(target, [trueReg], 0)        true branch  (returned as final insn)
	var zero word.Uint

	insns := append(trueInsns, falseInsns...)
	insns = append(insns, condInsns...)
	insns = append(insns, instruction.NewSkipIf[word.Uint](
		instruction.Condition(cmp.Operator), condRegs[0], condRegs[1], 2))
	insns = append(insns, instruction.NewAdd[word.Uint](target, []register.Id{falseRegs[0]}, zero))
	insns = append(insns, &instruction.Skip[word.Uint]{Skip: 1})

	return insns, instruction.NewAdd[word.Uint](target, []register.Id{trueRegs[0]}, zero)
}

func (p *Compiler) compileCast(e *expr.Cast[symbol.Resolved], mapping []uint, target register.Id,
) ([]MicroInstruction, MicroInstruction) {
	castWidth := e.CastType.AsUint(p.environment).BitWidth()
	sources, insns := p.compileArgs(mapping, e.Expr)
	//
	return insns, instruction.NewCast[word.Uint](target, sources[0], castWidth)
}

func (p *Compiler) compileAdd(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	//
	var (
		constant word.Uint
		nargs    []Expr
		w        word.Uint
		bitwidth = p.registers[target.Unwrap()].Width()
	)
	//
	for _, e := range args {
		var overflow bool
		//
		if c, ok := e.(*expr.Const[symbol.Resolved]); ok {
			constant, overflow = constant.Add(bitwidth, w.SetBigInt(&c.Constant))
		} else if p.isConstantAccess(e) {
			constant, overflow = constant.Add(bitwidth, p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
		// NOTE: this error should be caught and reported earlier in the
		// pipeline.
		if overflow {
			panic("compileAdd arithmetic overflow")
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(mapping, nargs...)
	// Done
	return insns, instruction.NewAdd[word.Uint](target, sources, constant)
}

func (p *Compiler) compileFunctionCall(e *expr.ExternAccess[symbol.Resolved], fn *Function, mapping []uint,
	targets ...register.Id) ([]MicroInstruction, MicroInstruction) {
	// Determine vm module identifier
	var id = mapping[e.Name.Index]
	// Compile arguments
	sources, insns := p.compileArgs(mapping, e.Args...)
	// determine type of read
	return insns, instruction.NewCall[word.Uint](id, targets, sources)
}

func (p *Compiler) compileMemoryRead(e *expr.ExternAccess[symbol.Resolved], mem *Memory, mapping []uint,
	targets ...register.Id) ([]MicroInstruction, MicroInstruction) {
	// Determine vm module identifier
	var id = mapping[e.Name.Index]
	// Compile arguments
	sources, insns := p.compileArgs(mapping, e.Args...)
	// determine type of read
	return insns, instruction.NewMemRead[word.Uint](id, targets, sources)
}

func (p *Compiler) compileMul(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	//
	var (
		constant word.Uint = word.Uint64[word.Uint](1)
		nargs    []Expr
		w        word.Uint
		bitwidth = p.registers[target.Unwrap()].Width()
	)
	//
	for _, e := range args {
		var overflow bool
		//
		if c, ok := e.(*expr.Const[symbol.Resolved]); ok {
			constant, overflow = constant.Mul(bitwidth, w.SetBigInt(&c.Constant))
		} else if p.isConstantAccess(e) {
			constant, overflow = constant.Mul(bitwidth, p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
		// NOTE: this error should be caught and reported earlier in the
		// pipeline.
		if overflow {
			panic("compileMul arithmetic overflow")
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(mapping, nargs...)
	// Done
	return insns, instruction.NewMul[word.Uint](target, sources, constant)
}

func (p *Compiler) compileDiv(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	// Compile all operands upfront.
	sources, insns := p.compileArgs(mapping, args...)
	// Chain divisions left-to-right: (((a / b) / c) / ...).
	value := sources[0]
	//
	for i := 1; i < len(sources)-1; i++ {
		tmp := p.allocate(p.registers[target.Unwrap()].Width())
		insns = append(insns, instruction.NewDiv[word.Uint](tmp, value, sources[i]))
		value = tmp
	}
	//
	return insns, instruction.NewDiv[word.Uint](target, value, sources[len(sources)-1])
}

func (p *Compiler) compileRem(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	// Compile all operands upfront.
	sources, insns := p.compileArgs(mapping, args...)
	// Chain remainders left-to-right: (((a % b) % c) % ...).
	value := sources[0]
	//
	for i := 1; i < len(sources)-1; i++ {
		tmp := p.allocate(p.registers[target.Unwrap()].Width())
		insns = append(insns, instruction.NewRem[word.Uint](tmp, value, sources[i]))
		value = tmp
	}
	//
	return insns, instruction.NewRem[word.Uint](target, value, sources[len(sources)-1])
}

func (p *Compiler) compileShl(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	// Compile all operands upfront.
	sources, insns := p.compileArgs(mapping, args...)
	// Chain shifts left-to-right: (((a << b) << c) << ...).
	value := sources[0]
	//
	for i := 1; i < len(sources)-1; i++ {
		tmp := p.allocate(p.registers[target.Unwrap()].Width())
		insns = append(insns, instruction.NewShl[word.Uint](tmp, value, sources[i]))
		value = tmp
	}
	//
	return insns, instruction.NewShl[word.Uint](target, value, sources[len(sources)-1])
}

func (p *Compiler) compileShr(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	// Compile all operands upfront.
	sources, insns := p.compileArgs(mapping, args...)
	// Chain shifts left-to-right: (((a >> b) >> c) >> ...).
	value := sources[0]
	//
	for i := 1; i < len(sources)-1; i++ {
		tmp := p.allocate(p.registers[target.Unwrap()].Width())
		insns = append(insns, instruction.NewShr[word.Uint](tmp, value, sources[i]))
		value = tmp
	}
	//
	return insns, instruction.NewShr[word.Uint](target, value, sources[len(sources)-1])
}

func (p *Compiler) compileSub(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	//
	var (
		constant word.Uint
		nargs    []Expr
		w        word.Uint
		bitwidth = p.registers[target.Unwrap()].Width()
	)
	//
	for i, e := range args {
		var overflow bool

		if c, ok := e.(*expr.Const[symbol.Resolved]); ok && i > 0 {
			constant, overflow = constant.Add(bitwidth, w.SetBigInt(&c.Constant))
		} else if p.isConstantAccess(e) && i > 0 {
			constant, overflow = constant.Add(bitwidth, p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
		// NOTE: this error should be caught and reported earlier in the
		// pipeline.
		if overflow {
			panic("compileSub arithmetic overflow")
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(mapping, nargs...)
	// Done
	return insns, instruction.NewSub[word.Uint](target, sources, constant)
}

func (p *Compiler) compileAnd(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	var (
		bitwidth = p.registers[target.Unwrap()].Width()
		// Identity for AND is all-ones within the target bitwidth.
		constant word.Uint
		nargs    []Expr
	)
	// Start with all-ones (identity for AND).
	constant = constant.Not(bitwidth)
	//
	for _, e := range args {
		if c, ok := e.(*expr.Const[symbol.Resolved]); ok {
			var w word.Uint

			constant = constant.And(bitwidth, w.SetBigInt(&c.Constant))
		} else if p.isConstantAccess(e) {
			constant = constant.And(bitwidth, p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(mapping, nargs...)
	//
	return insns, instruction.NewAnd[word.Uint](target, sources, constant)
}

func (p *Compiler) compileNot(e *expr.BitwiseNot[symbol.Resolved], mapping []uint, target register.Id,
) ([]MicroInstruction, MicroInstruction) {
	sources, insns := p.compileArgs(mapping, e.Expr)
	//
	return insns, instruction.NewNot[word.Uint](target, sources[0])
}

func (p *Compiler) compileOr(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	var (
		bitwidth = p.registers[target.Unwrap()].Width()
		constant word.Uint
		nargs    []Expr
	)
	//
	for _, e := range args {
		if c, ok := e.(*expr.Const[symbol.Resolved]); ok {
			var w word.Uint

			constant = constant.Or(bitwidth, w.SetBigInt(&c.Constant))
		} else if p.isConstantAccess(e) {
			constant = constant.Or(bitwidth, p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(mapping, nargs...)
	//
	return insns, instruction.NewOr[word.Uint](target, sources, constant)
}

func (p *Compiler) compileXor(args []Expr, mapping []uint, target register.Id) ([]MicroInstruction, MicroInstruction) {
	var (
		bitwidth = p.registers[target.Unwrap()].Width()
		constant word.Uint
		nargs    []Expr
	)
	//
	for _, e := range args {
		if c, ok := e.(*expr.Const[symbol.Resolved]); ok {
			var w word.Uint

			constant = constant.Xor(bitwidth, w.SetBigInt(&c.Constant))
		} else if p.isConstantAccess(e) {
			constant = constant.Xor(bitwidth, p.evalConstant(e))
		} else {
			nargs = append(nargs, e)
		}
	}
	// Compile arguments
	sources, insns := p.compileArgs(mapping, nargs...)
	//
	return insns, instruction.NewXor[word.Uint](target, sources, constant)
}

func (p *Compiler) compileArgs(mapping []uint, exprs ...Expr) ([]register.Id, []MicroInstruction) {
	var (
		insns   []MicroInstruction
		targets = make([]register.Id, len(exprs))
	)
	//
	for i, e := range exprs {
		//
		if r, ok := e.(*expr.LocalAccess[symbol.Resolved]); ok {
			targets[i] = register.NewId(r.Variable)
		} else {
			bitwidth := data.BitWidthOf(e.Type(), p.environment)
			// Allocate temporary variable
			targets[i] = p.allocate(bitwidth)
			// Compile expression, storing result in temporary
			insns = append(insns, p.compileExpr(e, mapping, targets[i])...)
		}
	}
	//
	return targets, insns
}

func (p *Compiler) evalConstant(e Expr) word.Uint {
	bitwidth := data.BitWidthOf(e.Type(), p.environment)
	//
	switch e := e.(type) {
	case *expr.Add[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		res, overflow := word.Sum(bitwidth, args...)
		// TODO: report a proper error
		if overflow {
			panic("evalConstantAdd arithmetic overflow")
		}
		//
		return res
	case *expr.BitwiseAnd[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		return word.BitwiseAnd(bitwidth, args...)
	case *expr.Const[symbol.Resolved]:
		var c word.Uint
		//
		return c.SetBigInt(&e.Constant)
	case *expr.Mul[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		res, overflow := word.Product(bitwidth, args...)
		// TODO: report a proper error
		if overflow {
			panic("evalConstantMul arithmetic overflow")
		}
		//
		return res
	case *expr.BitwiseNot[symbol.Resolved]:
		arg := p.evalConstant(e.Expr)
		return arg.Not(bitwidth)
	case *expr.BitwiseOr[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		return word.BitwiseOr(bitwidth, args...)
	case *expr.Shl[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		return word.BitwiseShl(bitwidth, args...)
	case *expr.Shr[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		return word.BitwiseShr(bitwidth, args...)
	case *expr.Xor[symbol.Resolved]:
		args := p.evalConstants(e.Exprs)
		return word.BitwiseXor(bitwidth, args...)
	case *expr.Cast[symbol.Resolved]:
		inner := p.evalConstant(e.Expr)
		width := e.CastType.AsUint(p.environment).BitWidth()

		sliced := inner.Slice(width)
		if inner.Cmp(sliced) != 0 {
			p.errors = append(p.errors, p.srcmaps.SyntaxErrors(e, "cast overflow")...)
		}

		return sliced
	case *expr.ExternAccess[symbol.Resolved]:
		var decl = p.components[e.Name.Index].(*Constant)
		return p.evalConstant(decl.ConstExpr)
	default:
		panic("unknown expression encountered")
	}
}

func (p *Compiler) evalConstants(es []Expr) []word.Uint {
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

func (p *Compiler) isConstantAccess(e Expr) bool {
	ne, ok := e.(*expr.ExternAccess[symbol.Resolved])
	//
	if !ok {
		return false
	}
	// Check whethe ris constant
	_, ok = p.components[ne.Name.Index].(*Constant)
	//
	return ok
}
