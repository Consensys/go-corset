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

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Compile attempts to compile a given high-level program into a low-level
// machine which can be used (for example) to execute this program with some
// given inputs.
func Compile(p *ast.Program) machine.Boot[word.Uint] {
	var (
		functions []function.Boot[word.Uint]
		statics   []memory.Boot[word.Uint]
		inputs    []memory.Boot[word.Uint]
		outputs   []memory.Boot[word.Uint]
		rams      []memory.Boot[word.Uint]
	)
	// Initialise components
	for i, c := range p.Components() {
		switch c := c.(type) {
		case *ast.Constant:
			// ignore
		case *ast.Function:
			functions = append(functions, compileFunction(uint(i), *p))
		case *ast.Memory:
			// construct suitable decoder
			var decoder = memory.NewBootDecoder[word.Uint](c.Address, c.Data)
			//
			switch c.Kind {
			case decl.PRIVATE_READ_ONLY_MEMORY, decl.PUBLIC_READ_ONLY_MEMORY:
				inputs = append(inputs, memory.NewArray[word.Uint](c.Name(), decoder))
			case decl.PRIVATE_WRITE_ONCE_MEMORY, decl.PUBLIC_WRITE_ONCE_MEMORY:
				outputs = append(outputs, memory.NewArray[word.Uint](c.Name(), decoder))
			case decl.PRIVATE_STATIC_MEMORY, decl.PUBLIC_STATIC_MEMORY:
				statics = append(statics, memory.NewArray[word.Uint](c.Name(), decoder))
			case decl.RANDOM_ACCESS_MEMORY:
				rams = append(rams, memory.NewArray[word.Uint](c.Name(), decoder))
			}
		default:
			panic(fmt.Sprintf("unknown declaration %s", c.Name()))
		}
	}
	// Construct machine (if no errors)
	return machine.NewBoot[word.Uint]().
		WithFunctions(functions...).
		WithStatics(statics...).
		WithInputs(inputs...).
		WithOutputs(outputs...).
		WithMemories(rams...)
}
