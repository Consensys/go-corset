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
package decl

import (
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Function contains information about an executable function in the system.  A
// function has one or more variables where: the first n are the parameters; the
// next m are the returns; and all remaining registers are internal.
// Additionally, a function has some number of "instructions" which capture its
// semantics (i.e. intended behaviour).  The notion of an instruction is
// specifically left undefined by this interface to support different levels of
// the compilation pipeline.  For example, a compiled function has instructions
// which are simply bytes (or words) for efficient execution.  However, the
// instructions of an "assembly" level function implement the Instruction
// interface, which is better suited to analysis and/or translation into
// constraints.
type Function[E any] struct {
	// Unique name of this function.
	name string
	// Registers describes zero or more variables of a given width.  Each
	// register can be designated as an input / output or temporary.
	Variables []variable.Descriptor
	// Number of input variables
	NumInputs uint
	// Number of output variables
	NumOutputs uint
	// Code defines the body of this function.
	Code []stmt.Instruction[E]
}

// NewFunction constructs a new function with the given variables and code
func NewFunction[E any](name string, variables []variable.Descriptor, code []stmt.Instruction[E]) *Function[E] {
	var (
		numInputs  = array.CountMatching(variables, func(r variable.Descriptor) bool { return r.IsParameter() })
		numOutputs = array.CountMatching(variables, func(r variable.Descriptor) bool { return r.IsReturn() })
	)
	//
	return &Function[E]{name, variables, numInputs, numOutputs, code}
}

// Name implementation for Declaration interface
func (p *Function[E]) Name() string {
	return p.name
}

// Externs implementation for Declaration interface
func (p *Function[E]) Externs() []E {
	panic("todo")
}

// Variable implementation for variable.Map interface
func (p *Function[E]) Variable(id variable.Id) variable.Descriptor {
	return p.Variables[id]
}
