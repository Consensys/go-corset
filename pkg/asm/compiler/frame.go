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
package compiler

// Framing is used to manage additional registers required to ensure soundness.
// In particular, framinging applies to multi-line functions as these require a
// program counter, and various control lines to manage padding and non-terminal
// states.
type Framing[T any, E Expr[T, E]] interface {
	// Guard provides a suitable guard for the instruction at a given PC offset.
	// This is optional as some forms of framing don't require it.
	Guard(pc uint) E
	// Goto indicates the current instruction is jumping to the given PC value.
	Goto(pc uint) E
	// Return provides a suitable transition to the next frame.
	Return() E
}

// NewAtomicFraming constructs a suitable framing for a one-line instruction.
func NewAtomicFraming[T any, E Expr[T, E]]() Framing[T, E] {
	return &OneLineFraming[T, E]{}
}

// NewMultiLineFraming constructs a suitable framing for a multi-line instruction.
func NewMultiLineFraming[T any, E Expr[T, E]](pc T, ret T) Framing[T, E] {
	return &MultiLineFraming[T, E]{pc, ret}
}

// ============================================================================
// Atomic (i.e. One-Line) Framing
// ============================================================================

// OneLineFraming is suitable for one-line functions, as these require no
// control lines.
type OneLineFraming[T any, E Expr[T, E]] struct {
}

// Goto implementation for Framing interface.
func (p *OneLineFraming[T, E]) Goto(pc uint) E {
	panic("unreachable")
}

// Guard implementation for Framing interface.
func (p *OneLineFraming[T, E]) Guard(pc uint) E {
	return True[T, E]()
}

// Return implementation for Framing interface.
func (p *OneLineFraming[T, E]) Return() E {
	return True[T, E]()
}

// ============================================================================
// Multi-Line Framing
// ============================================================================

// MultiLineFraming provides suitable control lines for multi-line functions.
type MultiLineFraming[T any, E Expr[T, E]] struct {
	// Program Counter indicates which instruction is being executed.
	pc T
	// Return indicates when an instruction returns from the current function.
	// That is, the current frame is terminated.
	ret T
}

// Goto implementation for Framing interface.
func (p *MultiLineFraming[T, E]) Goto(pc uint) E {
	// PC[i+1] = target
	var (
		zero   = Number[T, E](0)
		pc_ip1 = Variable[T, E](p.pc, 1)
		ret    = Variable[T, E](p.ret, 0)
	)
	// Next pc is target of this jump
	eq := pc_ip1.Equals(Number[T, E](pc))
	// Return flag cannot be set
	return eq.And(ret.Equals(zero))
}

// Guard implementation for Framing interface.
func (p *MultiLineFraming[T, E]) Guard(pc uint) E {
	return Variable[T, E](p.pc, 0).Equals(Number[T, E](pc))
}

// Return implementation for Framing interface.
func (p *MultiLineFraming[T, E]) Return() E {
	var one = Number[T, E](1)
	// return line must be high; next PC must be zero.
	return Variable[T, E](p.ret, 0).Equals(one)
}
