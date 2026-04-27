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
package micro

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/register"
)

// InOut captures input / output instructions for reading / writing to a bus.
type InOut struct {
	// Indicates whether Input or output instruction.
	Input bool
	// Local DataBus
	DataBus io.Bus
}

// NewIoRead constructs an instruction responsible for reading data to a given
// bus.
func NewIoRead(bus io.Bus) *InOut {
	return &InOut{true, bus}
}

// NewIoWrite constructs an instruction responsible for writing data to a given
// bus.
func NewIoWrite(bus io.Bus) *InOut {
	return &InOut{false, bus}
}

// Bus returns information about the bus.  Observe that prior to Link being
// called, this will return an unlinked bus.
func (p *InOut) Bus() io.Bus {
	return p.DataBus
}

// Clone this micro code.
func (p *InOut) Clone() Code {
	return &InOut{p.Input, p.DataBus}
}

// MicroExecute a given micro-code, using a given local state.  This may update
// the register values, and returns either the number of micro-codes to "skip
// over" when executing the enclosing instruction or, if skip==0, a destination
// program counter (which can signal return of enclosing function).
func (p *InOut) MicroExecute(state io.State) (uint, uint) {
	if p.Input {
		state.In(p.DataBus)
	} else {
		state.Out(p.DataBus)
	}
	//
	return 1, 0
}

// RegistersRead returns the set of registers read by this instruction.
func (p *InOut) RegistersRead() []io.RegisterId {
	if p.Input {
		return p.DataBus.Address()
	}
	//
	return append(p.DataBus.Address(), p.DataBus.Data()...)
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *InOut) RegistersWritten() []io.RegisterId {
	if p.Input {
		return p.DataBus.Data()
	}
	//
	return nil
}

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.
func (p *InOut) Split(mapping register.LimbsMap, _ agnostic.RegisterAllocator) []Code {
	// Split bus
	address := register.ApplyLimbsMap(mapping, p.DataBus.Address()...)
	data := register.ApplyLimbsMap(mapping, p.DataBus.Data()...)
	bus := io.NewBus(p.DataBus.Name, p.DataBus.BusId, address, data)
	// Done
	return []Code{&InOut{p.Input, bus}}
}

func (p *InOut) String(fn register.Map) string {
	if p.Input {
		return fmt.Sprintf("in %s", p.DataBus.Name)
	}

	return fmt.Sprintf("out %s", p.DataBus.Name)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *InOut) Validate(fieldWidth uint, fn register.Map) error {
	return nil
}
