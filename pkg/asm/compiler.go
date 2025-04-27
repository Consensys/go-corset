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
	"math/big"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/asm/instruction"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
)

var zero = *big.NewInt(0)
var one = *big.NewInt(1)

// CompileAssembly compiles a given set of assembly functions into a binary
// constraint file.
func CompileAssembly(assembly ...source.File) (*binfile.BinaryFile, []source.SyntaxError) {
	functions, errs := Assemble(assembly...)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	//
	return NewCompiler().Compile(functions...)
}

// Compiler packages up everything needed to compile a given assembly down into
// an HIR schema.  Observe that the compiler may fail if the assembly files are
// malformed in some way (e.g. fail type checking).
type Compiler struct {
	schema hir.Schema
	// maxInstances determines the maximum number of instances permitted for any
	// given function.
	maxInstances uint
	// types & reftables
	// sourcemap
}

// NewCompiler constructs a new compiler
func NewCompiler() *Compiler {
	return &Compiler{
		schema:       *hir.EmptySchema(),
		maxInstances: 32,
	}
}

// Compile compiles a given set of functions into a binary file.
func (p *Compiler) Compile(functions ...Function) (*binfile.BinaryFile, []source.SyntaxError) {
	for i := range functions {
		p.compileFunction(uint(i), functions)
	}

	return binfile.NewBinaryFile(nil, nil, &p.schema), nil
}

func (p *Compiler) compileFunction(id uint, functions []Function) {
	var (
		fn = functions[id]
		// Allocate module id
		mid = p.schema.AddModule(fn.Name)
		ctx = trace.NewContext(mid, 1)
		// Map fn registers to schema columns
		rids = make([]uint, len(fn.Registers))
	)
	// Allocate registers as columns
	for i, reg := range fn.Registers {
		typeName := fmt.Sprintf("%s:u%d", reg.Name, reg.Width)
		// Construct appropriate datatype
		datatype := schema.NewUintType(reg.Width)
		// Allocate register
		rids[i] = p.schema.AddDataColumn(ctx, reg.Name, datatype)
		// Add range constraint
		p.schema.AddRangeConstraint(typeName, ctx, hir.NewColumnAccess(rids[i], 0), datatype.Bound())
	}
	// Setup framing columns / constraints
	pcid := p.initFunctionFraming(ctx, fn)
	//
	for i, insn := range fn.Code {
		p.translateInsn(uint(i), pcid, ctx, rids, fn.Registers, insn)
	}
}

func (p *Compiler) initFunctionFraming(ctx trace.Context, fn Function) uint {
	// Determine max width of PC
	pcMax := uint64(len(fn.Code) - 1)
	pcWidth := uint(big.NewInt(int64(pcMax)).BitLen())
	// Allocate book keeping columns
	stamp := p.schema.AddDataColumn(ctx, "$stamp", schema.NewUintType(p.maxInstances))
	pc := p.schema.AddDataColumn(ctx, "$pc", schema.NewUintType(pcWidth))
	//
	stamp_i := hir.NewColumnAccess(stamp, 0)
	stamp_ip1 := hir.NewColumnAccess(stamp, 1)
	pc_i := hir.NewColumnAccess(pc, 0)
	pc_ip1 := hir.NewColumnAccess(pc, 1)
	// $stamp == 0 on first row
	p.schema.AddVanishingConstraint("first", ctx, util.Some(0), hir.Equals(stamp_i, hir.ZERO))
	// $stamp == 0 || $pc == pc_max on last row [BROKEN]
	p.schema.AddVanishingConstraint("last", ctx, util.Some(-1),
		hir.Disjunction(hir.Equals(stamp_i, hir.ZERO), hir.Equals(pc_i, hir.NewConst64(pcMax))))
	// next($stamp) == $stamp || next($stamp) == $stamp+1
	p.schema.AddVanishingConstraint("increment", ctx, util.None[int](),
		hir.Disjunction(hir.Equals(stamp_ip1, stamp_i), hir.Equals(stamp_ip1, hir.Sum(hir.ONE, stamp_i))))
	// next($stamp) == $stamp || next($pc) == 0
	p.schema.AddVanishingConstraint("reset", ctx, util.None[int](),
		hir.Disjunction(hir.Equals(stamp_ip1, stamp_i), hir.Equals(pc_ip1, hir.ZERO)))
	//
	return pc
}

func (p *Compiler) translateInsn(pc uint, pcid uint, ctx trace.Context, rids []uint, regs []Register,
	insn Instruction) {
	//
	switch insn := insn.(type) {
	case *instruction.Add:
		p.translateAddInsn(pc, pcid, ctx, rids, regs, insn)
	case *instruction.Jmp:
		p.translateJmpInsn(pc, pcid, ctx, rids, regs, insn)
	case *instruction.Jznz:
		if insn.Sign {
			p.translateJzInsn(pc, pcid, ctx, rids, regs, insn)
		} else {
			p.translateJnzInsn(pc, pcid, ctx, rids, regs, insn)
		}
	case *instruction.Mul:
		p.translateMulInsn(pc, pcid, ctx, rids, regs, insn)
	case *instruction.Ret:
		p.translateRetInsn(pc, pcid, ctx)
	case *instruction.Sub:
		p.translateSubInsn(pc, pcid, ctx, rids, regs, insn)
	default:
		panic("unknown instruction encountered")
	}
}

