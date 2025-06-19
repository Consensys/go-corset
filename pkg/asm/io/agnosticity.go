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
package io

import "github.com/consensys/go-corset/pkg/asm/io/agnosticity"

// SplittingEnvironment is used to assist with register splitting.
type SplittingEnvironment interface {
	// MaxWidth returns the maximum permitted register width.
	MaxWidth() uint
	// AllocateCarryRegister allocates a carry flag to hold bits which "overflow" the
	// left-hand side of an assignment (i.e. where sourceWidth is greater than
	// targetWidth).
	AllocateCarryRegister(targetWidth uint, sourceWidth uint) RegisterId
	// AllocateTargetLimbs allocates upto maxWidth bits from a given set of target
	// limbs.
	AllocateTargetLimbs(targetLimbs []RegisterId) (uint, []RegisterId, []RegisterId)
	// RegistersAfter returns the set of registers as they appear after splitting.
	RegistersAfter() []Register
	// RegistersBefore returns the set of registers as they appear after splitting.
	RegistersBefore() []Register
	// SplitSourceRegisters splits a given set of source registers into "packets" of
	// limbs.  For example, suppose r0 and r1 are source registers of bitwidth
	// (respectively) 16bits and 8bits.  Then, splitting for a maximum width of 8
	// yields 2 packets: {{r0'0,r1'0}, {r0'1}}
	SplitSourceRegisters(sources ...RegisterId) [][]RegisterId
	// SplitTargetRegisters splits a set of registers, e.g. for an assignment.  For
	// example, suppose we have:
	//
	// > b,x,y = ...
	//
	// Where x,y are 16bit registers and b is a 1bit overflow.  For a maximum
	// register width of 8bits, the above is transformed into:
	//
	// > b,x'1,x'0',y'1,y'0 = ...
	//
	// And this set of expanded target registers is returned.
	SplitTargetRegisters(targets ...RegisterId) []RegisterId
}

// SplitRegisters imposes requested bitwidth limits on registers and
// instructions, by splitting registers as necessary.  For example, suppose the
// maximum register width is set at 32bits.  Then, a 64bit register is split
// into two "limbs", each of which is 32bits wide.  Obviously, any register
// whose width is less than 32bits will not be split.  Instructions need to be
// split when the combined width of their target registers exceeds the maximum
// field width.  In such cases, carry flags are introduced to communicate
// overflow or underflow between the split instructions.
func SplitRegisters[T Instruction[T]](maxFieldWidth, maxRegisterWidth uint, f *Function[T]) *Function[T] {
	var (
		env = agnosticity.NewRegisterSplittingEnvironment(maxRegisterWidth, f.Registers())
		// Updated instruction sequence
		ninsns []T
	)
	// Split instructions
	for _, insn := range f.Code() {
		var ith Instruction[T] = insn
		//nolint
		if i, ok := ith.(SplittableInstruction[T]); ok {
			ninsns = append(ninsns, i.SplitRegisters(env))
		} else {
			panic("non-field agnostic instruction encountered")
		}
	}
	// Done
	nf := NewFunction(f.Id(), f.Name(), env.RegistersAfter(), ninsns)
	//
	return &nf
}
