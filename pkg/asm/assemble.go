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

	"github.com/consensys/go-corset/pkg/asm/assembler"
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source"
)

// Register describes a single register within a function.
type Register = io.Register

// MacroFunction is a function whose instructions are themselves macro
// instructions.  A macro function must be compiled down into a micro function
// before we can generate constraints.
type MacroFunction = io.Function[macro.Instruction]

// MicroFunction is a function whose instructions are themselves micro
// instructions.  A micro function represents the lowest representation of a
// function, where each instruction is made up of microcodes.
type MicroFunction = io.Function[micro.Instruction]

// MacroProgram represents a set of components at the macro level.
type MacroProgram = io.Program[macro.Instruction]

// MicroProgram represents a set of components at the micro level.
type MicroProgram = io.Program[micro.Instruction]

// MixedMacroProgram is a schema comprised of both macro assembly components and
// MIR components (hence the term mixed).
type MixedMacroProgram = schema.MixedSchema[*MacroFunction, mir.Module]

// MixedMicroProgram is a schema comprised of both micro assembly components and
// MIR components (hence the term mixed).
type MixedMicroProgram = schema.MixedSchema[*MicroFunction, mir.Module]

// Assemble takes a given set of assembly files, and parses them into a given
// set of functions.  This includes performing various checks on the files, such
// as type checking, etc.
func Assemble(assembly ...source.File) (
	MacroProgram, source.Maps[any], []source.SyntaxError) {
	//
	var (
		items   []assembler.AssemblyItem
		errors  []source.SyntaxError
		program MacroProgram
		srcmaps source.Maps[any]
	)
	// Parse each file in turn.
	for _, asm := range assembly {
		// Parse source file
		cs, errs := assembler.Parse(&asm)
		if len(errs) == 0 {
			items = append(items, cs)
		}
		//
		errors = append(errors, errs...)
	}
	// Link assembly
	if len(errors) != 0 {
		return program, srcmaps, errors
	}
	// Link assembly and resolve buses
	program, srcmaps = assembler.Link(items...)
	// Well-formedness checks (assuming unlimited field width).
	errors = assembler.Validate(math.MaxUint, program, srcmaps)
	// Done
	return program, srcmaps, errors
}

// LoweringConfig provides configuration options for configuring the lowering
// process.
type LoweringConfig struct {
	// Maximum number of bits the underlying field can hold.  This restricts the
	// combined bitwidth permitted for the target registers of an instruction.
	MaxFieldWidth uint
	// Maximum bitwidth permitted for registers.  This cannot be larger than the
	// maximum field width and, ideally, is somewhat smaller to accommodate
	// additions, etc.
	MaxRegisterWidth uint
	// Vectorize determines whether or not to enable vectorisation.  More
	// specifically, vectorisation attempts to combine multiple machine
	// instructions together into batches which can be "executed" concurrently,
	// roughly reminiscent of Very Long Instruction Word (VLIW) architectures.
	Vectorize bool
}

// LowerMixedMacroProgram a mixed macro program (i.e. schema) into a mixed micro program, using
// vectorisation if desired.  Specifically, any macro modules within the schema
// are lowered into "micro" modules (i.e. those using only micro instructions).
// This does not impact any externally defined (e.g. MIR) modules in the schema.
func LowerMixedMacroProgram(vectorize bool, p MixedMacroProgram) MixedMicroProgram {
	functions := make([]*MicroFunction, len(p.LeftModules()))
	//
	for i, f := range p.LeftModules() {
		nf := lowerFunction(vectorize, *f)
		functions[i] = &nf
	}
	// Construct program for validation
	program := io.NewProgram(functions...)
	// Validate generated program.  Whilst not strictly necessary, it is useful
	// from a debugging perspective.
	assembler.ValidateMicro(math.MaxUint, program)
	// Construct mixed micro schema
	return schema.NewMixedSchema(functions, p.RightModules())
}

