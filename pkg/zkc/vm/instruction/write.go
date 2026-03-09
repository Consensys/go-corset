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
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
)

// MemWrite represents an arbitrary memory write operation of the following
// form:
//
// mem[t0,...,tn] = s0,..,sn
//
// Here t0,...,tn are the target registers used for determine the address being
// written, whilst s0,...,sn are the source registers which fill the data lines.
type MemWrite struct {
	// Module identifyer for memory being read.
	Id uint
	// Target registers for assignment
	Targets []register.Id
	// Source registers for assignment
	Sources []register.Id
}

// NewMemWrite constructs a new instruction which writes data values to either a
// Random Access Memory (RAM) or a Write-Once Memory (WOM).
func NewMemWrite(id uint, targets []register.Id, sources []register.Id) *MemWrite {
	return &MemWrite{id, targets, sources}
}

// Uses implementation for Instruction interface
func (p *MemWrite) Uses() []register.Id {
	var data set.AnySortedSet[register.Id]
	//
	for _, t := range p.Targets {
		data.Insert(t)
	}
	//
	for _, s := range p.Sources {
		data.Insert(s)
	}
	//
	return data
}

// Definitions implementation for Instruction interface
func (p *MemWrite) Definitions() []register.Id {
	return nil
}

func (p *MemWrite) String(env register.Map) string {
	var builder strings.Builder
	//
	builder.WriteString(fmt.Sprintf("%d[", p.Id))
	builder.WriteString(registersToString(env, array.Reverse(p.Targets)...))
	builder.WriteString("] = ")
	//
	for i, rid := range p.Sources {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(env.Register(rid).Name())
	}
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *MemWrite) MicroValidate(_ uint, field field.Config, env register.Map) []error {
	return nil
}
