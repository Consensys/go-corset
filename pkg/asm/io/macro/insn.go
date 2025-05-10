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
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
)

// Alias for big integer representation of 0.
var zero big.Int = *big.NewInt(0)

// Alias for big integer representation of 1.
var one big.Int = *big.NewInt(1)

// Register is an alias for insn.Register
type Register = io.Register

// Instruction provides an abstract notion of a macro "machine instruction".
// Here, macro is intended to imply that the instruction may break down into
// multiple underlying "micro instructions".
type Instruction interface {
	io.Instruction[Instruction]
	// Bind any labels contained within this instruction using the given label map.
	Bind(labels []uint)
	// Link any buses used within this instruction using the given bus map.
	Link(buses []uint)
	// Lower this (macro) instruction into a sequence of one or more micro
	// instructions.
	Lower(pc uint) micro.Instruction
}

func assignmentToString(dsts []uint, srcs []uint, constant big.Int, regs []io.Register, c big.Int, op string) string {
	var builder strings.Builder
	//
	builder.WriteString(io.RegistersReversedToString(dsts, regs))
	builder.WriteString(" = ")
	//
	for i, r := range srcs {
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
