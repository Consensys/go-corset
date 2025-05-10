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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
)

// MacroTrace represents a program trace at the macro level.
type MacroTrace = io.Trace[macro.Instruction]

// MicroTrace represents a program trace at the micro level.
type MicroTrace = io.Trace[micro.Instruction]

// LowerTraces lowers macro-level traces to micro-level traces according to a given lowering config.
func LowerTraces(config LoweringConfig, traces ...io.Trace[macro.Instruction]) []io.Trace[micro.Instruction] {
	utraces := make([]MicroTrace, len(traces))
	//
	for i, tr := range traces {
		utrace := LowerMacroTrace(config, tr)
		utraces[i] = utrace
	}
	//
	return utraces
}

// LowerMicroTrace this micro trace to a set of raw columns.
func LowerMicroTrace(p MicroTrace) []tr.RawColumn {
	builder := NewTraceBuilder(p.Program())
	return builder.Build(p)
}

// LowerMacroTrace this macro trace into a micro trace according to a given
// lowering config.
func LowerMacroTrace(cfg LoweringConfig, trace MacroTrace) MicroTrace {
	var (
		macroProgram   = trace.Program()
		microProgram   = Lower(cfg, trace.Program())
		microInstances = make([]io.FunctionInstance, len(trace.Instances()))
	)
	//
	for i, inst := range trace.Instances() {
		microInstances[i] = io.SplitInstance(cfg.MaxRegisterWidth, inst, macroProgram)
	}
	// Done
	return io.NewTrace(microProgram, microInstances...)
}

// ============================================================================
// Helper
// ============================================================================

// sets the maximum width of the program counter.
const pc_width = uint(8)

// TraceBuilder provides a mechanical means of constructing a trace from a given
// schema and set of input columns.  The goal is to encapsulate all of the logic
// around building a trace.
type TraceBuilder[T io.Instruction[T]] struct {
	program io.Program[T]
}

// NewTraceBuilder constructs a new trace builder for a given set of functions.
func NewTraceBuilder[T io.Instruction[T]](program io.Program[T]) *TraceBuilder[T] {
	return &TraceBuilder[T]{program}
}

// Build constructs a complete trace, given a set of function instances.
func (p *TraceBuilder[T]) Build(trace io.Trace[T]) []tr.RawColumn {
	var columns []tr.RawColumn
	//
	for i := range p.program.Functions() {
		fncols := p.expandFunctionInstances(uint(i), trace)
		columns = append(columns, fncols...)
	}
	//
	return columns
}

func (p *TraceBuilder[T]) expandFunctionInstances(fid uint, trace io.Trace[T]) []tr.RawColumn {
	var (
		fn      = p.program.Function(fid)
		data    = make([][]big.Int, len(fn.Registers)+2)
		columns = make([]tr.RawColumn, len(fn.Registers)+2)
		stamp   = uint(1)
	)
	//
	for _, inst := range trace.Instances() {
		if inst.Function == fid {
			data = p.traceFunction(fid, stamp, data, inst)
			stamp = stamp + 1
		}
	}
	// Construct stamp column
	columns[0] = tr.RawColumn{
		Module: fn.Name,
		Name:   "$stamp",
		Data:   field.FrArrayFromBigInts(32, data[0]),
	}
	// Construct PC column
	columns[1] = tr.RawColumn{
		Module: fn.Name,
		Name:   "$pc",
		Data:   field.FrArrayFromBigInts(pc_width, data[1]),
	}
	// Construct register columns.
	for i, r := range fn.Registers {
		data := field.FrArrayFromBigInts(r.Width, data[i+2])
		columns[i+2] = tr.RawColumn{
			Module: fn.Name,
			Name:   r.Name,
			Data:   data,
		}
	}
	//
	return columns
}

func (p *TraceBuilder[T]) traceFunction(fid uint, stamp uint, trace [][]big.Int,
	instance io.FunctionInstance) [][]big.Int {
	//
	interpreter := NewInterpreter(p.program)
	// Initialise state
	init := interpreter.Bind(fid, instance.Inputs)
	biStamp := big.NewInt(int64(stamp))
	//
	interpreter.Enter(fid, init)
	//
	for !interpreter.HasTerminated() {
		pc := big.NewInt(int64(interpreter.State().pc))
		//
		trace[0] = append(trace[0], *biStamp)
		trace[1] = append(trace[1], *pc)
		// Execute
		interpreter.Execute(1)
		// Record register state
		for i, v := range interpreter.State().registers {
			// Clone value to prevent side-effects
			var ith big.Int
			//
			ith.Set(&v)
			// Update column for this register.
			trace[i+2] = append(trace[i+2], ith)
		}
	}
	//
	return trace
}
