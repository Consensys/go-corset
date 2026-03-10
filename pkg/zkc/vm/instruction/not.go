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

// Not represents an instruction of the following form:
//
// t := ~r
//
// Here, t is the target register and r is the source register. The complement
// is computed within the bit width of the source register.
type Not[W word.Word[W]] struct {
	// Target register for assignment
	Target register.Id
	// Source register
	Source register.Id
}

// NewNot constructs a new bitwise NOT instruction.
func NewNot[W word.Word[W]](target register.Id, source register.Id) *Not[W] {
	return &Not[W]{target, source}
}

// Uses implementation for Instruction interface.
func (p *Not[W]) Uses() []register.Id {
	return []register.Id{p.Source}
}

// Definitions implementation for Instruction interface.
func (p *Not[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

// String implementation for Instruction interface.
func (p *Not[W]) String(mapping register.Map) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, p.Target))
	builder.WriteString(" = ~")
	builder.WriteString(registersToString(mapping, p.Source))
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Not[W]) MicroValidate(_ uint, field field.Config, env register.Map) []error {
	return nil
}
