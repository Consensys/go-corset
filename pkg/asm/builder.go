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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
)

// sets the maximum width of the program counter.
const pc_width = uint(8)

// TraceBuilder provides a mechanical means of constructing a trace from a given
// schema and set of input columns.  The goal is to encapsulate all of the logic
// around building a trace.
type TraceBuilder[T insn.Instruction] struct {
	functions []Function[T]
}

// NewTraceBuilder constructs a new trace builder for a given set of functions.
func NewTraceBuilder[T insn.Instruction](functions ...Function[T]) *TraceBuilder[T] {
	return &TraceBuilder[T]{functions}
}

// Build constructs a complete trace, given a set of function instances.
func (p *TraceBuilder[T]) Build(instances []FunctionInstance) []trace.RawColumn {
	var columns []trace.RawColumn
	//
	for i := range p.functions {
		fncols := p.expandFunctionInstances(uint(i), instances)
		columns = append(columns, fncols...)
	}
	//
	return columns
}

func (p *TraceBuilder[T]) expandFunctionInstances(fid uint, instances []FunctionInstance) []trace.RawColumn {
	var (
		fn      = p.functions[fid]
		data    = make([][]big.Int, len(fn.Registers)+2)
		columns = make([]trace.RawColumn, len(fn.Registers)+2)
		stamp   = uint(1)
	)
	//
	for _, inst := range instances {
		if inst.Function == fid {
			data = p.traceFunction(fid, stamp, data, inst)
			stamp = stamp + 1
		}
	}
	// Construct stamp column
	columns[0] = trace.RawColumn{
		Module: fn.Name,
		Name:   "$stamp",
		Data:   field.FrArrayFromBigInts(32, data[0]),
	}
	// Construct PC column
	columns[1] = trace.RawColumn{
		Module: fn.Name,
		Name:   "$pc",
		Data:   field.FrArrayFromBigInts(pc_width, data[1]),
	}
	// Construct register columns.
	for i, r := range fn.Registers {
		data := field.FrArrayFromBigInts(r.Width, data[i+2])
		columns[i+2] = trace.RawColumn{
			Module: fn.Name,
			Name:   r.Name,
			Data:   data,
		}
	}
	//
	return columns
}

func (p *TraceBuilder[T]) traceFunction(fid uint, stamp uint, trace [][]big.Int,
	instance FunctionInstance) [][]big.Int {
	//
	interpreter := NewInterpreter(p.functions...)
	// Initialise state
	init := interpreter.Bind(fid, instance.Inputs)
	biStamp := big.NewInt(int64(stamp))
	//
	interpreter.Enter(fid, init)
	//
	for !interpreter.HasTerminated() {
		// Record state
		state := interpreter.State()
		pc := big.NewInt(int64(state.pc))
		//
		trace[0] = append(trace[0], *biStamp)
		trace[1] = append(trace[1], *pc)
		//
		for i, v := range state.registers {
			// Clone value to prevent side-effects
			var ith big.Int
			//
			ith.Set(&v)
			// Update column for this register.
			trace[i+2] = append(trace[i+2], ith)
		}
		// Execute
		interpreter.Execute(1)
	}
	//
	return trace
}
