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
package macro

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
)

// Call represents a function call providing one or more arguments and accepting
// zero or more values in return.  A function call requires a "bus" to read and
// write its arguments / returns.  A bus is a set of dedicated registers
// providing an I/O communication channel to a given peripheral (in this case,
// another function).
type Call struct {
	// Bus identifies the relevant bus for this instruction.
	bus io.Bus
	// Target registers for addition
	Targets []uint
	// Source registers (i.e. arguments) for call
	Sources []uint
}

// NewCall constructs a new call instruction.
func NewCall(bus io.Bus, targets []uint, sources []uint) *Call {
	return &Call{bus, targets, sources}
}

// Bus returns information about the bus.  Observe that prior to Link being
// called, this will return an unlinked bus.
func (p *Call) Bus() io.Bus {
	return p.bus
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *Call) Execute(state io.State) uint {
	// Setup read address
	address := state.LoadN(p.Sources)
	// Set bus address lines
	state.StoreN(p.bus.Address(), address)
	// Perform I/O read
	state.In(p.bus)
	// Load bus data lines
	values := state.LoadN(p.bus.Data())
	// Write back results
	state.StoreN(p.Targets, values)
	//
	return state.Next()
}

// Link links the bus.  Observe that this can only be called once on any
// given instruction.
func (p *Call) Link(bus io.Bus) {
	if !p.bus.IsUnlinked() {
		panic("bus already linked")
	}
	//
	p.bus = bus
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Call) Lower(pc uint) micro.Instruction {
	var (
		code    []micro.Code
		address = p.bus.Address()
		data    = p.bus.Data()
	)
	// Write address lines
	for i, input := range p.Sources {
		insn := &micro.Add{Targets: []uint{address[i]}, Sources: []uint{input}}
		code = append(code, insn)
	}
	// For read / write on bus
	code = append(code, micro.NewIoRead(p.bus))
	//
	// Read output lines
	for i, output := range p.Targets {
		insn := &micro.Add{Targets: []uint{output}, Sources: []uint{data[i]}}
		code = append(code, insn)
	}
	// Append final branch
	code = append(code, &micro.Jmp{Target: pc + 1})
	// Done
	return micro.NewInstruction(code...)
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Call) RegistersRead() []uint {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Call) RegistersWritten() []uint {
	return p.Targets
}

func (p *Call) String(fn io.Function[Instruction]) string {
	var (
		builder strings.Builder
		regs    = fn.Registers()
	)
	//
	builder.WriteString(io.RegistersToString(p.Targets, regs))
	builder.WriteString(fmt.Sprintf(" = %s(", p.bus.Name))
	builder.WriteString(io.RegistersToString(p.Sources, regs))
	builder.WriteString(")")
	//
	return builder.String()
}

// Validate checks whether or not this instruction well-formed.
func (p *Call) Validate(fieldWidth uint, fn io.Function[Instruction]) error {
	// Check bus is assigned
	if p.bus.IsUnlinked() {
		return fmt.Errorf("unknown function")
	}
	// Sanity check arguments and returns
	busInputs := p.bus.Address()
	busOutputs := p.bus.Data()
	//
	if len(busInputs) != len(p.Sources) {
		return fmt.Errorf("incorrect arguments (found %d expected %d)", len(p.Sources), len(busInputs))
	} else if len(busOutputs) != len(p.Targets) {
		return fmt.Errorf("incorrect returns (found %d expected %d)", len(p.Targets), len(busOutputs))
	}
	// Check arguments
	for i, src := range p.Sources {
		src_w := fn.Register(src).Width
		bus_w := fn.Register(busInputs[i]).Width
		//
		if src_w != bus_w {
			return fmt.Errorf("incorrect width for argument %d (found %d expected %d)", i+1, src_w, bus_w)
		}
	}
	// Check returns
	for i, rtn := range p.Targets {
		rtn_w := fn.Register(rtn).Width
		bus_w := fn.Register(busOutputs[i]).Width
		//
		if rtn_w != bus_w {
			return fmt.Errorf("incorrect width for return %d (found %d expected %d)", i+1, rtn_w, bus_w)
		}
	}
	// Done
	return nil
}
