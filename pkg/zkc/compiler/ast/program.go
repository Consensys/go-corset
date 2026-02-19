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
	"math"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/fun"
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

// BootMemory is the concrete memory type used by a booted machine: a
// pointer to a flat Array of big.Int words addressed via an AddressDecoder.
type BootMemory = *memory.Array[big.Int, AddressDecoder]

// BootState is the concrete runtime state of a booted machine.  It bundles
// the function table together with all memory banks (statics, inputs, outputs,
// RAMs) and the call stack, and is passed by value into each BootExecutor step.
type BootState = machine.BaseState[big.Int, Instruction, BootMemory]

// BootExecutor is the instruction interpreter for resolved AST programs.  It
// is a zero-size struct that implements machine.Executor by dispatching on
// each Instruction variant (assign, if-goto, goto, fail, return) against the
// current BootState.
type BootExecutor = Executor[BootState]

// BootMachine is the fully assembled VM returned by Program.BootMachine.  It
// combines BootState and BootExecutor inside a machine.Base, and exposes an
// Execute method that steps the machine until the call stack is empty or an
// error occurs.
type BootMachine = *machine.Base[big.Int, Instruction, BootMemory, BootExecutor]

// BootMachine attempts to boot a fresh machine to execute this program with the
// given inputs.  However, this can fail with one or more errors if the inputs
// are malformed (e.g. an input is missing or unknown or conflicting).
func (p *Program) BootMachine(input map[string][]byte, mainFn string) (BootMachine, []error) {
	var (
		vm        BootMachine
		functions []fun.Function[Instruction]
		statics   []BootMemory
		inputs    []BootMemory
		outputs   []BootMemory
		rams      []BootMemory
		errors    []error
		visited        = make(map[string]bool)
		main      uint = math.MaxUint
	)
	// Initialise components
	for _, c := range p.declarations {
		switch c := c.(type) {
		case *Function:
			if c.Name() == mainFn {
				main = uint(len(functions))
			}
			//
			functions = append(functions, toFunction(*c))
		case *Memory:
			// Record this memory has seen
			visited[c.Name()] = true
			//
			switch c.Kind {
			case decl.PRIVATE_READ_ONLY_MEMORY, decl.PUBLIC_READ_ONLY_MEMORY:
				// inputs
				inputs, errors = initInputMemory(c, input, inputs, errors)
			case decl.PRIVATE_WRITE_ONCE_MEMORY, decl.PUBLIC_WRITE_ONCE_MEMORY:
				// outputs
				outputs, errors = initOtherMemory(c, input, outputs, errors)
			case decl.PRIVATE_STATIC_MEMORY, decl.PUBLIC_STATIC_MEMORY:
				// static
				statics, errors = initInputMemory(c, input, statics, errors)
			case decl.RANDOM_ACCESS_MEMORY:
				// random-access
				rams, errors = initOtherMemory(c, input, rams, errors)
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
	// Sanity check for main function
	if main == math.MaxUint {
		errors = append(errors, fmt.Errorf("unknown boot function \"%s\"", mainFn))
	}
	// Construct machine (if no errors)
	if len(errors) == 0 {
		vm = machine.New[big.Int, Instruction, BootMemory, Executor[BootState]]().
			WithFunctions(functions...).
			WithStatics(statics...).
			WithInputs(inputs...).
			WithOutputs(outputs...).
			WithMemories(rams...).
			Boot(main)
	}
	// Done
	return vm, errors
}

// initInputMemory initialises a read-only or static memory whose initial
// contents are supplied by the caller through the input map.  The byte slice
// stored under the memory's name is decoded according to the memory's declared
// data type and used to populate a fresh Array.  If the key is absent from
// input an error is appended and acc is returned unchanged.
func initInputMemory(mem *Memory, input map[string][]byte, acc []BootMemory, errs []error) ([]BootMemory, []error) {
	if bytes, ok := input[mem.Name()]; ok {
		decoder := NewAddressDecoder(mem.Address, mem.Data)
		ints := data.DecodeAll(mem.Data, bytes)
		//
		return append(acc, memory.NewArray[big.Int](mem.Name(), decoder, ints...)), nil
	}
	// Error
	return acc, append(errs, fmt.Errorf("missing input \"%s\"", mem.Name()))
}

// initOtherMemory initialises a write-once or random-access memory.  Neither
// kind accepts initial data from the caller; if the memory's name appears in
// input an error is appended.  Write-once memories start empty and will be
// populated by the program; random-access memories likewise start empty.
func initOtherMemory(mem *Memory, input map[string][]byte, acc []BootMemory, errs []error) ([]BootMemory, []error) {
	if _, ok := input[mem.Name()]; ok {
		return acc, append(errs, fmt.Errorf("unexpected input \"%s\"", mem.Name()))
	}
	//
	decoder := NewAddressDecoder(mem.Address, mem.Data)
	//
	return append(acc, memory.NewArray[big.Int](mem.Name(), decoder)), nil
}

// AddressDecoder translates a multi-dimensional logical address into the
// half-open index range [start, end) within the backing flat slice of a
// memory.Array.  The address tuple arrives as a slice of big.Int values,
// decoded from the memory's address data type.  addressGeometry records the
// bit width of each address component; dataGeometry records how many data
// words make up a single row, so that multi-word rows are addressed
// contiguously.
type AddressDecoder struct {
	addressGeometry []uint
	dataGeometry    uint
}

// NewAddressDecoder constructs an AddressDecoder for a memory whose address bus
// has the given address type and whose data bus has the given data type.
// addressGeometry is populated by flattening the address type and collecting
// each leaf's bit width.  dataGeometry is the number of leaves produced by
// flattening the data type (i.e. the number of data words per row).
func NewAddressDecoder(address data.Type, dataType data.Type) AddressDecoder {
	var addrGeom []uint
	address.Flattern("", func(_ string, bitwidth uint) {
		addrGeom = append(addrGeom, bitwidth)
	})

	var dataGeom uint
	dataType.Flattern("", func(_ string, _ uint) {
		dataGeom++
	})

	return AddressDecoder{addrGeom, dataGeom}
}

// Decode maps address (a tuple of big.Int values representing a logical memory
// address) to the half-open index range [start, end) within the backing flat
// slice.  The length end-start always equals dataGeometry, i.e. the number of
// data words per row.
//
// The linear row index is computed by packing the address components
// big-endian: each component is shifted left by the total bit width of all
// subsequent components, then OR-ed in.  For a scalar address this reduces to
// index = address[0]; for a tuple (u8, u16) it gives
// index = address[0]<<16 | address[1].
func (p AddressDecoder) Decode(address []big.Int) (uint64, uint64) {
	var index uint64
	for i, component := range address {
		index = (index << p.addressGeometry[i]) | component.Uint64()
	}

	start := index * uint64(p.dataGeometry)

	return start, start + uint64(p.dataGeometry)
}

// Convert a decl.Function instance into a fun.Function instance by flattening
// the variable descriptors into register descriptors.  Each variable may
// expand into one or more registers (e.g. a tuple variable produces one
// register per element).
func toFunction(fn Function) fun.Function[Instruction] {
	var (
		registers []register.Register
		padding   big.Int // zero padding
	)
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
	return fun.New(fn.Name(), registers, fn.Code)
}
