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
package io

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
	"github.com/consensys/go-corset/pkg/util/word"
)

// WordPool provides a useful alias
type WordPool = word.Pool[uint, word.BigEndian]

// Assignment represents a wrapper around an instruction in order for it to
// conform to the schema.Assignment interface.
type Assignment[T Instruction[T]] Function[T]

// Bounds implementation for schema.Assignment interface.
func (p Assignment[T]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute implementation for schema.Assignment interface.
func (p Assignment[T]) Compute(trace tr.Trace[word.BigEndian], schema sc.AnySchema,
) ([]array.MutArray[word.BigEndian], error) {
	//
	var (
		trModule = trace.Module(p.id)
		states   []State
	)
	//
	for i := range trModule.Height() {
		sts := p.trace(i, trModule, nil)
		states = append(states, sts...)
	}
	//
	return p.states2columns(trModule.Width(), states, trace.Pool()), nil
}

// Consistent implementation for schema.Assignment interface.
func (p Assignment[T]) Consistent(sc.AnySchema) []error {
	return nil
}

// Lisp implementation for schema.Assignment interface.
func (p Assignment[T]) Lisp(schema sc.AnySchema) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("compute"),
		sexp.NewSymbol(p.name),
	})
}

// RegistersExpanded implementation for schema.Assignment interface.
func (p Assignment[T]) RegistersExpanded() []sc.RegisterRef {
	return p.RegistersRead()
}

// RegistersRead implementation for schema.Assignment interface.
func (p Assignment[T]) RegistersRead() []sc.RegisterRef {
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
func (p Assignment[T]) RegistersWritten() []sc.RegisterRef {
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
func (p Assignment[T]) trace(row uint, trace tr.Module[word.BigEndian], iomap Map) []State {
	var (
		code   = p.code
		states []State
		// Construct local state
		state = p.initialState(row, trace, iomap)
		// Program counter position
		pc uint = 0
	)
	// Keep executing until we're done.
	for pc != RETURN {
		insn := code[pc]
		// execute given instruction
		pc = insn.Execute(state)
		// record internal state
		states = append(states, finaliseState(row, pc == RETURN, state, trace))
		// update state pc
		state.Goto(pc)
	}
	// Done
	return states
}

func (p Assignment[T]) initialState(row uint, trace tr.Module[word.BigEndian], io Map) State {
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
	return State{0, false, state, p.registers, io}
}

// Convert a given set of states into a corresponding set of array columns.
func (p Assignment[T]) states2columns(width uint, states []State, pool WordPool) []array.MutArray[word.BigEndian] {
	var (
		cols      = make([]array.MutArray[word.BigEndian], width)
		nrows     = uint(len(states))
		multiLine = len(p.code) > 1
	)
	// Initialise register columns
	for i, r := range p.registers {
		cols[i] = word.NewArray(nrows, r.Width, pool)
	}
	// Initialise control columns (if applicable)
	// transcribe values
	for row, st := range states {
		for i := range p.registers {
			var (
				val word.BigEndian
				rid = schema.NewRegisterId(uint(i))
			)
			//
			val.SetBytes(st.Load(rid).Bytes())
			//
			cols[i].Set(uint(row), val)
		}
	}
	// Set control registers for multi-line functions
	if multiLine {
		p.assignControlRegisters(cols, states, pool)
	}
	// Done
	return cols
}

func (p Assignment[T]) assignControlRegisters(cols []array.MutArray[word.BigEndian], states []State, pool WordPool) {
	var (
		zero  = word.Uint64[word.BigEndian](0)
		one   = word.Uint64[word.BigEndian](1)
		nrows = uint(len(states))
		pc    = uint(len(p.registers))
		ret   = pc + 1
		// Calculate minimum size of PC; NOTE: +1 because PC==0 is reserved for padding.
		pcWidth = bit.Width(uint(len(p.code) + 1))
	)
	// Initialise columns
	cols[pc] = word.NewArray(nrows, pcWidth, pool)
	cols[ret] = word.NewArray(nrows, 1, pool)
	// Assign values
	for row, st := range states {
		npc := word.Uint64[word.BigEndian](uint64(st.Pc() + 1))
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
func finaliseState(row uint, terminated bool, state State, trace tr.Module[word.BigEndian]) State {
	// Clone state
	var nstate = state.Clone()
	// Now, ensure output registers retain their original values.
	if terminated {
		for i, reg := range state.registers {
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
