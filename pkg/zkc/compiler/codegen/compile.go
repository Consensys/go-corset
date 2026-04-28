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
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
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

// Compiler is responsible for compiling high-level programs into low-level
// machines which can be used (for example) to execute this program with some
// given inputs.  A compile is configurable in certain aspects.
type Compiler struct {
	env     data.ResolvedEnvironment
	srcmaps source.Maps[any]
	// configuration
	config Config
}

// NewCompiler constructs a code generator parameterised by a configuration,
// the resolved type environment, and the source maps recorded by earlier
// pipeline stages.  The configuration controls optional passes such as
// vectorisation; cfg=DEFAULT_CONFIG matches the prover-facing defaults.  The
// environment supplies type information needed when lowering expressions
// (e.g. bit-widths of named types), and the source maps allow generated
// instructions and any errors raised during compilation to be tied back to
// their originating source positions.
func NewCompiler(cfg Config, env data.ResolvedEnvironment, srcmaps source.Maps[any]) *Compiler {
	return &Compiler{
		env:     env,
		srcmaps: srcmaps,
		config:  cfg,
	}
}

// Compile attempts to compile a given high-level program into a low-level
// machine which can be used (for example) to execute this program with some
// given inputs.
func (p *Compiler) Compile(declarations []Declaration) (*machine.Base[word.Uint], []source.SyntaxError) {
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
			_, errs := p.compileStaticInitialisers(declarations, p.env, p.srcmaps, c.ConstExpr)
			//
			errors = append(errors, errs...)
		case *decl.ResolvedTypeAlias:
			// ignore
		case *decl.ResolvedFunction:
			fn, errs := p.compileFunction(uint(i), mapping, declarations)
			modules = append(modules, fn)
			errors = append(errors, errs...)
		case *decl.ResolvedInclude:
			// ignore
		case *decl.ResolvedMemory:
			var regs = toMemoryRegisters(c.Address, c.Data, p.env)
			//
			switch c.Kind {
			case decl.PRIVATE_READ_ONLY_MEMORY, decl.PUBLIC_READ_ONLY_MEMORY:
				modules = append(modules, memory.NewReadOnly[word.Uint](c.Name(), regs))
			case decl.PRIVATE_WRITE_ONCE_MEMORY, decl.PUBLIC_WRITE_ONCE_MEMORY:
				modules = append(modules, memory.NewWriteOnce[word.Uint](c.Name(), regs))
			case decl.PRIVATE_STATIC_MEMORY, decl.PUBLIC_STATIC_MEMORY:
				// Compile the static initialiser
				words, errs := p.compileStaticInitialisers(declarations, p.env, p.srcmaps, c.Contents...)
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
	// Vectorize modules (if no errors)
	if len(errors) == 0 && p.config.vectorize {
		Vectorize(modules, p.srcmaps)
	}
	// Construct machine
	return machine.New(p.config.field, modules...), errors
}

// compileStaticInitialise evaluates the compile-time constant expressions from a static
// memory declaration into the word.Uint representation required by the VM.
func (p *Compiler) compileStaticInitialisers(
	components []Declaration, env data.ResolvedEnvironment,
	srcmaps source.Maps[any], contents ...expr.Resolved,
) ([]word.Uint, []source.SyntaxError) {
	//
	var (
		words  = make([]word.Uint, len(contents))
		errors []source.SyntaxError
	)
	//
	for i, v := range contents {
		var errMsg string

		words[i], errMsg = EvalConstant(v, true, components, env)
		if errMsg != "" {
			errors = append(errors, srcmaps.SyntaxErrors(v, errMsg)...)
		}
	}

	return words, errors
}

// Convert a decl.Function instance into a fun.Function instance by flattening
// the variable descriptors into register descriptors.  Each variable may
// expand into one or more registers (e.g. a tuple variable produces one
// register per element).
func (p *Compiler) compileFunction(
	id uint, mapping []uint, program []Declaration,
) (*function.Boot[word.Uint], []source.SyntaxError) {
	//
	var (
		fn        = program[id].(*decl.ResolvedFunction)
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

		flattern(v.DataType, v.Name, p.env, func(name string, bitwidth uint) {
			registers = append(registers, register.New(kind, name, bitwidth, padding))
		})
	}
	//
	compiler := StmtCompiler{program, fn.Variables, registers, p.env, p.config.field, p.srcmaps, nil}
	//
	for i, stmt := range fn.Code {
		bootCode[i] = compiler.compileStatement(uint(i), mapping, stmt)
	}
	//
	return function.New(fn.Name(), compiler.registers, bootCode), compiler.errors
}

func toMemoryRegisters(address []VariableDescriptor, datas []VariableDescriptor, env data.ResolvedEnvironment,
) []register.Register {
	var (
		registers []register.Register
		padding   big.Int
	)
	// Flattern address lines
	for _, v := range address {
		flattern(v.DataType, v.Name, env, func(name string, bitwidth uint) {
			registers = append(registers, register.NewInput(name, bitwidth, padding))
		})
	}
	// Flattern data lines
	for _, v := range datas {
		flattern(v.DataType, v.Name, env, func(name string, bitwidth uint) {
			registers = append(registers, register.NewOutput(name, bitwidth, padding))
		})
	}
	//
	return registers
}
