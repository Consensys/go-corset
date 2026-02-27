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
package ast

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
)

// ResolvedSymbol provides linkage information about the given component being
// referenced.  Each component is referred to by its kind (function, RAM, ROM,
// etc) and its index of that kind.
type ResolvedSymbol struct {
	Index uint
}

// Instruction represents a macro instruction  where external identifiers
// are otherwise resolved. As such, it should not be possible that such a
// declaration refers to unknown (or otherwise incorrect) external components.
type Instruction = stmt.Instruction[ResolvedSymbol]

// Declaration represents a declaration which can contain macro
// instructions and where external identifiers are otherwise resolved. As such,
// it should not be possible that such a declaration refers to unknown (or
// otherwise incorrect) external components.
type Declaration = decl.Declaration[ResolvedSymbol]

// Function represents a function which contains instructions whose external
// identifiers are otherwise resolved. As such, it should not be possible that
// such a declaration refers to unknown (or otherwise incorrect) external
// components.
type Function = decl.Function[ResolvedSymbol]

// Memory represents a memory whose external identifiers are otherwise resolved.
// As such, it should not be possible that such a declaration refers to unknown
// (or otherwise incorrect) external components.
type Memory = decl.Memory[ResolvedSymbol]

// UnresolvedSymbol identifies an expect record in the symbol table.  For functions, this
// includes the number of expected inputs and outputs.
type UnresolvedSymbol struct {
	Name            string
	Inputs, Outputs uint
}

// UnresolvedInstruction represents an instruction whose identifiers for external
// components are unresolved linkage records.  As such, its possible that such a
// instruction may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedInstruction = stmt.Instruction[UnresolvedSymbol]

// UnresolvedDeclaration represents a declaration which contains string identifies
// for external (i.e. unlinked) components.  As such, its possible that such a
// declaration may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedDeclaration = decl.Declaration[UnresolvedSymbol]

// UnresolvedFunction represents a function which contains string identifiers
// for external (i.e. unlinked) components.  As such, its possible that such a
// function may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedFunction = decl.Function[UnresolvedSymbol]

// UnresolvedMemory represents a memory which contains string identifiers
// for external (i.e. unlinked) components.  As such, its possible that such a
// function may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedMemory = decl.Memory[UnresolvedSymbol]

// RawProgram encapsulates one of more functions together, such that one may call
// another, etc.  Furthermore, it provides an interface between assembly
// components and the notion of a Schema.
type RawProgram[I any] struct {
	declarations []decl.Declaration[I]
}

// Component returns the ith entity in this program.
func (p *RawProgram[I]) Component(id uint) decl.Declaration[I] {
	return p.declarations[id]
}

// Components returns all functions making up this program.
func (p *RawProgram[I]) Components() []decl.Declaration[I] {
	return p.declarations
}

// Program represents a program whose declarations contain only resolved
// external identifiers. As such, it should not be possible that any
// declarations contained within refer to unknown (or otherwise incorrect)
// external components.
type Program struct {
	RawProgram[ResolvedSymbol]
}

// NewProgram constructs a new program using a given level of instruction.
func NewProgram(components []Declaration) Program {
	//
	decls := make([]Declaration, len(components))
	copy(decls, components)

	return Program{RawProgram[ResolvedSymbol]{decls}}
}

// MapInputs configures a given set of input bytes appropriately for the boot
// program.  If there are missing, unknown or conflicting inputs, then errors
// are returned.
func (p *Program) MapInputs(input map[string][]byte) (map[string][]big.Int, []error) {
	var (
		output  = make(map[string][]big.Int)
		visited = make(map[string]bool)
		errors  []error
	)
	// Initialise components
	for _, c := range p.declarations {
		switch c := c.(type) {
		case *Function:
			// ignore
		case *Memory:
			// Record this memory has seen
			visited[c.Name()] = true
			//
			switch c.Kind {
			case decl.PRIVATE_READ_ONLY_MEMORY, decl.PUBLIC_READ_ONLY_MEMORY:
				if bytes, ok := input[c.Name()]; ok {
					output[c.Name()] = data.DecodeAll(variable.DescriptorsToType(c.Data...), bytes)
				} else {
					errors = append(errors, fmt.Errorf("unexpected input \"%s\"", c.Name()))
				}
			default:
				if _, ok := input[c.Name()]; ok {
					errors = append(errors, fmt.Errorf("unexpected input \"%s\"", c.Name()))
				}
			}
		default:
			panic(fmt.Sprintf("unknown declaration %s", c.Name()))
		}
	}
	// Sanity check for extraneous inputs
	for k := range input {
		if _, ok := visited[k]; !ok {
			errors = append(errors, fmt.Errorf("unknown input \"%s\"", k))
		}
	}
	//
	return output, errors
}

// BuildMachine attempts to build a fresh (bootable) machine which can be used
// (for example) to execute this program with some given inputs.
func (p *Program) BuildMachine() machine.Boot {
	var (
		functions []function.Boot
		statics   []memory.Boot
		inputs    []memory.Boot
		outputs   []memory.Boot
		rams      []memory.Boot
	)
	// Initialise components
	for _, c := range p.declarations {
		switch c := c.(type) {
		case *Function:
			functions = append(functions, compileFunction(*c))
		case *Memory:
			// construct suitable decoder
			var decoder = memory.NewBootDecoder(c.Address, c.Data)
			//
			switch c.Kind {
			case decl.PRIVATE_READ_ONLY_MEMORY, decl.PUBLIC_READ_ONLY_MEMORY:
				inputs = append(inputs, memory.NewArray[big.Int](c.Name(), decoder))
			case decl.PRIVATE_WRITE_ONCE_MEMORY, decl.PUBLIC_WRITE_ONCE_MEMORY:
				outputs = append(outputs, memory.NewArray[big.Int](c.Name(), decoder))
			case decl.PRIVATE_STATIC_MEMORY, decl.PUBLIC_STATIC_MEMORY:
				statics = append(statics, memory.NewArray[big.Int](c.Name(), decoder))
			case decl.RANDOM_ACCESS_MEMORY:
				rams = append(rams, memory.NewArray[big.Int](c.Name(), decoder))
			}
		default:
			panic(fmt.Sprintf("unknown declaration %s", c.Name()))
		}
	}
	// Construct machine (if no errors)
	return machine.NewBoot().
		WithFunctions(functions...).
		WithStatics(statics...).
		WithInputs(inputs...).
		WithOutputs(outputs...).
		WithMemories(rams...)
}

// Convert a decl.Function instance into a fun.Function instance by flattening
// the variable descriptors into register descriptors.  Each variable may
// expand into one or more registers (e.g. a tuple variable produces one
// register per element).
func compileFunction(fn Function) function.Boot {
	var (
		registers []register.Register
		padding   big.Int // zero padding
		bootCode  = make([]instruction.Boot, len(fn.Code))
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
	for i, stmt := range fn.Code {
		bootCode[i], registers = compileStatement(stmt, registers)
	}
	//
	return function.New[instruction.Boot](fn.Name(), registers, bootCode)
}

func compileStatement(s Instruction, registers []register.Register) (instruction.Boot, []register.Register) {
	switch s.(type) {
	case *stmt.Fail[ResolvedSymbol]:
		return &instruction.Fail{}, registers
	case *stmt.Return[ResolvedSymbol]:
		return &instruction.Return{}, registers
	default:
		panic("unknown statement encountered")
	}
}
