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
package macro

import (
	"encoding/gob"
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
)

// Alias for big integer representation of 0.
var zero big.Int = *big.NewInt(0)

// Alias for big integer representation of 1.
var one big.Int = *big.NewInt(1)

// Alias for big integer representation of -1.
var minusOne big.Int = *big.NewInt(-1)

// Register is an alias for insn.Register
type Register = io.Register

// Instruction provides an abstract notion of a macro "machine instruction".
// Here, macro is intended to imply that the instruction may break down into
// multiple underlying "micro instructions".
type Instruction interface {
	io.Instruction[Instruction]
	// Lower this (macro) instruction into a sequence of one or more micro
	// instructions.
	Lower(pc uint) micro.Instruction
}

// BranchInstruction captures those instructions which may branch to some
// location.
type BranchInstruction interface {
	// Bind any labels contained within this instruction using the given label map.
	Bind(labels []uint)
}

// IoInstruction provides an abstraction notion of a macro instruction which
// uses a bus (e.g. to implement a function call).
type IoInstruction interface {
	io.InOutInstruction
	// Link links the bus.  Observe that this can only be called once on any
	// given instruction.
	Link(bus io.Bus)
}

func assignmentToString(dsts []io.RegisterId, srcs []io.RegisterId, constant big.Int, fn schema.Module,
	c big.Int, op string) string {
	//
	var (
		builder strings.Builder
		regs    = fn.Registers()
	)
	//
	builder.WriteString(io.RegistersReversedToString(dsts, regs))
	builder.WriteString(" = ")
	//
	for i, id := range srcs {
		r := id.Unwrap()
		//
		if i != 0 {
			builder.WriteString(op)
		}
		//
		if r < uint(len(regs)) {
			builder.WriteString(regs[r].Name)
		} else {
			builder.WriteString(fmt.Sprintf("?%d", r))
		}
	}
	//
	if len(srcs) == 0 || constant.Cmp(&c) != 0 {
		if len(srcs) > 0 {
			builder.WriteString(op)
		}
		//
		builder.WriteString("0x")
		builder.WriteString(constant.Text(16))
	}
	//
	return builder.String()
}

func init() {
	gob.Register(Instruction(&Add{}))
	gob.Register(Instruction(&Call{}))
	gob.Register(Instruction(&Goto{}))
	gob.Register(Instruction(&IfGoto{}))
	gob.Register(Instruction(&Mul{}))
	gob.Register(Instruction(&Return{}))
	gob.Register(Instruction(&Sub{}))
}
