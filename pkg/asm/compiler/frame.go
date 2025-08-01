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
	// IsAtomic indicates a "one line function".  In such case, no framing is
	// required.
	IsAtomic() bool
	// Guard provides a suitable guard for the instruction at a given PC offset.
	// This is optional as some forms of framing don't require it.
	Guard(pc uint) E
	// Terminate provides a suitable transition to the next frame.
	Terminate() E
	// ProgramCounter returns the identifier for the PC register.  Observe, however,
	// that this will not exist for atomic functions and, in such case, this will
	// panic.
	ProgramCounter() T
}

// NewAtomicFraming constructs a suitable framing for a one-line instruction.
func NewAtomicFraming[T any, E Expr[T, E]]() Framing[T, E] {
	return &OneLineFraming[T, E]{}
}

// NewMultiLineFraming constructs a suitable framing for a multi-line instruction.
func NewMultiLineFraming[T any, E Expr[T, E]](pc T, terminal T, enable T) Framing[T, E] {
	panic("got here")
}

// ============================================================================
// Atomic (i.e. One-Line) Framing
// ============================================================================

// OneLineFraming is suitable for one-line functions, as these require no
// control lines.
type OneLineFraming[T any, E Expr[T, E]] struct {
}

// IsAtomic implementation for Framining interface.
func (p *OneLineFraming[T, E]) IsAtomic() bool {
	return true
}

// ProgramCounter implementation for Framining interface.
func (p *OneLineFraming[T, E]) ProgramCounter() T {
	panic("atomic functions have no PC")
}

// Guard implementation for Framining interface.
func (p *OneLineFraming[T, E]) Guard(pc uint) E {
	return True[T, E]()
}

// Terminate implementation for Framining interface.
func (p *OneLineFraming[T, E]) Terminate() E {
	return True[T, E]()
}

// ============================================================================
// Multi-Line Framing
// ============================================================================

// MultiLineFraming provides suitable control lines for multi-line functions.
type MultiLineFraming[T any, E Expr[T, E]] struct {
	pc T
}

// IsAtomic implementation for Framining interface.
func (p *MultiLineFraming[T, E]) IsAtomic() bool {
	return false
}

// ProgramCounter implementation for Framining interface.
func (p *MultiLineFraming[T, E]) ProgramCounter() T {
	return p.pc
}

// Guard implementation for Framining interface.
func (p *MultiLineFraming[T, E]) Guard(pc uint) E {
	return Variable[T, E](p.pc, 0).Equals(Number[T, E](pc))
}

// Terminate implementation for Framining interface.
func (p *MultiLineFraming[T, E]) Terminate() E {
	// var (
	// 	stamp_i   = p.Stamp(false)
	// 	stamp_ip1 = p.Stamp(true)
	// 	one       = Number[T, E](1)
	// )
	// // STAMP[i]+1 == STAMP[i+1]
	// eqn := one.Add(stamp_i).Equals(stamp_ip1)
	panic("todo")
}
