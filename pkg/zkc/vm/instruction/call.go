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
type Call struct {
	// Module identifyer for memory being read.
	Id uint
	// Target registers for assignment
	Targets []register.Id
	// Source registers for assignment
	Sources []register.Id
}

// NewCall constructs a new instruction which reads the value from either a
// Random Access Memory (RAM) or a Read-Only Memory (ROM).
func NewCall(id uint, targets []register.Id, sources []register.Id) *Call {
	return &Call{id, targets, sources}
}

// Uses implementation for Instruction interface
func (p *Call) Uses() []register.Id {
	return p.Sources
}

// Definitions implementation for Instruction interface
func (p *Call) Definitions() []register.Id {
	return p.Targets
}

func (p *Call) String(env register.Map) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(env, array.Reverse(p.Targets)...))
	builder.WriteString(" = ")
	//
	builder.WriteString(fmt.Sprintf("%d(", p.Id))
	//
	for i, rid := range p.Sources {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(env.Register(rid).Name())
	}
	//
	builder.WriteString(")")
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Call) MicroValidate(_ uint, field field.Config, env register.Map) []error {
	return nil
}
