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
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	tr "github.com/consensys/go-corset/pkg/trace"
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

// Lower this micro trace to a set of raw columns.
func LowerMicroTrace(p MicroTrace) []tr.RawColumn {
	builder := NewTraceBuilder(p.Program())
	return builder.Build(p)
}

// Lower this macro trace into a micro trace according to a given lowering
// config.
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