func (p *Compiler) translateAddInsn(pc uint, pcid uint, ctx trace.Context, rids []uint, regs []Register,
	insn *instruction.Add) {
	//
	var (
		name  = fmt.Sprintf("pc%d_add", pc)
		pc_i  = hir.NewColumnAccess(pcid, 0)
		guard = hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
	)
	// build up the lhs
	lhs := p.buildAssignmentLhs(insn.Targets, rids, regs)
	// build up the rhs
	rhs := p.buildAssignmentRhs(insn.Sources, rids)
	// include constant if this makes sense
	if insn.Constant.Cmp(&zero) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&insn.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// construct equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Sum(rhs...))
	// construct constraint
	p.schema.AddVanishingConstraint(name, ctx, util.None[int](), hir.Disjunction(guard, eqn))
	// increment program counter
	p.pcIncrement(pc, pcid, ctx)
	// register constancies
	p.constantExcept(pc, pcid, ctx, insn.Targets, rids, regs)
}

func (p *Compiler) translateJmpInsn(pc uint, pcid uint, ctx trace.Context, rids []uint, regs []Register,
	insn *instruction.Jmp) {
	//
	pc_i := hir.NewColumnAccess(pcid, 0)
	pc_ip1 := hir.NewColumnAccess(pcid, 1)
	name := fmt.Sprintf("pc%d_clk", pc)
	target := hir.NewConst64(uint64(insn.Target))
	// Jump PC
	p.schema.AddVanishingConstraint(name, ctx, util.None[int](),
		hir.Disjunction(hir.NotEquals(pc_i, hir.NewConst64(uint64(pc))), hir.Equals(pc_ip1, target)))
	// register constancies
	p.constantExcept(pc, pcid, ctx, nil, rids, regs)
}

func (p *Compiler) translateJzInsn(pc uint, pcid uint, ctx trace.Context, rids []uint, regs []Register,
	insn *instruction.Jznz) {
	//
	pc_i := hir.NewColumnAccess(pcid, 0)
	pc_ip1 := hir.NewColumnAccess(pcid, 1)
	reg_i := hir.NewColumnAccess(rids[insn.Source], 0)
	target := hir.NewConst64(uint64(insn.Target))
	// taken
	p.schema.AddVanishingConstraint(fmt.Sprintf("pc%d_jz", pc), ctx, util.None[int](),
		hir.Disjunction(hir.NotEquals(pc_i, hir.NewConst64(uint64(pc))),
			hir.NotEquals(reg_i, hir.ZERO),
			hir.Equals(pc_ip1, target)))
	// not taken
	p.schema.AddVanishingConstraint(fmt.Sprintf("pc%d_clk", pc), ctx, util.None[int](),
		hir.Disjunction(hir.NotEquals(pc_i, hir.NewConst64(uint64(pc))),
			hir.Equals(reg_i, hir.ZERO),
			hir.Equals(pc_ip1, hir.Sum(pc_i, hir.ONE))))
	// register constancies
	p.constantExcept(pc, pcid, ctx, nil, rids, regs)
}

func (p *Compiler) translateJnzInsn(pc uint, pcid uint, ctx trace.Context, rids []uint, regs []Register,
	insn *instruction.Jznz) {
	//
	pc_i := hir.NewColumnAccess(pcid, 0)
	pc_ip1 := hir.NewColumnAccess(pcid, 1)
	reg_i := hir.NewColumnAccess(rids[insn.Source], 0)
	target := hir.NewConst64(uint64(insn.Target))
	// taken
	p.schema.AddVanishingConstraint(fmt.Sprintf("pc%d_jnz", pc), ctx, util.None[int](),
		hir.Disjunction(hir.NotEquals(pc_i, hir.NewConst64(uint64(pc))),
			hir.Equals(reg_i, hir.ZERO),
			hir.Equals(pc_ip1, target)))
	// not taken
	p.schema.AddVanishingConstraint(fmt.Sprintf("pc%d_clk", pc), ctx, util.None[int](),
		hir.Disjunction(hir.NotEquals(pc_i, hir.NewConst64(uint64(pc))),
			hir.NotEquals(reg_i, hir.ZERO),
			hir.Equals(pc_ip1, hir.Sum(pc_i, hir.ONE))))
	// register constancies
	p.constantExcept(pc, pcid, ctx, nil, rids, regs)
}

