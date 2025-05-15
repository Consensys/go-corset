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
package io

import (
	"fmt"
	"math/big"
)

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

// Trace represents the trace of a given program (either macro or micro).
type Trace[T any] interface {
	// Program for which this is a trace of
	Program() Program[T]
	// Input / Outputs of all functions
	Instances() []FunctionInstance
	// Insert all instances into this trace
	InsertAll(instances []FunctionInstance)
}

// NewTrace constructs a new trace for a given program from a given set of
// instances.
func NewTrace[T any](program Program[T], instances ...FunctionInstance) Trace[T] {
	return &trace[T]{program, instances}
}

// SplitInstance an instance applicable for a given set of registers to one fitting a
// given maximum register width.
func SplitInstance[T any](maxRegisterWidth uint, instance FunctionInstance, program Program[T]) FunctionInstance {
	var (
		inputs  map[string]big.Int = make(map[string]big.Int)
		outputs map[string]big.Int = make(map[string]big.Int)
		fn                         = program.Function(instance.Function)
	)
	//
	for _, reg := range fn.Registers() {
		if reg.IsInput() {
			input, ok := instance.Inputs[reg.Name]
			//
			if !ok {
				panic(fmt.Sprintf("missing value for input register %s", reg.Name))
			}
			//
			inputs = SplitRegisterValue(maxRegisterWidth, reg, input, inputs)
		} else if reg.IsOutput() {
			output, ok := instance.Outputs[reg.Name]
			//
			if !ok {
				panic(fmt.Sprintf("missing value for output register %s", reg.Name))
			}
			//
			outputs = SplitRegisterValue(maxRegisterWidth, reg, output, outputs)
		}
	}
	//
	return FunctionInstance{Function: instance.Function, Inputs: inputs, Outputs: outputs}
}

// ============================================================================
// Helpers
// ============================================================================

// MicroTrace represents the trace of a micro program.
type trace[T any] struct {
	// Program for which this is a trace of
	program Program[T]
	// Input / Outputs of given function
	instances []FunctionInstance
}

// Program for which this is a trace of
func (p *trace[T]) Program() Program[T] {
	return p.program
}

// Instances returns the input / outputs of all functions
func (p *trace[T]) Instances() []FunctionInstance {
	return p.instances
}

// InsertAll inserts all the given function instances into this trace.
func (p *trace[T]) InsertAll(instances []FunctionInstance) {
	// FIXME: sort and remove duplicates
	p.instances = append(p.instances, instances...)
}
