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

// Add represents an instruction of the following form:
//
// tn, .., t0 := r0 + ... + rn + c
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, r0 .. rn are the source registers and c is a
// constant (which can be 0).  For example, consider this case:
//
// c, r0 := r1 + 1
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 + 1 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to c.  Thus, in this
// particular example, c represents a carry flag.
type Add[W word.Word[W]] struct {
	// Target register for assignment
	Target register.Id
	// Source registers for assignment
	Sources []register.Id
	// Constant for assignment
	Constant W
}

// NewAdd constructs a new addition instruction
func NewAdd[W word.Word[W]](target register.Id, sources []register.Id, constant W) *Add[W] {
	return &Add[W]{target, sources, constant}
}

// Uses implementation for Instruction interface
func (p *Add[W]) Uses() []register.Id {
	return p.Sources
}

// Definitions implementation for Instruction interface
func (p *Add[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

func (p *Add[W]) String(mapping register.Map) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, p.Target))
	builder.WriteString(" = ")
	builder.WriteString(expressionToString("+", p.Sources, p.Constant, mapping))
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Add[W]) MicroValidate(_ uint, field field.Config, env register.Map) []error {
	return nil
}
