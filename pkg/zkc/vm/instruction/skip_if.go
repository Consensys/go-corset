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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
)

const (
	// EQ indicates an equality condition
	EQ Condition = 0
	// NEQ indicates a non-equality condition
	NEQ Condition = 1
	// LT indicates a less-than condition
	LT Condition = 2
	// GT indicates a greater-than condition
	GT Condition = 3
	// LTEQ indicates a less-than-or-equals condition
	LTEQ Condition = 4
	// GTEQ indicates a greater-than-or-equals condition
	GTEQ Condition = 5
)

// Condition represents the set of permission comparitors for a SkipIf
// instruction.
type Condition uint

// SkipIf microcode performs a conditional skip over a given number of codes. The
// condition is either that two registers are equal, or that they are not equal.
// This has two variants: register-register; and, register-constant.  The latter
// is indiciated when the right register is marked as UNUSED.
type SkipIf struct {
	Cond Condition
	// Left and right comparisons
	Left register.Vector
	//
	Right register.Vector
	// Skip
	Skip uint
}

// NewSkipIf constructs a fresh conditional skip instruction.
func NewSkipIf(condition Condition, left, right register.Id, skip uint) *SkipIf {
	return &SkipIf{condition, register.NewVector(left), register.NewVector(right), skip}
}

// Uses implementation for Instruction interface
func (p *SkipIf) Uses() []register.Id {
	var regs []io.RegisterId
	// Add all registers on the left-hand side
	regs = append(regs, p.Left.Registers()...)
	// Add all registers on the right-hand side (if applicable)
	regs = append(regs, p.Right.Registers()...)
	//
	return regs
}

// Definitions implementation for Instruction interface
func (p *SkipIf) Definitions() []io.RegisterId {
	return nil
}

func (p *SkipIf) String(fn register.Map) string {
	var (
		l = p.Left.String(fn)
		r = p.Right.String(fn)
		o string
	)
	//
	switch p.Cond {
	case EQ:
		o = "=="
	case NEQ:
		o = "!="
	case LT:
		o = "<"
	case LTEQ:
		o = "<="
	case GT:
		o = ">"
	case GTEQ:
		o = ">="
	default:
		panic("unknown skip condition encountered")
	}
	//
	return fmt.Sprintf("skip %s %s %s %d", l, o, r, p.Skip)
}

// MicroValidate iumplementation for MicroInstruction interface
func (p *SkipIf) MicroValidate(n uint, _ field.Config, fn register.Map) []error {
	var (
		errors []error
		lw     = p.Left.BitWidth(fn)
		rw     = p.Right.BitWidth(fn)
	)
	//
	if lw < rw {
		errors = append(errors, fmt.Errorf("bit overflow (u%d into u%d)", lw, rw))
	}
	//
	return errors
}
