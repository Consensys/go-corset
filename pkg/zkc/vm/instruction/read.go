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

// MemRead represents an arbitrary memory read operation for a given type of
// memory.
type MemRead[W any] struct {
	// Module identifyer for memory being read.
	Id uint
	// Data registers for assignment
	Data []register.Id
	// Address registers for assignment
	Address []register.Id
}

// NewMemRead constructs a new instruction which reads the value from either a
// Random Access Memory (RAM) or a Read-Only Memory (ROM).
func NewMemRead[W any](id uint, data []register.Id, address []register.Id) *MemRead[W] {
	return &MemRead[W]{id, data, address}
}

// Uses implementation for Instruction interface
func (p *MemRead[W]) Uses() []register.Id {
	return p.Address
}

// Definitions implementation for Instruction interface
func (p *MemRead[W]) Definitions() []register.Id {
	return p.Data
}

func (p *MemRead[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, array.Reverse(p.Data)...))
	builder.WriteString(" = ")
	//
	builder.WriteString(fmt.Sprintf("%s[", mapping.Module(p.Id).Name()))
	//
	for i, rid := range p.Address {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(mapping.Register(rid).Name())
	}
	//
	builder.WriteString("]")
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *MemRead[W]) MicroValidate(_ uint, field field.Config, _ SystemMap[W]) []error {
	return nil
}
