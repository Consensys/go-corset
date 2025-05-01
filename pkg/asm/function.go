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
	"math"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/insn"
)

// Function defines a distinct functional entity within the system.  Functions
// accepts zero or more inputs and produce zero or more outputs.  Functions
// declare zero or more internal registers for use, and their interpretation is
// given by a sequence of zero or more instructions.
type Function struct {
	// Unique name of this function.
	Name string
	// Registers describes zero or more registers of a given width.  Each
	// register can be designated as an input / output or temporary.
	Registers []insn.Register
	// Code defines the body of this function.
	Code []insn.Instruction
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

// CheckInstance checks whether a given function instance is valid with respect
// to a given set of functions.  It returns an error if something goes wrong
// (e.g. the instance is malformed), and either true or false to indicate
// whether the trace is accepted or not.
func CheckInstance(instance FunctionInstance, fns []Function) (uint, error) {
	// Initialise a new interpreter
	interpreter := NewInterpreter(fns...)
	//
	init := interpreter.Bind(instance.Function, instance.Inputs)
	// Enter function
	interpreter.Enter(instance.Function, init)
	// Execute function to completion
	interpreter.Execute(math.MaxUint)
	// Extract outputs
	outputs := interpreter.Leave()
	// Checkout results
	for r, actual := range outputs {
		expected, ok := instance.Outputs[r]
		outcome := expected.Cmp(&actual) == 0
		// Check actual output matches expected output
		if !ok {
			return math.MaxUint, fmt.Errorf("missing output (%s)", r)
		} else if !outcome {
			// failure
			return 1, fmt.Errorf("incorrect output \"%s\" (was %s, expected %s)", r, actual.String(), expected.String())
		}
	}
	//
	if len(outputs) != len(instance.Outputs) {
		msg := fmt.Errorf("incorrect number of outputs (was %d but expected %d)", len(outputs), len(instance.Outputs))
		return math.MaxUint, msg
	}
	// Success
	return 0, nil
}
