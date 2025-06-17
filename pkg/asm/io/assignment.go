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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Assignment represents a wrapper around an instruction in order for it to
// conform to the schema.Assignment interface.
type Assignment[T Instruction[T]] Function[T]

// Bounds implementation for schema.Assignment interface.
func (p Assignment[T]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute implementation for schema.Assignment interface.
func (p Assignment[T]) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	var (
		trModule = trace.Module(p.id)
		states   []State
	)
	//
	for i := range trModule.Height() {
		_, sts := p.trace(i, trModule, nil)
		states = append(states, sts...)
	}
	//
	return states2columns(trModule.Width(), p.registers, states), nil
}

// Consistent implementation for schema.Assignment interface.
func (p Assignment[T]) Consistent(sc.AnySchema) []error {
	return nil
}

// Lisp implementation for schema.Assignment interface.
func (p Assignment[T]) Lisp(schema sc.AnySchema) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("function"),
		sexp.NewSymbol(p.name),
	})
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
	var regs []sc.RegisterRef
	//
	for i := range p.registers {
		// Trace expansion writes to all registers, including input/outputs.
		// This is because it may expand the I/O registers.
		rid := sc.NewRegisterId(uint(i))
		regs = append(regs, sc.NewRegisterRef(p.id, rid))
	}
	//
	return regs
}

// Trace a given function with the given arguments in a given I/O environment to
// produce a given set of output values, along with the complete set of internal
// traces.
func (p Assignment[T]) trace(row uint, trace tr.Module, iomap Map) ([]big.Int, []State) {
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
		states = append(states, state.Clone())
		// update state pc
		state.Goto(pc)
	}
	// Done
	return state.Outputs(), states
}

func (p Assignment[T]) initialState(row uint, trace tr.Module, io Map) State {
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
			val.BigInt(&ith)
			// NOTE: following safe because PC is always at index 0, and is a
			// computed register.
			state[i-1] = ith
			index = index + 1
		}
	}
	// Construct state
	return State{0, state, p.registers, io}
}

// Convert a given set of states into a corresponding set of array columns.
func states2columns(width uint, registers []Register, states []State) []tr.ArrayColumn {
	var (
		cols  = make([]tr.ArrayColumn, width)
		zero  = fr.NewElement(0)
		nrows = uint(len(states))
	)
	// Initialise register columns
	for i, r := range registers {
		arr := field.NewFrArray(nrows, r.Width)
		cols[i] = tr.NewArrayColumn(r.Name, arr, zero)
	}
	// transcribe values
	for row, st := range states {
		for i := range registers {
			var (
				val fr.Element
				rid = schema.NewRegisterId(uint(i))
			)
			//
			val.SetBigInt(st.Load(rid))
			//
			cols[i].Data().Set(uint(row), val)
		}
	}
	//
	return cols
}
