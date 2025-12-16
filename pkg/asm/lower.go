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
package asm

import (
	"math"
	"slices"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
)

// LoweringConfig provides configuration options for configuring the lowering
// process.
type LoweringConfig struct {
	// Field determines necessary parameters for the underlying field.  This
	// includes: the maximum field bandwidth, which is number of bits the
	// underlying field can hold; and, the maximum register width.
	Field field.Config
	// Vectorize determines whether or not to enable vectorisation.  More
	// specifically, vectorisation attempts to combine multiple machine
	// instructions together into batches which can be "executed" concurrently,
	// roughly reminiscent of Very Long Instruction Word (VLIW) architectures.
	Vectorize bool
}

func lowerFunction(vectorize bool, f MacroFunction) MicroFunction {
	insns := make([]micro.Instruction, len(f.Code()))
	// Lower macro instructions to micro instructions.
	for pc, insn := range f.Code() {
		insns[pc] = insn.Lower(uint(pc))
	}
	// Sanity checks (for now)
	fn := io.NewFunction(f.Name(), f.IsPublic(), f.Registers(), f.Buses(), insns)
	// Apply vectorisation (if enabled).
	if vectorize {
		fn = vectorizeFunction(fn)
	}
	//
	return fn
}

// Vectorize a given function by merging as many instructions together as
// possible.  For example, consider two micro instructions "x = y" and "a = b".
// Since this instructions do not conflict over any assigned register, they can
// be combined into a vector instruction "x=y;a=b".
func vectorizeFunction(f MicroFunction) MicroFunction {
	var insns = slices.Clone(f.Code())
	// Vectorize instructions as much as possible.
	for pc := range insns {
		insns[pc] = vectorizeInstruction(uint(pc), f.Code())
	}
	// Remove all uncreachable instructions and compact remainder.
	insns = pruneUnreachableInstructions(insns)
	//
	return io.NewFunction(f.Name(), f.IsPublic(), f.Registers(), f.Buses(), insns)
}

func vectorizeInstruction(pc uint, insns []micro.Instruction) micro.Instruction {
	var (
		insn    = insns[pc]
		changed = true
	)
	// Keep vectorizing until worklist empty.
	for changed {
		changed = false
		//
		for _, target := range insn.JumpTargets() {
			targetInsn := insns[target]
			//
			if target != pc && !conflictingInstructions(0, insn.Codes, bit.Set{}, target, targetInsn) {
				insn = inlineInstruction(insn, target, insns[target])
				changed = true
			}
		}
	}
	//
	return insn
}

func conflictingInstructions(cc uint, codes []micro.Code, writes bit.Set, target uint, insn micro.Instruction) bool {
	// set of written registers
	var written []io.RegisterId
	//
	switch code := codes[cc].(type) {
	case *micro.Assign:
		written = code.RegistersWritten()
	case *micro.Division:
		written = code.RegistersWritten()
	case *micro.Jmp:
		if code.Target == target {
			// Prevent inlining multiple times.
			return conflictingInstructions(0, insn.Codes, writes, math.MaxUint, insn)
		}
		//
		return false
	case *micro.Skip:
		// Check target location
		if conflictingInstructions(cc+1+code.Skip, codes, writes.Clone(), target, insn) {
			return true
		}
		// Fall through
	case *micro.Ret, *micro.Fail:
		return false
	}
	// Check conflicts, and update mutated registers
	for _, r := range written {
		if writes.Contains(r.Unwrap()) {
			return true
		}
		//
		writes.Insert(r.Unwrap())
	}
	// Proceed
	return conflictingInstructions(cc+1, codes, writes, target, insn)
}

