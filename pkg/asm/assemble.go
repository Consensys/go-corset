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
	"fmt"
	"math"
	"slices"

	"github.com/consensys/go-corset/pkg/asm/assembler"
	"github.com/consensys/go-corset/pkg/asm/compiler"
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/binfile"
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

// Assemble takes a given set of assembly files, and parses them into a given
// set of functions.  This includes performing various checks on the files, such
// as type checking, etc.
func Assemble(assembly ...source.File) (io.Program[macro.Instruction], source.Maps[macro.Instruction], []source.SyntaxError) {
	var (
		components []MacroFunction
		errors     []source.SyntaxError
		srcmaps    source.Maps[macro.Instruction] = *source.NewSourceMaps[macro.Instruction]()
	)
	// Parse each file in turn.
	for _, asm := range assembly {
		// Parse source file
		cs, srcmap, errs := assembler.Parse(&asm)
		if len(errs) == 0 {
			components = append(components, cs...)
		}
		// Join srcmap
		srcmaps.Join(srcmap)
		//
		errors = append(errors, errs...)
	}
	// Well-formedness checks
	for _, fn := range components {
		errors = append(errors, assembler.Validate(fn, srcmaps)...)
	}
	// Done
	return io.NewProgram(components...), srcmaps, errors
}

// CompileAssembly compiles a given set of assembly functions into a binary
// constraint file.
func CompileAssembly(cfg LoweringConfig, assembly ...source.File) (*binfile.BinaryFile, []source.SyntaxError) {
	macroProgram, _, errs := Assemble(assembly...)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Lower macro program into a binary program.
	microProgram := Lower(cfg, macroProgram)
	//
	return Compile(microProgram), nil
}

// Compile a microprogram into a binary constraint file.
func Compile(microProgram io.Program[micro.Instruction]) *binfile.BinaryFile {
	compiler := compiler.NewCompiler()
	//
	for i := range microProgram.Functions() {
		fn := microProgram.Function(uint(i))
		compiler.Compile(fn.Name, fn.Registers, fn.Code)
	}

	return binfile.NewBinaryFile(nil, nil, compiler.Schema())
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

// Lower a given macro program into a micro program which only uses registers of
// a given width.  This is a relatively involved procress consisting of several
// steps: firstly, all macro instructions are lowered to micro instructions;
// secondly, vectorization is applied to the resulting microprogram; finally,
// registers exceeding the target width (and instructions which use them) are
// split accordingly.  The latter can introduce additional registers, for
// example to hold carry values.
func Lower(cfg LoweringConfig, p MacroProgram) MicroProgram {
	functions := make([]MicroFunction, len(p.Functions()))
	// Sanity checks
	if cfg.MaxFieldWidth < cfg.MaxRegisterWidth {
		panic(
			fmt.Sprintf("field width (%dbits) smaller than register width (%dbits)", cfg.MaxFieldWidth, cfg.MaxRegisterWidth))
	}
	//
	for i, f := range p.Functions() {
		functions[i] = lowerFunction(cfg, f)
	}
	//
	return io.NewProgram(functions...)
}

// ============================================================================
// Helpers
// ============================================================================

func lowerFunction(cfg LoweringConfig, f MacroFunction) MicroFunction {
	insns := make([]micro.Instruction, len(f.Code))
	// Lower macro instructions to micro instructions.
	for pc, insn := range f.Code {
		insns[pc] = insn.Lower(uint(pc))
	}
	// Sanity checks (for now)
	fn := MicroFunction{f.Name, f.Registers, insns}
	// Split registers as necessary to meet limits.
	fn = splitRegisters(cfg, fn)
	// Apply vectorisation (if enabled).
	if cfg.Vectorize {
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
		env = micro.NewRegisterSplittingEnvironment(cfg.MaxRegisterWidth, f.Registers)
		// Updated instruction sequence
		ninsns []micro.Instruction
		//
		ninsn micro.Instruction
	)
	// Split instructions
	for _, insn := range f.Code {
		ninsn = splitMicroInstruction(insn, cfg, env)
		// Split instruction based on split registers
		ninsns = append(ninsns, ninsn)
	}
	// Done
	return MicroFunction{f.Name, env.RegistersAfter(), ninsns}
}

func splitMicroInstruction(insn micro.Instruction, cfg LoweringConfig,
	env *micro.RegisterSplittingEnvironment) micro.Instruction {
	//
	var ncodes []micro.Code
	//
	for _, code := range insn.Codes {
		split := code.Split(env)
		ncodes = append(ncodes, split...)
	}
	// Sanity check split codes are valid.  This is not strictly necessary, but
	// is useful for debugging.
	for _, code := range ncodes {
		if err := code.Validate(cfg.MaxFieldWidth, env.RegistersAfter()); err != nil {
			panic(err.Error())
		}
	}
	//
	return micro.Instruction{Codes: ncodes}
}

// Vectorize a given function by merging as many instructions together as
// possible.  For example, consider two micro instructions "x = y" and "a = b".
// Since this instructions do not conflict over any assigned register, they can
// be combined into a vector instruction "x=y;a=b".
func vectorizeFunction(f MicroFunction) MicroFunction {
	var insns = slices.Clone(f.Code)
	// Vectorize instructions as much as possible.
	for pc := range insns {
		insns[pc] = vectorizeInstruction(uint(pc), f.Code)
	}
	// Remove all uncreachable instructions and compact remainder.
	insns = pruneUnreachableInstructions(insns)
	//
	return MicroFunction{Name: f.Name, Registers: f.Registers, Code: insns}
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
