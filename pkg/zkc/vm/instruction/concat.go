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

// Concat represents an instruction of the following form:
//
// tn, .., t0 := r0::...::rn
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, r0 .. rn are the source registers which are
// concatenated together in bitwise form.
type Concat[W word.Word[W]] struct {
	// Target register for assignment
	Target register.Id
	// Source registers for assignment
	Sources []register.Id
}

// NewConcat constructs a new concatenation instruction which concatenates the
// source registers and writes them into the target register.  Observe that we
// have a little endian ordering here for the source registers.  That is, the
// value of the register sources[0] will occupy the least significant bits of
// the result.
func NewConcat[W word.Word[W]](target register.Id, sources []register.Id) *Concat[W] {
	return &Concat[W]{target, sources}
}

// Uses implementation for Instruction interface
func (p *Concat[W]) Uses() []register.Id {
	return p.Sources
}

// Definitions implementation for Instruction interface
func (p *Concat[W]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

func (p *Concat[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(mapping, p.Target))
	builder.WriteString(" = ")
	//
	for i := 0; i < len(p.Sources); i++ {
		var rid = p.Sources[i]
		//
		if i != 0 {
			builder.WriteString("::")
		}
		//
		builder.WriteString(mapping.Register(rid).Name())
	}
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Concat[W]) MicroValidate(_ uint, field field.Config, _ SystemMap[W]) []error {
	return nil
}
