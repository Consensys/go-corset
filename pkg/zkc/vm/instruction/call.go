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
package instruction

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Call represents an arbitrary memory read operation for a given type of
// memory.
type Call[W any] struct {
	// Module identifyer for memory being read.
	Id uint
	// Return registers for function call.  That is, registers which should hold
	// the result of the call after it has completed.
	Returns []register.Id
	// Argumwent registers for function call.  That is, registers which should
	// hold the arguments for the call which are used to initialise the callee
	// frame.
	Arguments []register.Id
}

// NewCall constructs a new instruction which reads the value from either a
// Random Access Memory (RAM) or a Read-Only Memory (ROM).
func NewCall[W any](id uint, targets []register.Id, sources []register.Id) *Call[W] {
	return &Call[W]{id, targets, sources}
}

// Uses implementation for Instruction interface
func (p *Call[W]) Uses() []register.Id {
	return p.Arguments
}

// Definitions implementation for Instruction interface
func (p *Call[W]) Definitions() []register.Id {
	return p.Returns
}

func (p *Call[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	if len(p.Returns) > 0 {
		builder.WriteString(registersToString(mapping, array.Reverse(p.Returns)...))
		builder.WriteString(" = ")
	}
	//
	builder.WriteString(fmt.Sprintf("%s(", mapping.Module(p.Id).Name()))
	//
	for i, rid := range p.Arguments {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(mapping.Register(rid).Name())
	}
	//
	builder.WriteString(")")
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Call[W]) MicroValidate(_ uint, field field.Config, _ SystemMap[W]) []error {
	return nil
}
