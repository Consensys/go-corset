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
package debug

import (
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

type Tracer[W word.Word[W]] struct {
	depth uint
	fun   *function.Function[instruction.Instruction[W]]
	insn  instruction.MicroInstruction[W]
	pc    machine.ProgramCounter
	// list of steps being constructed by this tracer.
	Steps []ExecutionStep[W]
}

// PreExecution implementation for Observer interface
func (p *Tracer[W]) PreExecution(machine *machine.Base[W]) {

}

// PostExecution implementation for Observer interface
func (p *Tracer[W]) PostExecution(machine *machine.Base[W]) {
	var (
		depth = machine.Depth()
	)
	//
	if depth > 0 {
		frame := machine.StackFrame(depth - 1)
		p.NewStep(frame.PC())
	}
}

func (p *Tracer[W]) NewStep(pc machine.ProgramCounter) {
	p.Steps = append(p.Steps, ExecutionStep[W]{Pc: pc})
}
