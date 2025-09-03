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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
)

// InferPadding attempts to infer suitable padding values for a function, based
// on those padding values provided for its inputs (which default to 0).  In
// essence, this constructs a witness for the function in question.
func InferPadding[F field.Element[F], T io.Instruction[T]](fn io.Function[F, T], executor *Executor[F, T]) {
	//
	if fn.IsAtomic() {
		// Only infer padding for one-line functions.
		var (
			insn      = fn.CodeAt(0)
			registers = fn.Registers()
			state     = initialState(registers, executor)
		)
		// Execute the one instruction
		_ = insn.Execute(state)
		// Assign padding values
		for i := range registers {
			var (
				val big.Int
				rid = schema.NewRegisterId(uint(i))
			)
			// Load ith register value
			val.Set(state.Load(rid))
			// Update padding value
			registers[i].Padding = val
		}
	}
}

// Construct initial state from the given padding values.
func initialState(registers []Register, iomap io.Map) io.State {
	var (
		state = make([]big.Int, len(registers))
		index = 0
	)
	// Initialise arguments
	for i, reg := range registers {
		if reg.IsInput() {
			var ith big.Int
			// Clone big int.
			ith.SetBytes(reg.Padding.Bytes())
			// Assign to ith register
			state[i] = ith
			index = index + 1
		}
	}
	//
	return io.InitialState(state, registers, iomap)
}
