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
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Code provides an abstract notion of an atomic "machine operation", where a
// single instruction is comprised of multiple such microcodes.  To ensure
// efficiency, we want to pack as many microcodes into each instruction as we
// can.  However, there are restrictions here meaning we cannot pack arbitrarily
// many microcodes into a single instruction.  For example, we cannot pack two
// microcodes together which have conflicting writes (i.e. both write to the
// same register).
type Code interface {
	// Clone this instruction
	Clone() Code
	// Execute a given micro-code, using a given local state.  This may update
	// the register values, and returns either the number of micro-codes to
	// "skip over" when executing the enclosing instruction or, if skip==0, a
	// destination program counter (which can signal return of enclosing
	// function).
	MicroExecute(state io.State) (skip uint, pc uint)
	// Registers returns the set of registers read this micro instruction.
	RegistersRead() []io.RegisterId
	// Registers returns the set of registers written by this micro instruction.
	RegistersWritten() []io.RegisterId
	// Produce a suitable string representation of this instruction.  This is
	// primarily used for debugging.
	String(register.Map) string
	// Split this micro code using registers of arbirary width into one or more
	// micro codes using registers of a fixed maximum width.
	Split(mapping register.LimbsMap, env agnostic.RegisterAllocator) []Code
	// Validate that this instruction is well-formed.  For example, that it is
	// balanced, that there are no conflicting writes, that all temporaries have
	// been allocated, etc.  The maximum bit capacity of the underlying field is
	// needed for this calculation, so as to allow an instruction to check it
	// does not overflow the underlying field.
	Validate(fieldWidth uint, fn register.Map) error
}

// Instruction represents the composition of one or more micro instructions
// which are to be executed "in parallel".  This roughly following the ideas of
// vector machines and vectorisation.  In order to ensure parallel execution is
// safe, there are restrictions on how microcodes can be combined.  For example,
// two microcodes writing to the same register are said to be "conflicting" and,
// hence, this is not permitted.  Likewise, it is not possible to branch into
// the middle of a microinstruction.
type Instruction struct {
	Codes []Code
}

// NewInstruction constructs a new instruction from a given set of micro-codes.
func NewInstruction(codes ...Code) Instruction {
	return Instruction{codes}
}

// Terminal checks whether or not this instruction can result in a return from
// the enclosing function.  That is, whether or not this instruction contains a
// "ret" micro-code.
func (p Instruction) Terminal() bool {
	for _, c := range p.Codes {
		if _, ok := c.(*Ret); ok {
			return true
		}
	}
	//
	return false
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p Instruction) Execute(state io.State) uint {
	var (
		skip uint = 1
		pc   uint
	)
	//
	for cc := uint(0); skip != 0; {
		// Decode next micro-code
		code := p.Codes[cc]
		// Execut micro-code
		skip, pc = code.MicroExecute(state)
		// Skip as requested
		cc += skip
	}
	//
	return pc
}

// JumpTargets returns the set of all jump targets used within this instruction.
// This is relatively easy to determine simply by looking for jmp codes.
func (p Instruction) JumpTargets() []uint {
	var targets []uint
	//
	for _, code := range p.Codes {
		if jmp, ok := code.(*Jmp); ok {
			targets = append(targets, jmp.Target)
		}
	}
	//
	return targets
}

// Registers returns the set of registers read/written by this instruction.
func (p Instruction) Registers() []io.RegisterId {
	return append(p.RegistersRead(), p.RegistersWritten()...)
}

// RegistersRead returns the set of registers read by this instruction.
func (p Instruction) RegistersRead() []io.RegisterId {
	var (
		regs bit.Set
		read []io.RegisterId
	)
	//
	for _, c := range p.Codes {
		for _, id := range c.RegistersRead() {
			if !regs.Contains(id.Unwrap()) {
				regs.Insert(id.Unwrap())
				read = append(read, id)
			}
		}
	}
	//
	return read
}

