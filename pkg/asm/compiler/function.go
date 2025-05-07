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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/asm/micro"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

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

// Schema returns the generated schema
func (p *Compiler) Schema() *hir.Schema {
	return &p.schema
}

func (p *Compiler) Compile(name string, regs []insn.Register, code []micro.Instruction) {
	var (
		// Allocate module id
		mid = p.schema.AddModule(name)
		// Map fn registers to schema columns
		rids = make([]uint, len(regs))
	)
	// Configure enclosing context
	ctx := trace.NewContext(mid, 1)
	// Allocate registers as columns
	for i, reg := range regs {
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
	stampID, pcID := p.initFunctionFraming(ctx, rids, regs, code)
	// Construct appropriate mapping
	mapping := Translator{
		Schema:    &p.schema,
		StampID:   stampID,
		PcID:      pcID,
		Context:   ctx,
		RegIDs:    rids,
		Registers: regs,
	}
	// Compile each instruction in turn
	for pc, inst := range code {
		// Core translation
		mapping.Translate(uint(pc), inst)
	}
}

func (p *Compiler) initFunctionFraming(ctx trace.Context, rids []uint, regs []insn.Register,
	code []micro.Instruction) (uint, uint) {
	//
	pcMax := uint64(len(code) - 1)
	// Determine max width of PC
	pcWidth := uint(big.NewInt(int64(pcMax)).BitLen())
	// Allocate book keeping columns
	stamp := p.schema.AddDataColumn(ctx, "$stamp", schema.NewUintType(p.maxInstances))
	pc := p.schema.AddDataColumn(ctx, "$pc", schema.NewUintType(pcWidth))
	//
	stamp_i := hir.NewColumnAccess(stamp, 0)
	stamp_im1 := hir.NewColumnAccess(stamp, -1)
	pc_i := hir.NewColumnAccess(pc, 0)
	// $stamp == 0 on first row
	p.schema.AddVanishingConstraint("first", ctx, util.Some(0), hir.Equals(stamp_i, hir.ZERO))
	// $stamp == 0 || $pc == ...
	p.schema.AddVanishingConstraint("last", ctx, util.Some(-1),
		hir.If(hir.NotEquals(stamp_i, hir.ZERO), terminators(pc_i, code)))
	//
	// prev($stamp) == $stamp || prev($stamp)+1== $stamp
	p.schema.AddVanishingConstraint("increment", ctx, util.None[int](),
		hir.If(hir.NotEquals(stamp_im1, stamp_i), hir.Equals(hir.Sum(hir.ONE, stamp_im1), stamp_i)))
	// prev($stamp) == $stamp || $pc == 0
	p.schema.AddVanishingConstraint("reset", ctx, util.None[int](),
		hir.If(hir.NotEquals(stamp_im1, stamp_i), hir.Equals(pc_i, hir.ZERO)))
	// Add constancies for all input registers
	for i, r := range regs {
		rid := rids[i]
		//
		if r.IsInput() {
			name := fmt.Sprintf("const_%s", r.Name)
			reg_i := hir.NewColumnAccess(rid, 0)
			reg_im1 := hir.NewColumnAccess(rid, -1)
			//
			p.schema.AddVanishingConstraint(name, ctx, util.None[int](),
				hir.If(hir.NotEquals(pc_i, hir.ZERO), hir.Equals(reg_im1, reg_i)))
		}
	}
	//
	return stamp, pc
}

func terminators(pc_i hir.Expr, code []micro.Instruction) hir.Expr {
	var (
		terminator hir.Expr
		first      = true
	)
	//
	for pc, insn := range code {
		if insn.Terminal() {
			ith := hir.Equals(pc_i, hir.NewConst64(uint64(pc)))
			if first {
				terminator = ith
				first = false
			} else {
				terminator = hir.Disjunction(terminator, ith)
			}
		}
	}
	//
	return terminator
}
