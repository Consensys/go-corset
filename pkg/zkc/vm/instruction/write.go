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
type MemWrite[W any] struct {
	// Module identifyer for memory being read.
	Id uint
	// Address registers for assignment
	Address []register.Id
	// Data registers for assignment
	Data []register.Id
}

// NewMemWrite constructs a new instruction which writes data values to either a
// Random Access Memory (RAM) or a Write-Once Memory (WOM).
func NewMemWrite[W any](id uint, targets []register.Id, sources []register.Id) *MemWrite[W] {
	return &MemWrite[W]{id, targets, sources}
}

// Uses implementation for Instruction interface
func (p *MemWrite[W]) Uses() []register.Id {
	var data set.AnySortedSet[register.Id]
	//
	for _, t := range p.Address {
		data.Insert(t)
	}
	//
	for _, s := range p.Data {
		data.Insert(s)
	}
	//
	return data
}

// Definitions implementation for Instruction interface
func (p *MemWrite[W]) Definitions() []register.Id {
	return nil
}

func (p *MemWrite[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	builder.WriteString(fmt.Sprintf("%s[", mapping.Module(p.Id).Name()))
	builder.WriteString(registersToString(mapping, array.Reverse(p.Address)...))
	builder.WriteString("] = ")
	//
	for i, rid := range p.Data {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(mapping.Register(rid).Name())
	}
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *MemWrite[W]) MicroValidate(_ uint, _ field.Config, env SystemMap[W]) []error {
	return nil
}
