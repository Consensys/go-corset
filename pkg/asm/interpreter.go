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
	"math"
	"math/big"
)

// Interpreter encapsulates all state needed for executing a given instruction
// sequence.
type Interpreter struct {
	// Set of functions being interpreted
	functions []Function
}

// NewInterpreter intialises an interpreter for executing a given instruction
// sequence.
func NewInterpreter(fns ...Function) *Interpreter {
	return &Interpreter{fns}
}

// Execute the given instruction sequence embodied by this interpreter with the
// given set of initial arguments (i.e. register values), producing the given
// set of outputs.
func (p *Interpreter) Execute(fn uint, arguments map[string]big.Int) map[string]big.Int {
	var (
		f      = p.functions[fn]
		pc     = uint(0)
		regs   = make([]big.Int, len(f.Registers))
		widths = make([]uint, len(f.Registers))
	)
	// Initialise arguments
	for i, reg := range f.Registers {
		if reg.Kind == INPUT_REGISTER {
			regs[i] = arguments[reg.Name]
		}
		//
		widths[i] = reg.Width
	}
	// Continue executing until return signalled.
	for pc != math.MaxUint {
		insn := f.Code[pc]
		pc = insn.Execute(pc, regs, widths)
	}
	// Construct outputs
	outputs := make(map[string]big.Int, 0)
	//
	for i, reg := range f.Registers {
		if reg.Kind == OUTPUT_REGISTER {
			outputs[reg.Name] = regs[i]
		}
	}
	//
	return outputs
}
