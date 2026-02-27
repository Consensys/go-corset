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
package stmt

import "github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"

// Stmt provides an abstract notion of a macro "machine instruction".
// Here, macro is intended to imply that the instruction may break down into
// multiple underlying "micro instructions".
type Stmt[N any] interface {
	// Buses identifies any external components (i.e. functions, memories,
	// types) used by this instruction.  For example, a function call will
	// return the identifier of the function being called, etc.
	Buses() []N
	// Uses returns the set of variables used (i.e. read) by this instruction.
	Uses() []variable.Id
	// Definitions returns the set of variables registers defined (i.e. written)
	// by this instruction.
	Definitions() []variable.Id
	// Validate that this instruction is well-formed.  For example, that it is
	// balanced, that there are no conflicting writes, that all temporaries have
	// been allocated, etc.
	Validate(env variable.Map) error
	// Provide human readable form of instruction
	String(env variable.Map) string
}