// LowerMixedMicroProgram lowers a mixed micro program into a unform schema of
// MIR modules.  To do this, it translates all assembly components (e.g.
// functions) into MIR modules to ensure uniformity at the end.
func LowerMixedMicroProgram(p MixedMicroProgram) schema.UniformSchema[mir.Module] {
	var (
		n                    = len(p.LeftModules())
		modules []mir.Module = make([]mir.Module, p.Width())
		program              = io.NewProgram(p.LeftModules()...)
	)
	// Lower assembly components
	for i := range p.LeftModules() {
		modules[i] = compileFunction(uint(i), program)
	}
	// Copy of legacy components
	for i, m := range p.RightModules() {
		modules[i+n] = m
	}
	//
	return schema.NewUniformSchema(modules)
}

// ============================================================================
// Helpers
// ============================================================================

// Compiler a given micro function into an MIR module.
func compileFunction(index uint, program MicroProgram) mir.Module {
	// For now.
	panic("todo")
}

func lowerFunction(vectorize bool, f MacroFunction) MicroFunction {
	insns := make([]micro.Instruction, len(f.Code()))
	// Lower macro instructions to micro instructions.
	for pc, insn := range f.Code() {
		insns[pc] = insn.Lower(uint(pc))
	}
	// Sanity checks (for now)
	fn := io.NewFunction(f.Name(), f.Registers(), insns)
	// Apply vectorisation (if enabled).
	if vectorize {
		fn = vectorizeFunction(fn)
	}
	//
	return fn
}

// Impose requested bitwidth limits on registers and instructions, by splitting
// registers as necessary.  For example, suppose the maximum register width is
// set at 32bits.  Then, a 64bit register is split into two "limbs", each of
// which is 32bits wide.  Obviously, any register whose width is less than
// 32bits will not be split.  Instructions need to be split when the combined
// width of their target registers exceeds the maximum field width.  In such
// cases, carry flags are introduced to communicate overflow or underflow
// between the split instructions.
func splitRegisters(cfg LoweringConfig, f MicroFunction) MicroFunction {
	var (
		env = micro.NewRegisterSplittingEnvironment(cfg.MaxRegisterWidth, f.Registers())
		// Updated instruction sequence
		ninsns []micro.Instruction
		//
		ninsn micro.Instruction
	)
	// Split instructions
	for _, insn := range f.Code() {
		ninsn = splitMicroInstruction(insn, env)
		// Split instruction based on split registers
		ninsns = append(ninsns, ninsn)
	}
	// Done
	return io.NewFunction(f.Name(), env.RegistersAfter(), ninsns)
}

func splitMicroInstruction(insn micro.Instruction, env *micro.RegisterSplittingEnvironment) micro.Instruction {
	//
	var ncodes []micro.Code
	//
	for _, code := range insn.Codes {
		split := code.Split(env)
		ncodes = append(ncodes, split...)
	}
	//
	return micro.Instruction{Codes: ncodes}
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
	return io.NewFunction(f.Name(), f.Registers(), insns)
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
	var written []uint
	//
	switch code := codes[cc].(type) {
	case *micro.Add:
		written = code.RegistersWritten()
	case *micro.Jmp:
		if code.Target == target {
			// Prevent inlining multiple times.
			return conflictingInstructions(0, insn.Codes, writes, math.MaxUint, insn)
		}
		//
		return false
	case *micro.Mul:
		written = code.RegistersWritten()
	case *micro.Skip:
		// Check target location
		if conflictingInstructions(cc+1+code.Skip, codes, writes.Clone(), target, insn) {
			return true
		}
		// Fall through
	case *micro.Sub:
		written = code.RegistersWritten()
	case *micro.Ret:
		return false
	}
	// Check conflicts, and update mutated registers
	for _, r := range written {
		if writes.Contains(r) {
			return true
		}
		//
		writes.Insert(r)
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
		Left:     code.Left,
		Right:    code.Right,
		Constant: code.Constant,
		Skip:     target,
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
	// Start with entry block
	worklist.Push(0)
	//
	for !worklist.Empty() {
		pc := worklist.Pop()
		//
		worklist.PushAll(insns[pc].JumpTargets())
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
