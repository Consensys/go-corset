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
	"fmt"
	"slices"

	"github.com/consensys/go-corset/pkg/asm/macro"
	"github.com/consensys/go-corset/pkg/asm/micro"
)

// MacroProgram represents a set of functions at the macro level.  Thus, a macro
// program can be lowered to a given micro program, etc.
type MacroProgram struct {
	Functions []Function[macro.Instruction]
}

// MicroProgram represents a set of functions at the micro level.
type MicroProgram struct {
	Functions []Function[micro.Instruction]
}

// Lower a given macro program into a micro program which only uses registers of
// a given width.  This is a relatively involved procress consisting of several
// steps: firstly, all macro instructions are lowered to micro instructions;
// secondly, vectorization is applied to the resulting microprogram; finally,
// registers exceeding the target width (and instructions which use them) are
// split accordingly.  The latter can introduce additional registers, for
// example to hold carry values.
func (p *MacroProgram) Lower(maxwidth uint) MicroProgram {
	functions := make([]MicroFunction, len(p.Functions))
	//
	for i, f := range p.Functions {
		functions[i] = lowerFunction(maxwidth, f)
	}
	//
	return MicroProgram{Functions: functions}
}

func lowerFunction(maxwidth uint, f MacroFunction) MicroFunction {
	insns := make([]micro.Instruction, len(f.Code))
	//
	for i, insn := range f.Code {
		insns[i] = insn.Lower()
	}
	// Sanity checks (for now)
	for _, reg := range f.Registers {
		if reg.Width > maxwidth {
			panic(fmt.Sprintf("register %s exceeds max width", reg.Name))
		}
	}
	//
	regs := slices.Clone(f.Registers)
	//
	return MicroFunction{f.Name, regs, insns}
}
