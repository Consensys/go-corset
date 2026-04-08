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

// Destruct represents an instruction of the following form:
//
// tn::t0 := r0
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, r0 is the source register which are.
type Destruct[W word.Word[W]] struct {
	// Target registers for assignment
	Targets []register.Id
	// Source register for assignment
	Source register.Id
}

// NewDestruct constructs a new concatenation instruction which concatenates the
// source registers and writes them into the target register.  Observe that we
// have a little endian ordering here for the target registers.  That is, the
// value of the register targets[0] will be assigned the least significant bits of
// the source value.
func NewDestruct[W word.Word[W]](targets []register.Id, source register.Id) *Destruct[W] {
	return &Destruct[W]{targets, source}
}

// Uses implementation for Instruction interface
func (p *Destruct[W]) Uses() []register.Id {
	return []register.Id{p.Source}
}

// Definitions implementation for Instruction interface
func (p *Destruct[W]) Definitions() []register.Id {
	return p.Targets
}

func (p *Destruct[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	for i := 0; i < len(p.Targets); i++ {
		var rid = p.Targets[i]
		//
		if i != 0 {
			builder.WriteString("::")
		}
		//
		builder.WriteString(mapping.Register(rid).Name())
	}
	//
	builder.WriteString(" = ")
	builder.WriteString(registersToString(mapping, p.Source))
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Destruct[W]) MicroValidate(_ uint, field field.Config, _ SystemMap[W]) []error {
	return nil
}