func (p *Compiler) translateMulInsn(pc uint, pcid uint, ctx trace.Context, rids []uint, regs []Register,
	insn *instruction.Mul) {
	//
	var (
		name  = fmt.Sprintf("pc%d_add", pc)
		pc_i  = hir.NewColumnAccess(pcid, 0)
		guard = hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
	)
	// build up the lhs
	lhs := p.buildAssignmentLhs(insn.Targets, rids, regs)
	// build up the rhs
	rhs := p.buildAssignmentRhs(insn.Sources, rids)
	// include constant if this makes sense
	if insn.Constant.Cmp(&one) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&insn.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// construct equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Product(rhs...))
	// construct constraint
	p.schema.AddVanishingConstraint(name, ctx, util.None[int](), hir.Disjunction(guard, eqn))
	// increment program counter
	p.pcIncrement(pc, pcid, ctx)
	// register constancies
	p.constantExcept(pc, pcid, ctx, insn.Targets, rids, regs)
}

func (p *Compiler) translateRetInsn(pc uint, pcid uint, ctx trace.Context) {
	pc_i := hir.NewColumnAccess(pcid, 0)
	pc_ip1 := hir.NewColumnAccess(pcid, 1)
	name := fmt.Sprintf("pc%d_clk", pc)
	// Reset PC
	p.schema.AddVanishingConstraint(name, ctx, util.None[int](),
		hir.Disjunction(hir.NotEquals(pc_i, hir.NewConst64(uint64(pc))), hir.Equals(pc_ip1, hir.ZERO)))
}

func (p *Compiler) translateSubInsn(pc uint, pcid uint, ctx trace.Context, rids []uint, regs []Register,
	insn *instruction.Sub) {
	//
	var (
		name  = fmt.Sprintf("pc%d_sub", pc)
		pc_i  = hir.NewColumnAccess(pcid, 0)
		guard = hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
	)
	// build up the lhs
	lhs := p.buildAssignmentLhs(insn.Targets, rids, regs)
	// build up the rhs
	rhs := p.buildAssignmentRhs(insn.Sources, rids)
	// include constant if this makes sense
	if insn.Constant.Cmp(&zero) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&insn.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// construct equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Subtract(rhs...))
	// construct constraint
	p.schema.AddVanishingConstraint(name, ctx, util.None[int](), hir.Disjunction(guard, eqn))
	// increment program counter
	p.pcIncrement(pc, pcid, ctx)
	// register constancies
	p.constantExcept(pc, pcid, ctx, insn.Targets, rids, regs)
}

// pc = pc + 1
func (p *Compiler) pcIncrement(pc uint, pcid uint, ctx trace.Context) {
	pc_i := hir.NewColumnAccess(pcid, 0)
	pc_ip1 := hir.NewColumnAccess(pcid, 1)
	//
	name := fmt.Sprintf("pc%d_clk", pc)
	// pc != $PC
	guard := hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
	// pc = pc + 1
	inc := hir.Equals(pc_ip1, hir.Sum(hir.ONE, pc_i))
	//
	p.schema.AddVanishingConstraint(name, ctx, util.None[int](), hir.Disjunction(guard, inc))
}

func (p *Compiler) buildAssignmentLhs(targets []uint, rids []uint, regs []Register) []hir.Expr {
	lhs := make([]hir.Expr, len(targets))
	offset := big.NewInt(1)
	// build up the lhs
	for i, dst := range targets {
		// FIXME: shift required!!!
		lhs[i] = hir.NewColumnAccess(rids[dst], 1)
		//
		if i != 0 {
			var elem fr.Element
			//
			elem.SetBigInt(offset)
			lhs[i] = hir.Product(hir.NewConst(elem), lhs[i])
		}
		// left shift offset by given register width.
		offset.Lsh(offset, regs[dst].Width)
	}
	//
	return util.Reverse(lhs)
}

func (p *Compiler) buildAssignmentRhs(sources []uint, rids []uint) []hir.Expr {
	rhs := make([]hir.Expr, len(sources))
	// build up the lhs
	for i, src := range sources {
		rhs[i] = hir.NewColumnAccess(rids[src], 0)
	}
	//
	return rhs
}

// Add constancy constraints for all registers not assigned by a given instruction.
func (p *Compiler) constantExcept(pc uint, pcid uint, ctx trace.Context, targets []uint, rids []uint, regs []Register) {
	var (
		pc_i  = hir.NewColumnAccess(pcid, 0)
		guard = hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
	)
	//
	for i, r := range regs {
		if !slices.Contains(targets, uint(i)) {
			r_i := hir.NewColumnAccess(rids[i], 0)
			r_ip1 := hir.NewColumnAccess(rids[i], 1)
			eqn := hir.Equals(r_i, r_ip1)
			name := fmt.Sprintf("pc%d_%s", pc, r.Name)
			p.schema.AddVanishingConstraint(name, ctx, util.None[int](), hir.Disjunction(guard, eqn))
		}
	}
}
