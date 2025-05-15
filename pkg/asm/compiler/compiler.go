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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// MicroFunction is a function composed entirely of micro instructions.
type MicroFunction = io.Function[micro.Instruction]

type busInfo struct {
	// Name of the bus
	name string
	// Registers
	registers []io.Register
	// Underlying HIR column ids
	columns []uint
}

// Width returns the width of this bus.  That is, the number of input/output
// registers.
func (p *busInfo) Width() uint {
	return uint(len(p.registers))
}

// Compiler packages up everything needed to compile a given assembly down into
// an HIR schema.  Observe that the compiler may fail if the assembly files are
// malformed in some way (e.g. fail type checking).
type Compiler struct {
	schema hir.Schema
	// maxInstances determines the maximum number of instances permitted for any
	// given function.
	maxInstances uint
	// Bus records
	buses []busInfo
	// Mapping  of bus names to bus records.
	busMap map[string]trace.Context
	// types & reftables
	// sourcemap
}

// NewCompiler constructs a new compiler
func NewCompiler() *Compiler {
	schema := hir.EmptySchema()
	//
	return &Compiler{
		schema:       *schema,
		maxInstances: 32,
		buses:        nil,
		busMap:       make(map[string]trace.Context),
	}
}

// Schema returns the generated schema
func (p *Compiler) Schema() *hir.Schema {
	return &p.schema
}

// RegisterBus registers a bus with a given name and width (i.e. number of
// address / value lines).
func (p *Compiler) RegisterBus(name string, inputs []io.Register, outputs []io.Register) {
	// Allocate module id
	id := p.schema.AddModule(name, hir.VOID)
	// sanity check bus id matches module id
	if id != uint(len(p.buses)) {
		panic("invalid module <=> bus mapping")
	}
	//
	ctx := trace.NewContext(id, 1)
	// Add I/O lines only
	inputColumns := p.allocateIoLines(ctx, inputs, true)
	outputColumns := p.allocateIoLines(ctx, outputs, false)
	//
	p.buses = append(p.buses, busInfo{
		name,
		append(inputs, outputs...),
		append(inputColumns, outputColumns...),
	})
	// Allocate bus context
	p.busMap[name] = ctx
}

// Compile a function with the given name, registers and micro-instructions into
// constraints.
func (p *Compiler) Compile(fn MicroFunction) {
	ctx := p.busMap[fn.Name()]
	// Determine correct register ids
	rids := p.initFunctionRegisters(ctx, fn.Registers())
	// Setup framing columns / constraints
	stampID, pcID := p.initFunctionFraming(ctx, rids, fn)
	// Initialise buses required for this code sequence
	p.initBuses(ctx, fn, rids)
	// Construct appropriate mapping
	mapping := Translator{
		Schema:    &p.schema,
		StampID:   stampID,
		PcID:      pcID,
		Context:   ctx,
		RegIDs:    rids,
		Registers: fn.Registers(),
	}
	// Compile each instruction in turn
	for pc, inst := range fn.Code() {
		// Core translation
		mapping.Translate(uint(pc), inst)
	}
}

// Initialise the mapping from registers to HIR column identifiers.  Observe
// that input / output registers will have already been allocated during bus
// initialisation.  Therefore, we have to extract their identifiers rather than
// allocate new columns.
func (p *Compiler) initFunctionRegisters(ctx trace.Context, regs []io.Register) []uint {
	var (
		bus     = p.buses[ctx.ModuleId]
		columns = make([]uint, len(regs))
		ioreg   uint
	)
	//
	for i, reg := range regs {
		// Sanity checks
		if reg.IsInput() || reg.IsOutput() {
			ioName := bus.registers[ioreg].Name
			// sanity check
			if reg.Name != ioName {
				panic(fmt.Sprintf("mis-aligned I/O register %s <=> %s", reg.Name, ioName))
			}
			// input / output register, so lookup existing line.
			columns[i] = bus.columns[ioreg]
			ioreg++
		} else {
			// internal register, so allocate new line.
			columns[i] = p.allocateRegisterLine(ctx, reg)
		}
	}
	// Done
	return columns
}

