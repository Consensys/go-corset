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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/asm/micro"
)

// Register is an alias for insn.Register
type Register = insn.Register

// Alias for big integer representation of 1.
var zero big.Int = *big.NewInt(0)

// Instruction provides an abstract notion of a macro "machine instruction".
// Here, macro is intended to imply that the instruction may break down into
// multiple underlying "micro instructions".
type Instruction interface {
	insn.Instruction
	// Bind any labels contained within this instruction using the given label map.
	Bind(labels []uint)
	// Lower this (macro) instruction into a sequence of one or more micro
	// instructions.
	Lower(pc uint) micro.Instruction
}
