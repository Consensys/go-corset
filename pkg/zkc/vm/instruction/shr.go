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
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Shr represents an instruction of the following form:
//
// t := v >> a
//
// Here, t is the target register, v is the value register, and a is the shift
// amount register.
type Shr[W word.Word[W]] struct {
	// Target register for assignment
	Target register.Id
	// Value register (the operand being shifted)
	Value register.Id
	// Amount register (the shift amount)
	Amount register.Id
}

// NewShr constructs a new bitwise right-shift instruction.
func NewShr[W word.Word[W]](target register.Id, value register.Id, amount register.Id) *Shr[W] {
	return &Shr[W]{target, value, amount}
}

// Uses implementation for Instruction interface.
func (p *Shr[W]) Uses() []register.Id {
	return []register.Id{p.Value, p.Amount}
}

// Definitions implementation for Instruction interface.
func (p *Shr[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

// String implementation for Instruction interface.
func (p *Shr[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, p.Target))
	builder.WriteString(" = ")
	builder.WriteString(registersToString(mapping, p.Value))
	builder.WriteString(" >> ")
	builder.WriteString(registersToString(mapping, p.Amount))
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Shr[W]) MicroValidate(_ uint, field field.Config, _ SystemMap[W]) []error {
	return nil
}
