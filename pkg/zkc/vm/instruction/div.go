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

// Div represents an instruction of the following form:
//
// t := d / r
//
// Here, t is the target register, d is the dividend register, and r is the
// divisor register.  The instruction panics at runtime if r is zero.
type Div[W word.Word[W]] struct {
	// Target register for assignment
	Target register.Id
	// Dividend register
	Dividend register.Id
	// Divisor register
	Divisor register.Id
}

// NewDiv constructs a new division instruction.
func NewDiv[W word.Word[W]](target register.Id, dividend register.Id, divisor register.Id) *Div[W] {
	return &Div[W]{target, dividend, divisor}
}

// Uses implementation for Instruction interface.
func (p *Div[W]) Uses() []register.Id {
	return []register.Id{p.Dividend, p.Divisor}
}

// Definitions implementation for Instruction interface.
func (p *Div[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

// String implementation for Instruction interface.
func (p *Div[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, p.Target))
	builder.WriteString(" = ")
	builder.WriteString(registersToString(mapping, p.Dividend))
	builder.WriteString(" / ")
	builder.WriteString(registersToString(mapping, p.Divisor))
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Div[W]) MicroValidate(_ uint, field field.Config, _ SystemMap[W]) []error {
	return nil
}
