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
	return expandTrace(p)
}

// LowerMacroTrace this macro trace into a micro trace according to a given
// lowering config.
func LowerMacroTrace(cfg LoweringConfig, trace MacroTrace) MicroTrace {
	// var (
	// 	macroProgram   = trace.Program()
	// 	microProgram   = Lower(cfg, trace.Program())
	// 	microInstances = make([]io.FunctionInstance, len(trace.Instances()))
	// )
	// //
	// for i, inst := range trace.Instances() {
	// 	microInstances[i] = io.SplitInstance(cfg.MaxRegisterWidth, inst, macroProgram)
	// }
	// // Done
	// return io.NewTrace(microProgram, microInstances...)
	panic("todo")
}

// ============================================================================
// Helper
// ============================================================================

// sets the maximum width of the program counter.
const pc_width = uint(8)

func expandTrace[T io.Instruction[T]](trace io.Trace[T]) []tr.RawColumn {
	var (
		columns []tr.RawColumn
		program = trace.Program()
		tracer  = NewTracingExecutor(program)
	)
	//
	for _, instance := range trace.Instances() {
		fn := trace.Program().Function(instance.Function)
		//
		arguments := extractInstanceArguments(instance, fn)
		// Trace function (and any subsequent calls)
		tracer.Read(instance.Function, arguments)
	}
	//
	for i, fn := range program.Functions() {
		ith_traces := tracer.Traces(uint(i))
		ith_columns := expandFunctionTrace(*fn, ith_traces)
		columns = append(columns, ith_columns...)
	}
	//
	return columns
}

func expandFunctionTrace[T io.Instruction[T]](fn io.Function[T], traces []io.State) []tr.RawColumn {
	var (
		n       = len(fn.Registers())
		data    = make([][]big.Int, n+2)
		columns = make([]tr.RawColumn, n+2)
		stamp   = int64(0)
	)
	// Initialise data columns
	for i := 0; i < n+2; i++ {
		data[i] = make([]big.Int, len(traces))
	}
	// Fill data columns
	for i, ith := range traces {
		pc := big.NewInt(int64(ith.Pc))
		// Check for new instance
		if ith.Pc == 0 {
			stamp++
		}
		//
		st := big.NewInt(stamp)
		// Write control values
		data[n][i] = *st
		data[n+1][i] = *pc
		// Write registers
		for j, v := range ith.State {
			// Clone value to prevent side-effects
			var ith big.Int
			// Update column for this register.
			data[j][i] = *ith.Set(&v)
		}
	}
	// Finalise stamp column
	columns[n] = tr.RawColumn{Module: fn.Name(), Name: "$stamp",
		Data: field.FrArrayFromBigInts(32, data[n]),
	}
	// Finalise pc column
	columns[n+1] = tr.RawColumn{Module: fn.Name(), Name: "$pc",
		Data: field.FrArrayFromBigInts(pc_width, data[n+1]),
	}
	// Finalise register columns.
	for i, r := range fn.Registers() {
		data := field.FrArrayFromBigInts(r.Width, data[i])
		columns[i] = tr.RawColumn{Module: fn.Name(), Name: r.Name,
			Data: data,
		}
	}
	// Done
	return columns
}