func (p *Compiler) initFunctionFraming(ctx trace.Context, rids []uint, fn MicroFunction) (uint, uint) {
	//
	pcMax := uint64(len(fn.Code()) - 1)
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
		hir.If(hir.NotEquals(stamp_i, hir.ZERO), terminators(pc_i, fn)))
	//
	// prev($stamp) == $stamp || prev($stamp)+1== $stamp
	p.schema.AddVanishingConstraint("increment", ctx, util.None[int](),
		hir.If(hir.NotEquals(stamp_im1, stamp_i), hir.Equals(hir.Sum(hir.ONE, stamp_im1), stamp_i)))
	// prev($stamp) == $stamp || $pc == 0
	p.schema.AddVanishingConstraint("reset", ctx, util.None[int](),
		hir.If(hir.NotEquals(stamp_im1, stamp_i), hir.Equals(pc_i, hir.ZERO)))
	// Add constancies for all input registers
	for i, r := range fn.Registers() {
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

func (p *Compiler) allocateIoLines(ctx trace.Context, lines []io.Register, inputs bool) []uint {
	var columns []uint
	//
	for _, reg := range lines {
		// Sanity checks
		if inputs && !reg.IsInput() {
			panic(fmt.Sprintf("invalid input register %s", reg.Name))
		} else if !inputs && !reg.IsOutput() {
			panic(fmt.Sprintf("invalid output register %s", reg.Name))
		}
		// Allocate register
		columns = append(columns, p.allocateRegisterLine(ctx, reg))
	}

	return columns
}

// Allocate a given register into the underlying schema, producing an HIR column
// identifier.
func (p *Compiler) allocateRegisterLine(ctx trace.Context, reg io.Register) uint {
	typeName := fmt.Sprintf("%s:u%d", reg.Name, reg.Width)
	// Construct appropriate datatype
	datatype := schema.NewUintType(reg.Width)
	// Allocate register
	cid := p.schema.AddDataColumn(ctx, reg.Name, datatype)
	// Add range constraint
	p.schema.AddRangeConstraint(typeName, ctx,
		hir.NewColumnAccess(cid, 0), datatype.Bound())
	// Done
	return cid
}

// Initialise the buses linked in a given function.
func (p *Compiler) initBuses(caller trace.Context, fn MicroFunction, columns []uint) {
	//
	for _, bus := range localBuses(fn) {
		// Callee represents the function being called by this bus.
		var (
			callee        = trace.NewContext(bus.BusId, 1)
			name          = fmt.Sprintf("%s=>%s", fn.Name(), bus.Name)
			callerRegs    = append(bus.Address(), bus.Data()...)
			callerLines   = make([]hir.Expr, len(callerRegs))
			calleeColumns = p.buses[bus.BusId].columns
			calleeLines   = make([]hir.Expr, len(calleeColumns))
		)
		// Initialise caller lines
		for i, r := range callerRegs {
			callerLines[i] = hir.NewColumnAccess(columns[r], 0)
		}
		// Initialise callee lines
		for i, c := range calleeColumns {
			calleeLines[i] = hir.NewColumnAccess(c, 0)
		}
		// Add lookup constraint
		p.schema.AddLookupConstraint(name, caller, callee, callerLines, calleeLines)
	}
}

// Determine the set of buses used within a function, by inspecting each
// instruction in turn.  Observe the resulting array does not contain duplicate
// entries.
func localBuses(fn MicroFunction) []io.Bus {
	var (
		insns = fn.Code()
		// Set of buses already seen
		seen bit.Set
		// Collected buses
		buses []io.Bus
	)
	//
	for _, insn := range insns {
		for _, ucode := range insn.Codes {
			if bi, ok := ucode.(io.InOutInstruction); ok {
				bus := bi.Bus()
				//
				if !seen.Contains(bus.BusId) {
					buses = append(buses, bus)
					seen.Insert(bus.BusId)
				}
			}
		}
	}
	//
	return buses
}

func terminators(pc_i hir.Expr, fn MicroFunction) hir.Expr {
	var (
		terminator hir.Expr
		first      = true
	)
	//
	for pc, insn := range fn.Code() {
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
