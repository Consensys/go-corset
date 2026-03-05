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

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Expr represents an expression whose external identifiers are otherwise
// resolved. As such, it should not be possible that such a declaration refers
// to unknown (or otherwise incorrect) external components.
type Expr = expr.Expr[symbol.Resolved]

// Condition represents a condition whose external identifiers are otherwise
// resolved. As such, it should not be possible that such a declaration refers
// to unknown (or otherwise incorrect) external components.
type Condition = expr.Condition[symbol.Resolved]

// Stmt represents a macro instruction  where external identifiers
// are otherwise resolved. As such, it should not be possible that such a
// declaration refers to unknown (or otherwise incorrect) external components.
type Stmt = stmt.Stmt[symbol.Resolved]

// Declaration represents a declaration which can contain macro
// instructions and where external identifiers are otherwise resolved. As such,
// it should not be possible that such a declaration refers to unknown (or
// otherwise incorrect) external components.
type Declaration = decl.Declaration[symbol.Resolved]

// Constant represents a constant whose expression uses only external
// identifiers which are resolved. As such, it should not be possible that such
// a declaration refers to unknown (or otherwise incorrect) external components.
type Constant = decl.Constant[symbol.Resolved]

// Function represents a function which contains instructions whose external
// identifiers are otherwise resolved. As such, it should not be possible that
// such a declaration refers to unknown (or otherwise incorrect) external
// components.
type Function = decl.Function[symbol.Resolved]

// Memory represents a memory whose external identifiers are otherwise resolved.
// As such, it should not be possible that such a declaration refers to unknown
// (or otherwise incorrect) external components.
type Memory = decl.Memory[symbol.Resolved]

// TypeAlias represents an alias whose external identifiers are otherwise resolved.
// As such, it should not be possible that such a declaration refers to unknown
// (or otherwise incorrect) external components.
type TypeAlias = decl.TypeAlias[symbol.Resolved]

// UnresolvedInstruction represents an instruction whose identifiers for external
// components are unresolved linkage records.  As such, its possible that such a
// instruction may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedInstruction = stmt.Stmt[symbol.Unresolved]

// UnresolvedDeclaration represents a declaration which contains string identifies
// for external (i.e. unlinked) components.  As such, its possible that such a
// declaration may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedDeclaration = decl.Declaration[symbol.Unresolved]

// UnresolvedConstant represents a constant whose expression may  contain string
// identifiers for external (i.e. unlinked) components.  As such, its possible
// that such an expression may fail with an error at link time due to an
// unresolvable reference to an external component (e.g. function, RAM, ROM,
// etc).
type UnresolvedConstant = decl.Constant[symbol.Unresolved]

// UnresolvedFunction represents a function which contains string identifiers
// for external (i.e. unlinked) components.  As such, its possible that such a
// function may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedFunction = decl.Function[symbol.Unresolved]

// UnresolvedMemory represents a memory which contains string identifiers
// for external (i.e. unlinked) components.  As such, its possible that such a
// memory may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedMemory = decl.Memory[symbol.Unresolved]

// UnresolvedTypeAlias represents a type alias which contains string identifiers
// for external (i.e. unlinked) components.  As such, its possible that such a
// memory may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedTypeAlias = decl.TypeAlias[symbol.Unresolved]

// UnresolvedExpr represents an expression whose identifiers for external
// components are unresolved linkage records.  As such, its possible that such
// an expression instruction may fail with an error at link time due to an
// unresolvable reference to an external component (e.g. function, RAM, ROM,
// etc).
type UnresolvedExpr = expr.Expr[symbol.Unresolved]

// UnresolvedCondition represents a condition whose identifiers for external
// components are unresolved linkage records.  As such, its possible that such
// an expression instruction may fail with an error at link time due to an
// unresolvable reference to an external component (e.g. function, RAM, ROM,
// etc).
type UnresolvedCondition = expr.Condition[symbol.Unresolved]

// UnresolvedAlias represents a data alias which contains string identifies
// for external (i.e. unlinked) components.  As such, its possible that such a
// declaration may fail with an error at link time due to an unresolvable
// reference to an external component (e.g. function, RAM, ROM, etc).
type UnresolvedAlias = data.Alias[symbol.Unresolved]

// Alias represents an alias whose external identifiers are otherwise resolved.
// As such, it should not be possible that such a type refers to unknown
// (or otherwise incorrect) type.
type Alias = data.Alias[symbol.Resolved]

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
	RawProgram[symbol.Resolved]
}

// NewProgram constructs a new program using a given level of instruction.
func NewProgram(components []Declaration) Program {
	//
	decls := make([]Declaration, len(components))
	copy(decls, components)

	return Program{RawProgram[symbol.Resolved]{decls}}
}

// MapInputs configures a given set of input bytes appropriately for the boot
// program.  If there are missing, unknown or conflicting inputs, then errors
// are returned.
func (p *Program) MapInputs(input map[string][]byte) (map[string][]word.Uint, []error) {
	var (
		output  = make(map[string][]word.Uint)
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
					errors = append(errors, fmt.Errorf("missing input \"%s\"", c.Name()))
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
