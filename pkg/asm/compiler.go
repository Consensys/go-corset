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

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
)

// CompileAssembly compiles a given set of assembly functions into a binary
// constraint file.
func CompileAssembly(assembly ...source.File) (*binfile.BinaryFile, []source.SyntaxError) {
	functions, _, errs := Assemble(assembly...)
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
		// Map fn registers to schema columns
		rids = make([]uint, len(fn.Registers))
	)
	// Configure enclosing context
	ctx := trace.NewContext(mid, 1)
	// Allocate registers as columns
	for i, reg := range fn.Registers {
		typeName := fmt.Sprintf("%s:u%d", reg.Name, reg.Width)
		// Construct appropriate datatype
		datatype := schema.NewUintType(reg.Width)
		// Allocate register
		rids[i] = p.schema.AddDataColumn(ctx, reg.Name, datatype)
		// Add range constraint
		p.schema.AddRangeConstraint(typeName, ctx,
			hir.NewColumnAccess(rids[i], 0), datatype.Bound())
	}
	// Setup framing columns / constraints
	stampID, pcID := p.initFunctionFraming(ctx, rids, fn)
	// Construct appropriate mapping
	mapping := insn.StateMapping{
		Schema:    &p.schema,
		StampID:   stampID,
		PcID:      pcID,
		Context:   ctx,
		RegIDs:    rids,
		Registers: fn.Registers,
	}
	// Compile each instruction in turn
	for pc, inst := range fn.Code {
		// Initialise state translator
		state := insn.NewStateTranslator(mapping, uint(pc))
		// Core translation
		p.compileInstruction(inst, state)
	}
}

func (p *Compiler) compileInstruction(inst Instruction, st insn.StateTranslator) {
	for _, microinsn := range inst.Instructions {
		microinsn.Translate(&st)
	}
	// Finalise state translation
	st.Finalise()
}

func (p *Compiler) initFunctionFraming(ctx trace.Context, rids []uint, fn Function) (uint, uint) {
	pcMax := uint64(len(fn.Code) - 1)
	// Determine max width of PC
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
	for _, tpc := range terminators(fn) {
		p.schema.AddVanishingConstraint("last", ctx, util.Some(-1),
			hir.Disjunction(hir.Equals(stamp_i, hir.ZERO), hir.Equals(pc_i, hir.NewConst64(tpc))))
	}
	// next($stamp) == $stamp || next($stamp) == $stamp+1
	p.schema.AddVanishingConstraint("increment", ctx, util.None[int](),
		hir.Disjunction(hir.Equals(stamp_ip1, stamp_i), hir.Equals(stamp_ip1, hir.Sum(hir.ONE, stamp_i))))
	// next($stamp) == $stamp || next($pc) == 0
	p.schema.AddVanishingConstraint("reset", ctx, util.None[int](),
		hir.Disjunction(hir.Equals(stamp_ip1, stamp_i), hir.Equals(pc_ip1, hir.ZERO)))
	// Add constancies for all input registers
	for i, r := range fn.Registers {
		rid := rids[i]
		//
		if r.IsInput() {
			name := fmt.Sprintf("const_%s", r.Name)
			reg_i := hir.NewColumnAccess(rid, 0)
			reg_ip1 := hir.NewColumnAccess(rid, 1)
			//
			p.schema.AddVanishingConstraint(name, ctx, util.None[int](),
				hir.Disjunction(hir.Equals(pc_ip1, hir.ZERO), hir.Equals(reg_ip1, reg_i)))
		}
	}
	//
	return stamp, pc
}

func terminators(fn Function) []uint64 {
	var terminals []uint64
	//
	for pc, insn := range fn.Code {
		if insn.Terminal() {
			terminals = append(terminals, uint64(pc))
		}
	}
	//
	return terminals
}
