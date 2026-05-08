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
package base

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
)

// SkipIf microcode performs a conditional skip over a given number of codes. The
// condition is either that two registers are equal, or that they are not equal.
// This has two variants: register-register; and, register-constant.  The latter
// is indiciated when the right register is marked as UNUSED.
type SkipIf struct {
	Cond opcode.Condition
	// Left and right comparisons
	Left register.Id
	//
	Right register.Id
	// Skip
	Skip uint
}

// IsWord implementation for instruction.Word interface
func (p *SkipIf) IsWord() bool {
	return true
}

// IsField implementation for instruction.Field interface
func (p *SkipIf) IsField() bool {
	return true
}

// OpCode implementation for Instruction interface
func (p *SkipIf) OpCode() opcode.OpCode {
	return opcode.SKIP_IF
}

// Uses implementation for Instruction interface
func (p *SkipIf) Uses() []register.Id {
	var regs []io.RegisterId
	// Add all registers on the left-hand side
	regs = append(regs, p.Left)
	// Add all registers on the right-hand side (if applicable)
	regs = append(regs, p.Right)
	//
	return regs
}

// Definitions implementation for Instruction interface
func (p *SkipIf) Definitions() []io.RegisterId {
	return nil
}

func (p *SkipIf) String(mapping SystemMap) string {
	var (
		l = mapping.Register(p.Left).Name()
		r = mapping.Register(p.Right).Name()
		o string
	)
	//
	switch p.Cond {
	case opcode.EQ:
		o = "=="
	case opcode.NEQ:
		o = "!="
	case opcode.LT:
		o = "<"
	case opcode.LTEQ:
		o = "<="
	case opcode.GT:
		o = ">"
	case opcode.GTEQ:
		o = ">="
	default:
		panic("unknown skip condition encountered")
	}
	//
	return fmt.Sprintf("skip_if %s %s %s %d", l, o, r, p.Skip)
}

// MicroValidate iumplementation for MicroInstruction interface
func (p *SkipIf) MicroValidate(n uint, _ field.Config, fn SystemMap) []error {
	var (
		errors []error
		lw     = fn.Register(p.Left).Width()
		rw     = fn.Register(p.Right).Width()
	)
	//
	if lw < rw {
		errors = append(errors, fmt.Errorf("bit overflow (u%d into u%d)", lw, rw))
	}
	//
	return errors
}
