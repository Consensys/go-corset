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
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Cast represents a truncating cast instruction of the following form:
//
//	t := (uN)s
//
// Here, t is the target register, s is the source register, and N is the cast
// bit width.  The N low-order bits of s are retained and written to t.
type Cast[W word.Word[W]] struct {
	// Target register for assignment
	Target register.Id
	// Source register
	Source register.Id
	// Width is the target bit width for truncation.
	Width uint
}

// NewCast constructs a new truncating cast instruction.
func NewCast[W word.Word[W]](target register.Id, source register.Id, width uint) *Cast[W] {
	return &Cast[W]{target, source, width}
}

// Uses implementation for Instruction interface.
func (p *Cast[W]) Uses() []register.Id {
	return []register.Id{p.Source}
}

// Definitions implementation for Instruction interface.
func (p *Cast[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

// String implementation for Instruction interface.
func (p *Cast[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, p.Target))
	fmt.Fprintf(&builder, " = (u%d) ", p.Width)
	builder.WriteString(registersToString(mapping, p.Source))
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Cast[W]) MicroValidate(_ uint, _ field.Config, _ SystemMap[W]) []error {
	return nil
}
