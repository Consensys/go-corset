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
package assembler

import (
	"fmt"
	"math"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
)

// Environment captures useful information used during the assembling process.
type Environment struct {
	// Labels identifies branch targets.
	labels []Label
	// Buses identifies connections with external peripherals.
	buses []string
	// Registers identifies set of declared registers.
	registers []io.Register
}

// BindBus associates a bus name with an abstract bus index.  The latter needs
// to be subsequently "aligned" with external bus definitions.
func (p *Environment) BindBus(name string) uint {
	panic("todo")
}

// BindLabel associates a label with a given index which can subsequently be
// used to determine a concrete program counter value.
func (p *Environment) BindLabel(name string) uint {
	// Check whether label already declared.
	for i, lab := range p.labels {
		if lab.name == name {
			return uint(i)
		}
	}
	// Determine index for new label
	index := uint(len(p.labels))
	// Create new label
	p.labels = append(p.labels, UnboundLabel(name))
	// Done
	return index
}

// DeclareLabel declares a given label at a given program counter position.  If
// a label with the same name already exists, this will panic.
func (p *Environment) DeclareLabel(name string, pc uint) {
	// First, check whether the label already exists
	for i, lab := range p.labels {
		if lab.name == name {
			if lab.pc == math.MaxUint {
				p.labels[i].pc = pc
				return
			}
			//
			panic("label already bound")
		}
	}
	// Create new label
	p.labels = append(p.labels, BoundLabel(name, pc))
}

// DeclareRegister declares a new register with the given name and bitwidth.  If
// a register with the same name already exists, this panics.
func (p *Environment) DeclareRegister(kind uint8, name string, width uint) {
	if p.IsRegister(name) {
		panic(fmt.Sprintf("register %s already declared", name))
	}
	//
	p.registers = append(p.registers, io.NewRegister(kind, name, width))
}

// IsRegister checks whether or not a given name is already declared as a
// register.
func (p *Environment) IsRegister(name string) bool {
	for _, reg := range p.registers {
		if reg.Name == name {
			return true
		}
	}
	//
	return false
}

// IsBoundLabel checks whether or not a given label has already been bound to a
// given PC.
func (p *Environment) IsBoundLabel(name string) bool {
	for _, l := range p.labels {
		if l.name == name && l.pc != math.MaxUint {
			return true
		}
	}
	//
	return false
}

// LookupRegister looks up the index for a given register.
func (p *Environment) LookupRegister(name string) uint {
	for i, reg := range p.registers {
		if reg.Name == name {
			return uint(i)
		}
	}
	//
	panic(fmt.Sprintf("unknown register %s", name))
}

// BindLabels processes a given set of instructions by mapping their label
// indexes to concrete program counter locations.
func (p *Environment) BindLabels(insns []macro.Instruction) {
	labels := make([]uint, len(p.labels))
	// Initial the label map
	for i := range labels {
		labels[i] = p.labels[i].pc
		// sanity check
		if labels[i] == math.MaxUint {
			panic(fmt.Sprintf("unbound label \"%s\"", p.labels[i].name))
		}
	}
	// Bind labels using the map
	for _, insn := range insns {
		insn.Bind(labels)
	}
}
