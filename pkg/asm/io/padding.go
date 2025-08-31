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

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/util/field"
)

// InferPadding attempts to infer suitable padding values for a function, based
// on those padding values provided for its inputs (which default to 0).  In
// essence, this constructs a witness for the function in question.
func InferPadding[F field.Element[F], T Instruction[T]](fns []*Function[F, T]) {
	for i := range fns {
		inferPaddingForFunction(i, fns)
	}
}

func inferPaddingForFunction[F field.Element[F], T Instruction[T]](i int, fns []*Function[F, T]) {
	var fn = fns[i]
	//
	if fn.IsAtomic() {
		// Only infer padding for one-line functions.
		var (
			insn  = fn.code[0]
			state = initialState(fn.registers, fns)
		)
		// Execute the one instruction
		_ = insn.Execute(state)
		// Assign padding values
		for i := range fn.registers {
			fn.registers[i].Padding = state.state[i]
		}
	}
}

func initialState(registers []Register, io Map) State {
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
	return State{0, false, state, registers, io}
}
