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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// STAMP_NAME determines the name used in the underlying constraint system for
// stamp column.
const STAMP_NAME = "$stamp"

// PC_NAME determines the name used in the underlying constraint system for the
// Program Counter (PC) column.
const PC_NAME = "$pc"

// MicroFunction is a function composed entirely of micro instructions.
type MicroFunction = io.Function[micro.Instruction]

// FunctionMapping provides information regarding the mapping of a
// assembly-level component (e.g. a function) to the corresponding columns in
// the underlying constraint system.
type FunctionMapping[T any] struct {
	// Name of the Bus
	name string
	// Registers
	registers []io.Register
	// Underlying column ids for registers
	columns []T
}

// ColumnsOf returns the underlying column identifiers for a given set of zero
// or more registers.
func (p *FunctionMapping[T]) ColumnsOf(registers ...io.RegisterId) []T {
	columns := make([]T, len(registers))
	//
	for i, r := range registers {
		columns[i] = p.columns[r.Unwrap()]
	}
	//
	return columns
}

// Bus returns the set of input/output columns which represent the "Bus" for
// this component.
func (p *FunctionMapping[T]) Bus() []T {
	var columns []T
	//
	for i, r := range p.registers {
		if r.IsInput() || r.IsOutput() {
			columns = append(columns, p.columns[i])
		}
	}
	//
	return columns
}

// Compiler packages up everything needed to compile a given assembly down into
// an HIR schema.  Observe that the compiler may fail if the assembly files are
// malformed in some way (e.g. fail type checking).
type Compiler[T any, E Expr[T, E], M Module[T, E, M]] struct {
	modules []M
	// maxInstances determines the maximum number of instances permitted for any
	// given function.
	maxInstances uint
	// Bus records
	buses []FunctionMapping[T]
	// Mapping  of Bus names to Bus records.
	busMap map[string]uint
	// types & reftables
	// sourcemap
}

// NewCompiler constructs a new compiler
func NewCompiler[T any, E Expr[T, E], M Module[T, E, M]]() *Compiler[T, E, M] {
	return &Compiler[T, E, M]{
		modules:      nil,
		maxInstances: 32,
		buses:        nil,
		busMap:       make(map[string]uint),
	}
}

// Modules returns the abstract modules constructed during compilation.
func (p *Compiler[T, E, M]) Modules() []M {
	return p.modules
}

// Compile a given set of micro functions
func (p *Compiler[T, E, M]) Compile(fns ...*MicroFunction) {
	p.modules = make([]M, len(fns))
	p.buses = make([]FunctionMapping[T], len(fns))
	// Initialise buses
	for i, f := range fns {
		p.initModule(uint(i), *f)
	}
	// Compiler functions
	for _, fn := range fns {
		p.compileFunction(*fn)
	}
}

// Compile a function with the given name, registers and micro-instructions into
// constraints.
func (p *Compiler[T, E, M]) compileFunction(fn MicroFunction) {
	busId := p.busMap[fn.Name()]
	// Setup framing columns / constraints
	framing := p.initFunctionFraming(busId, fn)
	// Initialise buses required for this code sequence
	p.initBuses(busId, fn)
	// Construct appropriate mapping
	mapping := Translator[T, E, M]{
		Module:    p.modules[busId],
		Framing:   framing,
		Registers: fn.Registers(),
		Columns:   p.buses[busId].columns,
	}
	// Compile each instruction in turn
	for pc, inst := range fn.Code() {
		// Core translation
		mapping.Translate(uint(pc), inst)
	}
}

// Create columns in the respective module for all registers associated with a
// given Bus component (e.g. function).
func (p *Compiler[T, E, M]) initModule(busId uint, fn MicroFunction) {
	var (
		module M
		bus    FunctionMapping[T]
	)
	// Initialise module correctly
	module = module.Initialise(fn, busId)
	p.modules[busId] = module
	//
	bus.name = fn.Name()
	bus.registers = fn.Registers()
	bus.columns = make([]T, len(fn.Registers()))
	//
	for i, reg := range fn.Registers() {
		// if i == 0 {
		// 	// Program Counter is always at index 0.  For a one-line function,
		// 	// the program counter is not used.
		// 	bus.columns[i] = module.NewUnusedColumn()
		// } else {
		// 	bus.columns[i] = module.NewColumn(reg.Kind, reg.Name, reg.Width)
		// }
		//
		bus.columns[i] = module.NewColumn(reg.Kind, reg.Name, reg.Width)
	}
	//
	p.buses[busId] = bus
	p.busMap[bus.name] = busId
}

