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
	"fmt"

	"github.com/consensys/go-corset/pkg/util/source"
)

// Assemble takes a given set of assembly files, and parses them into a given
// set of functions.  This includes performing various checks on the files, such
// as type checking, etc.
func Assemble(assembly ...source.File) ([]Function, []source.SyntaxError) {
	var (
		fns    []Function
		errors []source.SyntaxError
	)
	// Parse each file in turn.
	for _, asm := range assembly {
		fn, errs := Parse(&asm)
		if len(errs) == 0 {
			fns = append(fns, fn...)
		}
		//
		errors = append(errors, errs...)
	}
	// Well-formedness checks
	for _, fn := range fns {
		errors = append(errors, checkWellFormed(fn)...)
	}
	// Done
	return fns, errors
}

// check that a given set of functions are well-formed.  For example, an
// assignment "x,y = z" must be balanced (i.e. number of bits on lhs must match
// number on rhs).  Likewise, registers cannot be used before they are defined,
// and all control-flow paths must reach a "ret" instruction.  Finally, we
// cannot assign to an input register under the current calling convention.
func checkWellFormed(fn Function) []source.SyntaxError {
	errors := checkInstructionsBalance(fn)
	//
	return errors
}

// Check that each instruction in the function's body is correctly balanced.
// Amongst other things, this means ensuring the right number of bits are used
// on the left-hand side given the right-hand side.  For example, suppose "x :=
// y + 1" where both x and y are byte registers.  This does not balance because
// the right-hand side generates 9 bits but the left-hand side can only consume
// 8bits.
func checkInstructionsBalance(fn Function) []source.SyntaxError {
	var errors []source.SyntaxError

	for _, insn := range fn.Code {
		err := insn.IsBalanced(fn.Registers)
		//
		if err != nil {
			panic(fmt.Sprintf("unbalanced instruction! %s", err.Error()))
		}
	}
	//
	return errors
}