// Inline a given target instruction into existing instruction.  This means
// going through the existing instruction and replacing all jump's to the target
// address with the contents of the target instruction.  This is non-trivial as
// we must also correctly update internal code offsets for skip instructions,
// otherwise they could now skip over the wrong number of codes.
func inlineInstruction(insn micro.Instruction, target uint, targetInsn micro.Instruction) micro.Instruction {
	var (
		codes   = slices.Clone(insn.Codes)
		mapping = make([]uint, len(codes))
		npc     int
	)
	// First determine length of final sequence, and construct an appropriate
	// mapping from code offsets in the original instruction to those in the new
	// instruction.
	for cc := 0; cc < len(codes); cc, npc = cc+1, npc+1 {
		mapping[cc] = uint(npc)
		// Look for jump instruction
		if jmp, ok := codes[cc].(*micro.Jmp); ok {
			// Check whether its going to the right place
			if jmp.Target == target {
				// NOTE: -1 as will overwrite the jmp.
				npc += len(targetInsn.Codes) - 1
			}
		}
	}
	//
	ninsns := make([]micro.Code, npc)
	//
	for cc, npc := 0, 0; cc < len(codes); cc++ {
		code := codes[cc]
		//
		switch c := code.(type) {
		case *micro.Jmp:
			if c.Target == target {
				// copy over target instructions
				for _, c := range targetInsn.Codes {
					ninsns[npc] = c.Clone()
					npc++
				}
				//
				continue
			}
		case *micro.Skip:
			code = retargetSkip(uint(cc), uint(npc), *c, mapping)
		}
		//
		ninsns[npc] = code
		npc++
	}
	// Skip instructions may need updating here.
	return micro.Instruction{Codes: ninsns}
}

// Calculate the updated skip offset based on the mapping of old code offsets to
// new code offsets.
func retargetSkip(cc uint, npc uint, code micro.Skip, mapping []uint) micro.Code {
	// Determine absolute target
	target := mapping[cc+1+code.Skip]
	// Recalculate as relative offset
	target = target - npc - 1
	//
	return &micro.Skip{
		Left:  code.Left,
		Right: code.Right,
		Skip:  target,
	}
}

// Identify and remove all unreachable instructions.  A tricky aspect of this is
// that we must updating jump targets accordingly.
func pruneUnreachableInstructions(insns []micro.Instruction) []micro.Instruction {
	var (
		reachable bit.Set = determineReachableInstructions(insns)
		ninsns    []micro.Instruction
		mapping   []uint = make([]uint, len(insns))
	)
	// Remove all unreachable
	for i, insn := range insns {
		if reachable.Contains(uint(i)) {
			mapping[i] = uint(len(ninsns))
			ninsns = append(ninsns, insn)
		}
	}
	// Rebinding all existing jump targets
	for _, insn := range ninsns {
		for i := range insn.Codes {
			code := insn.Codes[i]
			if jmp, ok := code.(*micro.Jmp); ok {
				jmp.Target = mapping[jmp.Target]
				insn.Codes[i] = jmp
			}
		}
	}
	//
	return ninsns
}

func determineReachableInstructions(insns []micro.Instruction) bit.Set {
	var (
		worklist = VecWorklist{}
	)
	//
	if len(insns) > 0 {
		// Start with entry block
		worklist.Push(0)
		//
		for !worklist.Empty() {
			pc := worklist.Pop()
			//
			worklist.PushAll(insns[pc].JumpTargets())
		}
	}
	// Done
	return worklist.visited
}

// VecWorklist is a worklist suitable for the vectorisation algorithm.
type VecWorklist struct {
	// Set of target pc locations yet to be explored
	targets []uint
	// Set of target pc locations already explored.
	visited bit.Set
}

// Empty determines whether or not the worklist is empty.
func (p *VecWorklist) Empty() bool {
	return len(p.targets) == 0
}

// Pop returns the next item off the worklist.
func (p *VecWorklist) Pop() uint {
	n := len(p.targets) - 1
	item := p.targets[n]
	p.targets = p.targets[:n]
	//
	return item
}

// Push pushes a new target onto the worklist, provided it has not been
// previously visited.
func (p *VecWorklist) Push(target uint) {
	if !p.visited.Contains(target) {
		p.visited.Insert(target)
		p.targets = append(p.targets, target)
	}
}

// PushAll attempts to push all targets onto the worklist, whilst excluding
// those which have been visited already.
func (p *VecWorklist) PushAll(targets []uint) {
	for _, target := range targets {
		p.Push(target)
	}
}
