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

// ============================================================================
// Function Call
// ============================================================================

// Call invokes the function module identified by Id, passing the values of
// the argument registers as inputs and writing the function's outputs into
// the return registers of the caller's frame.  Execution of the calling
// frame is suspended until the callee returns; on return, control resumes
// at the instruction following the Call.
type Call[W any] struct{ OpIo[W] }

// NewCall constructs a new function call instruction.
func NewCall[W any](id uint, arguments []register.Id, returns []register.Id) *Call[W] {
	return &Call[W]{OpIo[W]{CALL, id, arguments, returns}}
}

// ============================================================================
// Memory Read
// ============================================================================

// MemRead reads from the memory module identified by Id at the address
// formed from the argument registers, depositing the resulting data words
// into the return registers.  The target module must be a Random Access
// Memory (RAM) or a Read-Only Memory (ROM).
type MemRead[W any] struct{ OpIo[W] }

// NewMemRead constructs a new instruction which reads the value from either a
// Random Access Memory (RAM) or a Read-Only Memory (ROM).
func NewMemRead[W any](id uint, address []register.Id, data []register.Id) *MemRead[W] {
	return &MemRead[W]{OpIo[W]{MEMORY_READ, id, address, data}}
}

// ============================================================================
// Memory Write
// ============================================================================

// MemWrite writes to the memory module identified by Id, using the argument
// registers as the data words and the return registers as the address.  The
// target module must be a Random Access Memory (RAM) or a Write-Once Memory
// (WOM).  Despite the name, the "return" registers here identify the
// destination address rather than receiving any output — a MemWrite defines
// no registers in the surrounding frame.
type MemWrite[W any] struct{ OpIo[W] }

// NewMemWrite constructs a new instruction which writes data values to either
// a Random Access Memory (RAM) or a Write-Once Memory (WOM).
func NewMemWrite[W any](id uint, address []register.Id, data []register.Id) *MemWrite[W] {
	return &MemWrite[W]{OpIo[W]{MEMORY_WRITE, id, address, data}}
}

// ============================================================================
// Opcode-Io-Regs-Regs instruction type
// ============================================================================

// OpIo represents an instruction operating on a target module identified
// by Id, with a list of return (output) registers and a list of argument
// (input) registers.  It is similar in shape to a function call, but
// additionally carries an opcode allowing the same structure to be reused for
// distinct operations (e.g. function calls and memory reads/writes).
type OpIo[W any] struct {
	// Op identifies the specific operation this instruction represents.
	Op OpCode
	// Module identifier for the target of the operation.
	Id uint
	// Argument registers providing the inputs to the operation.
	Arguments []register.Id
	// Return registers which receive the outputs of the operation.
	Returns []register.Id
}

// OpCode implementation for Instruction interface
func (p *OpIo[W]) OpCode() OpCode {
	return p.Op
}

// Address is an alias to help identify which are the data lines for a memory
// operation.
func (p *OpIo[W]) Address() []register.Id {
	return p.Arguments
}

// Data is an alias to help identify which are the data lines for a memory
// operation.
func (p *OpIo[W]) Data() []register.Id {
	return p.Returns
}

// Uses implementation for Instruction interface
func (p *OpIo[W]) Uses() []register.Id {
	// A memory write uses both the address and data registers.
	if p.Op == MEMORY_WRITE {
		var data set.AnySortedSet[register.Id]
		//
		for _, t := range p.Returns {
			data.Insert(t)
		}
		//
		for _, s := range p.Arguments {
			data.Insert(s)
		}
		//
		return data
	}
	//
	return p.Arguments
}

// Definitions implementation for Instruction interface
func (p *OpIo[W]) Definitions() []register.Id {
	// A memory write does not define any registers.
	if p.Op == MEMORY_WRITE {
		return nil
	}
	//
	return p.Returns
}

func (p *OpIo[W]) String(mapping SystemMap[W]) string {
	var builder strings.Builder
	//
	switch p.Op {
	case MEMORY_WRITE:
		// mem[address] = data
		fmt.Fprintf(&builder, "%s[", mapping.Module(p.Id).Name())
		builder.WriteString(registersToString(mapping, array.Reverse(p.Returns)...))
		builder.WriteString("] = ")
		//
		for i, rid := range p.Arguments {
			if i != 0 {
				builder.WriteString(", ")
			}
			//
			builder.WriteString(mapping.Register(rid).Name())
		}
	case MEMORY_READ:
		// data = mem[address]
		builder.WriteString(registersToString(mapping, array.Reverse(p.Returns)...))
		builder.WriteString(" = ")
		//
		fmt.Fprintf(&builder, "%s[", mapping.Module(p.Id).Name())
		//
		for i, rid := range p.Arguments {
			if i != 0 {
				builder.WriteString(", ")
			}
			//
			builder.WriteString(mapping.Register(rid).Name())
		}
		//
		builder.WriteString("]")
	default:
		// returns = ModuleName(arguments)
		if len(p.Returns) > 0 {
			builder.WriteString(registersToString(mapping, array.Reverse(p.Returns)...))
			builder.WriteString(" = ")
		}
		//
		fmt.Fprintf(&builder, "%s(", mapping.Module(p.Id).Name())
		//
		for i, rid := range p.Arguments {
			if i != 0 {
				builder.WriteString(", ")
			}
			//
			builder.WriteString(mapping.Register(rid).Name())
		}
		//
		builder.WriteString(")")
	}
	//
	return builder.String()
}

// MicroValidate implementation for MicroInstruction interface.
func (p *OpIo[W]) MicroValidate(_ uint, _ field.Config, _ SystemMap[W]) []error {
	return nil
}