func (p *Compiler[T, E, M]) initFunctionFraming(busId uint, fn MicroFunction) Framing[T, E] {
	// One line (i.e. atomic functions doen't require any framing.  They don't
	// even require a program counter!!
	if fn.IsAtomic() {
		return NewAtomicFraming[T, E]()
	}
	// Multi-line functions require proper framing.
	return p.initMultLineFunctionFraming(busId, fn)
}

func (p *Compiler[T, E, M]) initMultLineFunctionFraming(busId uint, fn MicroFunction) Framing[T, E] {
	var (
		Bus    = p.buses[busId]
		module = p.modules[busId]
		// allocate return line
		ret = module.NewColumn(schema.COMPUTED_REGISTER, STAMP_NAME, 1)
		// PC is always first register, therefore no need to create a new column for
		// it.
		pc = Bus.columns[0]
	)
	// FIXME: the following reliance on the length of registers is something of
	// a kludge.
	stampRef := sc.NewRegisterRef(busId, sc.NewRegisterId(uint(len(fn.Registers()))))
	module.NewAssignment(&StampAssignment{stampRef})
	//
	// stamp_i := Variable[T, E](stamp, 0)
	// stamp_im1 := Variable[T, E](stamp, -1)
	// pc_i := Variable[T, E](pc, 0)
	// zero := Number[T, E](0)
	// one := Number[T, E](1)
	// // stamp[0] == 0
	// module.NewConstraint("first", util.Some(0), stamp_i.Equals(zero))
	// // stamp[i] == 0 || pc[i] == ... [BROKEN]
	// // module.NewConstraint("last", util.Some(-1),
	// //	If(stamp_i.NotEquals(zero), terminators(pc_i, fn)))
	// // stamp[i-1] != stamp[i] ==> stamp[i-1]+1 == stamp[i]
	// module.NewConstraint("increment", util.None[int](),
	// 	If(stamp_im1.NotEquals(stamp_i), one.Add(stamp_im1).Equals(stamp_i)))
	// // stamp[i-1] != stamp[i] ==> pc[i] == 0
	// module.NewConstraint("reset", util.None[int](),
	// 	If(stamp_im1.NotEquals(stamp_i), pc_i.Equals(zero)))

	// Add constancies for all input registers (if applicable)
	p.addInputConstancies(busId, fn)
	//
	return NewMultiLineFraming[T, E](pc, ret)
}

// Add input constancies for the given function.  That is, constraints which
// ensure the inputs don't change within a given frame.  Observe that this only
// applies for multi-line functions, as one-line functions don't have internal
// states.
func (p *Compiler[T, E, M]) addInputConstancies(busId uint, fn MicroFunction) {
	var (
		Bus    = p.buses[busId]
		module = p.modules[busId]
		pc     = Bus.columns[0]
		pc_i   = Variable[T, E](pc, 0)
		zero   = Number[T, E](0)
	)
	// Constancies only required when there is more than one instruction
	// since otherwise pc==0 always.
	for i, r := range fn.Registers() {
		if r.IsInput() {
			name := fmt.Sprintf("const_%s", r.Name)
			reg_i := Variable[T, E](Bus.columns[i], 0)
			reg_im1 := Variable[T, E](Bus.columns[i], -1)
			//
			module.NewConstraint(name, util.None[int](),
				If(pc_i.NotEquals(zero), reg_im1.Equals(reg_i)))
		}
	}
}

// Initialise the buses linked in a given function.
func (p *Compiler[T, E, M]) initBuses(caller uint, fn MicroFunction) {
	var module = p.modules[caller]
	//
	for _, bus := range localBuses(fn) {
		// Callee represents the function being called by this Bus.
		var (
			name        = fmt.Sprintf("%s=>%s", fn.Name(), bus.Name)
			callerBus   = p.buses[caller].ColumnsOf(bus.AddressData()...)
			callerLines = make([]E, len(callerBus))
			calleeBus   = p.buses[bus.BusId].Bus()
			calleeLines = make([]E, len(calleeBus))
		)
		// Initialise caller lines
		for i, r := range callerBus {
			callerLines[i] = Variable[T, E](r, 0)
		}
		// Initialise callee lines
		for i, r := range calleeBus {
			calleeLines[i] = Variable[T, E](r, 0)
		}
		// Add lookup constraint
		module.NewLookup(name, callerLines, bus.BusId, calleeLines)
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
				Bus := bi.Bus()
				//
				if !seen.Contains(Bus.BusId) {
					buses = append(buses, Bus)
					seen.Insert(Bus.BusId)
				}
			}
		}
	}
	//
	return buses
}
