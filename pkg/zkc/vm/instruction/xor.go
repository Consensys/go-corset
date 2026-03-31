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

// Xor represents an instruction of the following form:
//
// t := r0 ^ ... ^ rn ^ c
//
// Here, t is the target register, r0 .. rn are the source registers, and c is
// a constant (which defaults to zero, i.e. the identity element for bitwise XOR).
type Xor[W word.Word[W]] struct {
	// Target register for assignment
	Target register.Id
	// Source registers for assignment
	Sources []register.Id
	// Constant for assignment
	Constant W
}

// NewXor constructs a new bitwise XOR instruction.
func NewXor[W word.Word[W]](target register.Id, sources []register.Id, constant W) *Xor[W] {
	return &Xor[W]{target, sources, constant}
}

// Uses implementation for Instruction interface.
func (p *Xor[W]) Uses() []register.Id {
	return p.Sources
}

// Definitions implementation for Instruction interface.
func (p *Xor[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

// String implementation for Instruction interface.
func (p *Xor[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, p.Target))
	builder.WriteString(" = ")
	builder.WriteString(expressionToString("^", p.Sources, p.Constant, mapping))
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Xor[W]) MicroValidate(_ uint, _ field.Config, _ SystemMap[W]) []error {
	return nil
}
