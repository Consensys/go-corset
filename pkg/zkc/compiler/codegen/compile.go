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
	"math"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Declaration represents a declaration which can contain macro
// instructions and where external identifiers are otherwise resolved. As such,
// it should not be possible that such a declaration refers to unknown (or
// otherwise incorrect) external components.
type Declaration = decl.Declaration[symbol.Resolved]

// VariableDescriptor represents a descriptor whose external identifiers are
// otherwise resolved. As such, it should not be possible that such a
// declaration refers to unknown (or otherwise incorrect) external components.
type VariableDescriptor = variable.Descriptor[symbol.Resolved]

// Compile attempts to compile a given high-level program into a low-level
// machine which can be used (for example) to execute this program with some
// given inputs.
func Compile(env data.ResolvedEnvironment, declarations []Declaration, srcmaps source.Maps[any],
) (*machine.Base[word.Uint], []source.SyntaxError) {
	//
	var (
		modules []machine.Module[word.Uint]
		mapping = make([]uint, len(declarations))
		index   = uint(0)
		errors  []source.SyntaxError
	)
	// Construct the mapping from ast declaration identifiers to vm module
	// identifiers.  Essentially, what is happening here is that some ast
	// declarations will no longer exist at the machine level.  So, when a
	// declaration is encountered that will no longer exist, then the id for all
	// declarations after it is shifted down.
	for i, d := range declarations {
		switch d.(type) {
		case *decl.ResolvedFunction, *decl.ResolvedMemory:
			mapping[i] = index
			index++
		default:
			mapping[i] = math.MaxUint
		}
	}
	// Initialise components
	for i, c := range declarations {
		switch c := c.(type) {
		case *decl.ResolvedConstant:
			// force detection of errors
			_, errs := compileStaticInitialisers(declarations, env, srcmaps, c.ConstExpr)
			//
			errors = append(errors, errs...)
		case *decl.ResolvedTypeAlias:
			// ignore
		case *decl.ResolvedFunction:
			fn, errs := compileFunction(uint(i), mapping, declarations, srcmaps, env)
			modules = append(modules, fn)
			errors = append(errors, errs...)
		case *decl.ResolvedInclude:
			// ignore
		case *decl.ResolvedMemory:
			var regs = toMemoryRegisters(c.Address, c.Data, env)
			//
			switch c.Kind {
			case decl.PRIVATE_READ_ONLY_MEMORY, decl.PUBLIC_READ_ONLY_MEMORY:
				modules = append(modules, memory.NewReadOnly[word.Uint](c.Name(), regs))
			case decl.PRIVATE_WRITE_ONCE_MEMORY, decl.PUBLIC_WRITE_ONCE_MEMORY:
				modules = append(modules, memory.NewWriteOnce[word.Uint](c.Name(), regs))
			case decl.PRIVATE_STATIC_MEMORY, decl.PUBLIC_STATIC_MEMORY:
				// Compile the static initialiser
				words, errs := compileStaticInitialisers(declarations, env, srcmaps, c.Contents...)
				//
				if len(errs) == 0 {
					// Construct the read-only memory
					modules = append(modules, memory.NewStaticReadOnly(c.Name(), regs, words...))
				}
				// Include all errors
				errors = append(errors, errs...)
			case decl.RANDOM_ACCESS_MEMORY:
				modules = append(modules, memory.NewRandomAccess[word.Uint](c.Name(), regs))
			}
		default:
			panic(fmt.Sprintf("unknown declaration %s", c.Name()))
		}
	}
	// Construct machine (if no errors)
	return machine.New(modules...), errors
}

// compileStaticInitialise evaluates the compile-time constant expressions from a static
// memory declaration into the word.Uint representation required by the VM.
func compileStaticInitialisers(components []Declaration, env data.ResolvedEnvironment,
	srcmaps source.Maps[any], contents ...expr.Resolved) ([]word.Uint, []source.SyntaxError) {
	//
	var (
		words    = make([]word.Uint, len(contents))
		compiler = Compiler{components, nil, nil, env, srcmaps, nil}
	)
	//
	for i, v := range contents {
		words[i] = compiler.evalConstant(v, true)
	}

	return words, compiler.errors
}

func toMemoryRegisters(address []VariableDescriptor, datas []VariableDescriptor, env data.ResolvedEnvironment,
) []register.Register {
	var (
		registers []register.Register
		padding   big.Int
	)
	// Flattern address lines
	for _, v := range address {
		data.Flattern(v.DataType, v.Name, env, func(name string, bitwidth uint) {
			registers = append(registers, register.NewInput(name, bitwidth, padding))
		})
	}
	// Flattern data lines
	for _, v := range datas {
		data.Flattern(v.DataType, v.Name, env, func(name string, bitwidth uint) {
			registers = append(registers, register.NewOutput(name, bitwidth, padding))
		})
	}
	//
	return registers
}
