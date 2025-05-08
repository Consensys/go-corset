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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/asm/macro"
	"github.com/consensys/go-corset/pkg/asm/micro"
)

// MacroFunction is a function whose instructions are themselves macro
// instructions.  A macro function must be compiled down into a micro function
// before we can generate constraints.
type MacroFunction = Function[macro.Instruction]

// MicroFunction is a function whose instructions are themselves micro
// instructions.  A micro function represents the lowest representation of a
// function, where each instruction is made up of microcodes.
type MicroFunction = Function[micro.Instruction]

// Function defines a distinct functional entity within the system.  Functions
// accepts zero or more inputs and produce zero or more outputs.  Functions
// declare zero or more internal registers for use, and their interpretation is
// given by a sequence of zero or more instructions.
type Function[T any] struct {
	// Unique name of this function.
	Name string
	// Registers describes zero or more registers of a given width.  Each
	// register can be designated as an input / output or temporary.
	Registers []insn.Register
	// Code defines the body of this function.
	Code []T
}

// FunctionInstance represents a specific instance of a function.  That is, a
// mapping from input values to expected output values.
type FunctionInstance struct {
	// Identifies corresponding function.
	Function uint
	// Inputs identifies the input arguments
	Inputs map[string]big.Int
	// Outputs identifies the outputs
	Outputs map[string]big.Int
}

// Lower a function instance for a given program to a function instance for the
// corresponding microprogram.
func (p *FunctionInstance) Lower(cfg LoweringConfig, program MacroProgram) FunctionInstance {
	var (
		maxWidth                    = cfg.MaxRegisterWidth
		inputs   map[string]big.Int = make(map[string]big.Int)
		outputs  map[string]big.Int = make(map[string]big.Int)
		fn                          = program.Function(p.Function)
	)
	//
	for _, reg := range fn.Registers {
		if reg.IsInput() {
			input, ok := p.Inputs[reg.Name]
			//
			if !ok {
				panic(fmt.Sprintf("missing value for input register %s", reg.Name))
			}
			//
			inputs = splitRegisterValue(maxWidth, reg, input, inputs)
		} else if reg.IsOutput() {
			output, ok := p.Outputs[reg.Name]
			//
			if !ok {
				panic(fmt.Sprintf("missing value for output register %s", reg.Name))
			}
			//
			outputs = splitRegisterValue(maxWidth, reg, output, outputs)
		}
	}
	//
	return FunctionInstance{Function: p.Function, Inputs: inputs, Outputs: outputs}
}

func splitRegisterValue(maxWidth uint, reg Register, value big.Int, iomap map[string]big.Int) map[string]big.Int {
	var (
		nlimbs = micro.NumberOfLimbs(maxWidth, reg.Width)
	)
	//
	if nlimbs == 1 {
		// no splitting required
		iomap[reg.Name] = value
	} else {
		// splitting required
		regs := micro.SplitRegister(maxWidth, reg)
		values := micro.SplitValueAcrossRegisters(&value, regs...)
		//
		for i, limb := range regs {
			iomap[limb.Name] = values[i]
		}
	}
	//
	return iomap
}