// RegistersWritten returns the set of registers written by this instruction.
func (p Instruction) RegistersWritten() []io.RegisterId {
	var (
		regs    bit.Set
		written []io.RegisterId
	)
	//
	for _, c := range p.Codes {
		for _, id := range c.RegistersWritten() {
			if !regs.Contains(id.Unwrap()) {
				regs.Insert(id.Unwrap())
				written = append(written, id)
			}
		}
	}
	//
	return written
}

// SplitRegisters implementation for the SplittableInstruction interface.  A key
// challenge for this method is the correct handling of skip instructions.
// Specifically, the targets for a skip change as the number of instructions
// increase.
func (p Instruction) SplitRegisters(mapping register.LimbsMap, env agnostic.RegisterAllocator) Instruction {
	var (
		ncodes  []Code
		packets [][]Code = make([][]Code, len(p.Codes))
		targets []uint   = make([]uint, len(p.Codes))
		index   uint
	)
	// Split micro-codes whilst retaining original indices.
	for i, code := range p.Codes {
		packets[i] = code.Split(mapping, env)
	}
	// Construct mapping
	for i := range targets {
		targets[i] = index
		index += uint(len(packets[i]))
	}
	// Finalise skip targets
	for i, packet := range packets {
		for j, c := range packet {
			c = retargetInsn(uint(i), uint(j), uint(len(packet)), c, targets)
			ncodes = append(ncodes, c)
		}
	}
	//
	return Instruction{Codes: ncodes}
}

func (p Instruction) String(fn register.Map) string {
	var builder strings.Builder
	//
	for i, code := range p.Codes {
		if i != 0 {
			builder.WriteString(" ; ")
		}
		//
		builder.WriteString(code.String(fn))
	}
	//
	return builder.String()
}

// Validate that this micro-instruction is well-formed.  For example, each
// micro-instruction contained within must be well-formed, and the overall
// requirements for a vector instruction must be met, etc.
func (p Instruction) Validate(fieldWidth uint, fn register.Map) error {
	var written bit.Set
	// Validate individual instructions
	for _, r := range p.Codes {
		if err := r.Validate(fieldWidth, fn); err != nil {
			return err
		}
	}
	//
	// TODO: check for unreachable instructions
	// TODO: check for conflicting function calls
	//
	// Check Write conflicts
	return validateWrites(0, written, p.Codes, fn)
}

func validateWrites(cc uint, writes bit.Set, codes []Code, fn register.Map) error {
	//
	switch code := codes[cc].(type) {
	case *Fail, *Ret, *Jmp:
		return nil
	case *Skip:
		if err := validateWrites(cc+code.Skip+1, writes.Clone(), codes, fn); err != nil {
			return err
		}
	default:
		//
		for _, dst := range code.RegistersWritten() {
			if writes.Contains(dst.Unwrap()) {
				// Extract register name
				name := fn.Register(dst).Name
				//
				return fmt.Errorf("conflicting write on register %s in %s", name, code.String(fn))
			}
			//
			writes.Insert(dst.Unwrap())
		}
	}
	// Fall through to next micro-code
	return validateWrites(cc+1, writes, codes, fn)
}

func retargetInsn(oldIndex uint, pktIndex, pktSize uint, code Code, mapping []uint) Code {
	var (
		newIndex     = mapping[oldIndex] + pktIndex
		leftInPacket = pktSize - pktIndex - 1
	)
	// First, check whether this is a skip instruction (or not) since only skip
	// instructions need to be retargeted.
	skip, ok := code.(*Skip)
	//
	if !ok {
		return code
	}
	// Determine true skip target
	target := oldIndex + 1 + (skip.Skip - leftInPacket)
	// Determine new location of skip target
	nTarget := mapping[target]
	//
	return &Skip{Left: skip.Left,
		Right: skip.Right,
		Skip:  nTarget - newIndex - 1,
	}
}
