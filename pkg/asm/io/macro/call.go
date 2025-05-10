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
	Bus uint
	// Target registers for addition
	Targets []uint
	// Source registers (i.e. arguments) for call
	Sources []uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *Call) Bind(labels []uint) {
	// no-op
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *Call) Execute(state io.State, iomap io.Map) uint {
	// Setup read address
	address := state.ReadN(p.Sources)
	// Perform I/O read
	values := iomap.Read(p.Bus, address)
	// Write back results
	state.WriteN(p.Targets, values)
	//
	return state.Next()
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Call) Lower(pc uint) micro.Instruction {
	code := &micro.Call{
		Bus:     p.Bus,
		Targets: p.Targets,
		Sources: p.Sources,
	}
	// Lowering here produces an instruction containing a single microcode.
	return micro.NewInstruction(code, &micro.Jmp{Target: pc + 1})
}

// Link any buses used within this instruction using the given bus map.
func (p *Call) Link(buses []uint) {
	p.Bus = buses[p.Bus]
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Call) RegistersRead() []uint {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Call) RegistersWritten() []uint {
	return p.Targets
}

func (p *Call) String(env io.Environment[Instruction]) string {
	var (
		builder strings.Builder
		regs    = env.Enclosing().Registers
		name    = env.Program.Function(p.Bus).Name
	)
	//
	builder.WriteString(io.RegistersToString(p.Targets, regs))
	builder.WriteString(fmt.Sprintf(" = %s(", name))
	builder.WriteString(io.RegistersToString(p.Sources, regs))
	builder.WriteString(")")
	//
	return builder.String()
}

// Validate checks whether or not this instruction well-formed.
func (p *Call) Validate(env io.Environment[Instruction]) error {
	return micro.ValidateCall(p.Bus, p.Targets, p.Sources, env)
}
