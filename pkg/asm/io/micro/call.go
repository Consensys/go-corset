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
	"slices"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
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

var _ io.BusInstruction = &Call{}

// BusId returns the bus that this instruction accesses.
func (p *Call) BusId() uint {
	//
	return p.Bus
}

// Clone this micro code.
func (p *Call) Clone() Code {
	//
	return &Call{
		p.Bus,
		slices.Clone(p.Targets),
		slices.Clone(p.Sources),
	}
}

// MicroExecute a given micro-code, using a given local state.  This may update
// the register values, and returns either the number of micro-codes to "skip
// over" when executing the enclosing instruction or, if skip==0, a destination
// program counter (which can signal return of enclosing function).
func (p *Call) MicroExecute(state io.State, iomap io.Map) (uint, uint) {
	// Setup read address
	address := state.ReadN(p.Sources)
	// Perform I/O read
	values := iomap.Read(p.Bus, address)
	// Write back results
	state.WriteN(p.Targets, values)
	//
	return 1, 0
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

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.  Here, regsBefore
// represents the registers are they are for this code, whilst regsAfter
// represent those for the resulting split codes.  The regMap provides a
// mapping from registers in regsBefore to those in regsAfter.
func (p *Call) Split(env *RegisterSplittingEnvironment) []Code {
	targets := env.SplitTargetRegisters(p.Targets...)
	sources := env.SplitTargetRegisters(p.Sources...)
	//
	code := &Call{
		Bus:     p.Bus,
		Targets: targets,
		Sources: sources,
	}
	//
	return []Code{code}
}

// Validate checks whether or not this instruction well-formed.
func (p *Call) Validate(env io.Environment[Instruction]) error {
	return ValidateCall(p.Bus, p.Targets, p.Sources, env)
}

// ValidateCall validates a calling instruction.  This is to avoid code
// duplication between micro and macro instructions.
func ValidateCall[T any](b uint, targets []uint, sources []uint, env io.Environment[T]) error {
	var (
		fn  = env.Enclosing()
		fns = env.Program.Functions()
	)
	// Check bus is assigned
	if b == io.UNKNOWN_BUS {
		return fmt.Errorf("unknown function")
	} else if b >= uint(len(fns)) {
		return fmt.Errorf("invalid function")
	}
	// Sanity check arguments and returns
	bus := fns[b]
	busInputs := bus.Inputs()
	busOutputs := bus.Outputs()
	//
	if len(busInputs) != len(sources) {
		return fmt.Errorf("incorrect arguments (found %d expected %d)", len(sources), len(busInputs))
	} else if len(busOutputs) != len(targets) {
		return fmt.Errorf("incorrect returns (found %d expected %d)", len(targets), len(busOutputs))
	}
	// Check arguments
	for i, src := range sources {
		src_w := fn.Registers[src].Width
		bus_w := busInputs[i].Width
		//
		if src_w != bus_w {
			return fmt.Errorf("incorrect width for argument %d (found %d expected %d)", i+1, src_w, bus_w)
		}
	}
	// Check returns
	for i, rtn := range targets {
		rtn_w := fn.Registers[rtn].Width
		bus_w := busOutputs[i].Width
		//
		if rtn_w != bus_w {
			return fmt.Errorf("incorrect width for return %d (found %d expected %d)", i+1, rtn_w, bus_w)
		}
	}
	// Done
	return nil
}
