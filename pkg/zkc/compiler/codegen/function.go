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
package codegen

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Instruction provides a convenient alias
type Instruction = instruction.Instruction[word.Uint]

// MicroInstruction provides a convenient alias
type MicroInstruction = instruction.MicroInstruction[word.Uint]

// Convert a decl.Function instance into a fun.Function instance by flattening
// the variable descriptors into register descriptors.  Each variable may
// expand into one or more registers (e.g. a tuple variable produces one
// register per element).
func compileFunction(id uint, mapping []uint, program []Declaration) *function.Boot[word.Uint] {
	var (
		fn        = program[id].(*Function)
		registers []register.Register
		padding   big.Int // zero padding
		bootCode  = make([]instruction.Instruction[word.Uint], len(fn.Code))
	)
	//
	for _, v := range fn.Variables {
		var kind register.Type

		switch v.Kind {
		case variable.PARAMETER:
			kind = register.INPUT_REGISTER
		case variable.RETURN:
			kind = register.OUTPUT_REGISTER
		case variable.LOCAL:
			kind = register.COMPUTED_REGISTER
		default:
			panic(fmt.Sprintf("unexpected variable kind %d", v.Kind))
		}

		v.DataType.Flattern(v.Name, func(name string, bitwidth uint) {
			registers = append(registers, register.New(kind, name, bitwidth, padding))
		})
	}
	//
	compiler := Compiler{program, fn.Variables, registers}
	//
	for i, stmt := range fn.Code {
		bootCode[i] = compiler.compileStatement(uint(i), mapping, stmt)
	}
	//
	return function.New[Instruction](fn.Name(), compiler.registers, bootCode)
}
