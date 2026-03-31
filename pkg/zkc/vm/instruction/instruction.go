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
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
)

// Module represents an either a function or memory within the machine.
type Module[W any] interface {
	// Name of this module
	Name() string
}

// SystemMap provides a global view of modules in the systemn.
type SystemMap[W any] interface {
	register.Map
	//
	Module(id uint) Module[W]
}

// Instruction provides an abstract notion of a "machine instruction".  That is, a single atomic unit which can be
type Instruction[W any] interface {
	// Uses returns the set of variables used (i.e. read) by this instruction.
	Uses() []register.Id
	// Definitions returns the set of variables registers defined (i.e. written)
	// by this instruction.
	Definitions() []register.Id
	// Provide human readable form of instruction
	String(SystemMap[W]) string
}

// NewSystemMap constructs a new system map
func NewSystemMap[W any](regs register.Map, modules []Module[W]) SystemMap[W] {
	return &systemMap[W]{regs, modules}
}

type systemMap[W any] struct {
	regs    register.Map
	modules []Module[W]
}

func (p *systemMap[W]) Module(id uint) Module[W] {
	return p.modules[id]
}

func (p *systemMap[W]) Name() trace.ModuleName {
	return p.regs.Name()
}

func (p *systemMap[W]) HasRegister(name string) (register.Id, bool) {
	return p.regs.HasRegister(name)
}

func (p *systemMap[W]) Register(id register.Id) register.Register {
	return p.regs.Register(id)
}

func (p *systemMap[W]) Registers() []register.Register {
	return p.regs.Registers()
}

func (p *systemMap[W]) String() string {
	return p.regs.String()
}
