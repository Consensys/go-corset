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
package program

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Assignment represents a wrapper around an instruction in order for it to
// conform to the schema.Assignment interface.
type Assignment[F field.Element[F], T io.Instruction[T]] struct {
	id        sc.ModuleId
	name      string
	registers []io.Register
	code      []T
	iomap     io.Map
}

// NewAssignment constructs a new assignment capable of trace filling for a
// given function.
func NewAssignment[F field.Element[F], T io.Instruction[T]](id sc.ModuleId, fn io.Function[T], iomap io.Map,
) *Assignment[F, T] {
	//
	return &Assignment[F, T]{
		id:        id,
		name:      fn.Name(),
		registers: fn.Registers(),
		code:      fn.Code(),
		iomap:     iomap,
	}
}

// Bounds implementation for schema.Assignment interface.
func (p Assignment[F, T]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute implementation for schema.Assignment interface.
func (p Assignment[F, T]) Compute(trace tr.Trace[F], schema sc.AnySchema[F]) ([]array.MutArray[F], error) {
	//
	var (
		trModule = trace.Module(p.id)
		states   []io.State
	)
	//
	for i := range trModule.Height() {
		sts := p.trace(i, trModule)
		states = append(states, sts...)
	}
	//
	return p.states2columns(trModule.Width(), states, trace.Builder()), nil
}

// Consistent implementation for schema.Assignment interface.
func (p Assignment[F, T]) Consistent(sc.AnySchema[F]) []error {
	return nil
}

// Lisp implementation for schema.Assignment interface.
func (p Assignment[F, T]) Lisp(schema sc.AnySchema[F]) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("compute"),
		sexp.NewSymbol(p.name),
	})
}

// RegistersExpanded implementation for schema.Assignment interface.
func (p Assignment[F, T]) RegistersExpanded() []sc.RegisterRef {
	return p.RegistersRead()
}

// RegistersRead implementation for schema.Assignment interface.
func (p Assignment[F, T]) RegistersRead() []sc.RegisterRef {
	var regs []sc.RegisterRef
	//
	for i, reg := range p.registers {
		if reg.IsInputOutput() {
			rid := sc.NewRegisterId(uint(i))
			regs = append(regs, sc.NewRegisterRef(p.id, rid))
		}
	}
	//
	return regs
}

// RegistersWritten implementation for schema.Assignment interface.
func (p Assignment[F, T]) RegistersWritten() []sc.RegisterRef {
	var (
		regs       []sc.RegisterRef
		nRegisters = len(p.registers)
		multiLine  = len(p.code) > 1
	)
	// Include control registers for multi-line functions.
	if multiLine {
		nRegisters += 2
	}
	// Trace expansion writes to all registers, including input/outputs.
	// This is because it may expand the I/O registers.
	for i := range nRegisters {
		rid := sc.NewRegisterId(uint(i))
		regs = append(regs, sc.NewRegisterRef(p.id, rid))
	}
	//
	return regs
}

// Trace a given function with the given arguments in a given I/O environment to
// produce a given set of output values, along with the complete set of internal
// traces.
func (p Assignment[F, T]) trace(row uint, trace tr.Module[F]) []io.State {
	var (
		code   = p.code
		states []io.State
		// Construct local state
		state = p.initialState(row, trace, p.iomap)
		// Program counter position
		pc uint = 0
	)
	// Keep executing until we're done.
	for pc != io.RETURN && pc != io.FAIL {
		insn := code[pc]
		// execute given instruction
		pc = insn.Execute(state)
		// record internal state
		states = append(states, finaliseState(row, pc == io.RETURN, state, trace))
		// update state pc
		state.Goto(pc)
	}
	// Done
	return states
}

func (p Assignment[F, T]) initialState(row uint, trace tr.Module[F], iomap io.Map) io.State {
	var (
		state = make([]big.Int, len(p.registers))
		index = 0
	)
	// Initialise arguments
	for i, reg := range p.registers {
		if reg.IsInput() {
			var (
				val = trace.Column(uint(i)).Data().Get(row)
				ith big.Int
			)
			// Clone big int.
			ith.SetBytes(val.Bytes())
			// Assign to ith register
			state[i] = ith
			index = index + 1
		}
	}
	// Construct state
	return io.InitialState(state, p.registers, iomap)
}

// Convert a given set of states into a corresponding set of array columns.
func (p Assignment[F, T]) states2columns(width uint, states []io.State, builder array.Builder[F]) []array.MutArray[F] {
	var (
		cols      = make([]array.MutArray[F], width)
		nrows     = uint(len(states))
		multiLine = len(p.code) > 1
	)
	// Initialise register columns
	for i, r := range p.registers {
		cols[i] = builder.NewArray(nrows, r.Width)
	}
	// Initialise control columns (if applicable)
	// transcribe values
	for row, st := range states {
		for i := range p.registers {
			var (
				val F
				rid = schema.NewRegisterId(uint(i))
			)
			//
			val = val.SetBytes(st.Load(rid).Bytes())
			//
			cols[i].Set(uint(row), val)
		}
	}
	// Set control registers for multi-line functions
	if multiLine {
		p.assignControlRegisters(cols, states, builder)
	}
	// Done
	return cols
}

func (p Assignment[F, T]) assignControlRegisters(cols []array.MutArray[F], states []io.State,
	builder array.Builder[F]) {
	//
	var (
		zero  = field.Zero[F]()
		one   = field.One[F]()
		nrows = uint(len(states))
		pc    = uint(len(p.registers))
		ret   = pc + 1
		// Calculate minimum size of PC; NOTE: +1 because PC==0 is reserved for padding.
		pcWidth = bit.Width(uint(len(p.code) + 1))
	)
	// Initialise columns
	cols[pc] = builder.NewArray(nrows, pcWidth)
	cols[ret] = builder.NewArray(nrows, 1)
	// Assign values
	for row, st := range states {
		npc := field.Uint64[F](uint64(st.Pc() + 1))
		// NOTE: +1 because PC==0 reserved for padding.
		cols[pc].Set(uint(row), npc)
		// Check whether this is a terminating state, or not.
		if st.IsTerminal() {
			cols[ret].Set(uint(row), one)
		} else {
			cols[ret].Set(uint(row), zero)
		}
	}
}

// Finalising a given state does two things: firstly, it clones the state;
// secondly, if the state has terminated, it makes sure the outputs match the
// original trace.
func finaliseState[F field.Element[F]](row uint, terminated bool, state io.State, trace tr.Module[F]) io.State {
	// Clone state
	var nstate = state.Clone()
	// Now, ensure output registers retain their original values.
	if terminated {
		for i, reg := range state.Registers() {
			if reg.IsOutput() {
				var (
					val = trace.Column(uint(i)).Data().Get(row)
					rid = sc.NewRegisterId(uint(i))
					ith big.Int
				)
				// Clone big int.
				ith.SetBytes(val.Bytes())
				//
				nstate.Store(rid, ith)
			}
		}
		// Mark state as terminated
		nstate.Terminate()
	}
	//
	return nstate
}
