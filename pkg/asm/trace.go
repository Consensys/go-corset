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
	"github.com/consensys/go-corset/pkg/asm/macro"
	"github.com/consensys/go-corset/pkg/asm/micro"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Trace represents the trace of a given program (either macro or micro).
type Trace[T any] interface {
	// Program for which this is a trace of
	Program() Program[T]
	// Input / Outputs of all functions
	Instances() []FunctionInstance
	// Insert all instances into this trace
	InsertAll(instances []FunctionInstance)
}

// LowerTraces lowers macro-level traces to micro-level traces according to a given lowering config.
func LowerTraces(config LoweringConfig, traces ...Trace[macro.Instruction]) []Trace[micro.Instruction] {
	utraces := make([]Trace[micro.Instruction], len(traces))
	//
	for i, tr := range traces {
		mtr := tr.(*MacroTrace) // ugly
		utrace := mtr.Lower(config)
		utraces[i] = &utrace
	}
	//
	return utraces
}

// ============================================================================
// Micro trace
// ============================================================================

// MicroTrace represents the trace of a micro program.
type MicroTrace struct {
	// Program for which this is a trace of
	program MicroProgram
	// Input / Outputs of given function
	instances []FunctionInstance
}

// Program for which this is a trace of
func (p *MicroTrace) Program() Program[micro.Instruction] {
	return &p.program
}

// Instances returns the input / outputs of all functions
func (p *MicroTrace) Instances() []FunctionInstance {
	return p.instances
}

// InsertAll inserts all the given function instances into this trace.
func (p *MicroTrace) InsertAll(instances []FunctionInstance) {
	// FIXME: sort and remove duplicates
	p.instances = append(p.instances, instances...)
}

// Lower this micro trace to a set of raw columns.
func (p *MicroTrace) Lower() []tr.RawColumn {
	builder := NewTraceBuilder(&p.program)
	return builder.Build(p)
}

// ============================================================================
// Macro trace
// ============================================================================

// MacroTrace represents the trace of a macro program.
type MacroTrace struct {
	// Program for which this is a trace of
	program MacroProgram
	// Input / Outputs of given function
	instances []FunctionInstance
}

// Program for which this is a trace of
func (p *MacroTrace) Program() Program[macro.Instruction] {
	return &p.program
}

// Instances returns the input / outputs of all functions
func (p *MacroTrace) Instances() []FunctionInstance {
	return p.instances
}

// InsertAll inserts all the given function instances into this trace.
func (p *MacroTrace) InsertAll(instances []FunctionInstance) {
	// FIXME: sort and remove duplicates
	p.instances = append(p.instances, instances...)
}

// Lower this macro trace into a micro trace according to a given lowering
// config.
func (p *MacroTrace) Lower(cfg LoweringConfig) MicroTrace {
	var (
		microProgram                      = p.program.Lower(cfg)
		microInstances []FunctionInstance = make([]FunctionInstance, len(p.instances))
	)
	//
	for i, inst := range p.instances {
		microInstances[i] = inst.Lower(cfg, p.program)
	}
	//
	return MicroTrace{microProgram, microInstances}
}
