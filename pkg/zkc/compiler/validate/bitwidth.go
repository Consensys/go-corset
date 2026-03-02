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
package validate

import (
	"fmt"
	"reflect"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Stmt is a convenient alias.
type Stmt = stmt.Stmt[symbol.Resolved]

// Bitwidths validates that each statement in a function's body is correctly
// balanced. Amongst other things, this means ensuring the right number of bits
// are used on the left-hand side given the right-hand side.  For example,
// suppose "x := y + 1" where both x and y are byte registers.  This does not
// balance because the right-hand side generates 9 bits but the left-hand side
// can only consume 8bits.
func Bitwidths(program ast.Program, srcmaps source.Maps[any]) []source.SyntaxError {
	var errors []source.SyntaxError
	//
	for _, d := range program.Components() {
		switch d := d.(type) {
		case *ast.Constant:
			errors = append(errors, validateConstantBitwidth(*d, srcmaps)...)
		case *ast.Function:
			errors = append(errors, validateFunction(*d, srcmaps)...)
		case *ast.Memory:
			// ignore
		default:
			panic(fmt.Sprintf("unknown component: %s", reflect.TypeOf(d).String()))
		}
	}
	//
	return errors
}

func validateConstantBitwidth(fn ast.Constant, srcmaps source.Maps[any]) []source.SyntaxError {
	panic("got here")
}

func validateFunction(fn ast.Function, srcmaps source.Maps[any]) []source.SyntaxError {
	var errors []source.SyntaxError

	for _, s := range fn.Code {
		switch s := s.(type) {
		case *stmt.Assign[symbol.Resolved]:
			errs := validateAssignment(s, fn, srcmaps)
			//
			errors = append(errors, errs...) //*srcmaps.SyntaxError(stmt, err.Error()))
		}
	}
	//
	return errors
}

func validateAssignment(s *stmt.Assign[symbol.Resolved], fn ast.Function, srcmaps source.Maps[any],
) []source.SyntaxError {
	var (
		lhs_bits         = sumTargetBits(s.Targets, &fn)
		rhs_bits, signed = expr.BitWidth[symbol.Resolved](s.Source, &fn)
	)
	// check
	if lhs_bits < rhs_bits {
		return srcmaps.SyntaxErrors(s, fmt.Sprintf("bit overflow (u%d into u%d)", rhs_bits, lhs_bits))
	} else if signed {
		// Sign bit required, so check there is one.
		if err := checkSignBit(s, &fn, srcmaps); err != nil {
			return err
		}
	}
	//
	return checkTargetRegisters(s, &fn, srcmaps)
}

// Sum the total number of bits used by the given set of target registers.
func sumTargetBits(targets []variable.Id, env variable.Map) uint {
	sum := uint(0)
	//
	for _, target := range targets {
		sum += env.Variable(target).BitWidth()
	}
	//
	return sum
}

// the sign bit check is necessary to ensure there is always exactly one sign bit.
func checkSignBit(s *stmt.Assign[symbol.Resolved], env variable.Map, srcmaps source.Maps[any]) []source.SyntaxError {
	var (
		n = len(s.Targets) - 1
	)
	// Sanity check targets
	if n < 0 {
		return srcmaps.SyntaxErrors(s, "malformed assignment")
	}
	// Determine width of sign bit
	signBitWidth := env.Variable(s.Targets[n]).BitWidth()
	// Check it is a single bit
	if signBitWidth == 1 {
		return nil
	}
	// Problem, no alignment.
	return srcmaps.SyntaxErrors(s, fmt.Sprintf("missing sign bit (found u%d most significant bits)", signBitWidth))
}

// CheckTargetRegisters performs some simple checks on a set of target registers
// being written.  Firstly, they cannot be input registers (as this are always
// constant).  Secondly, we cannot write to the same register more than once
// (i.e. a conflicting write).
func checkTargetRegisters(s *stmt.Assign[symbol.Resolved], env variable.Map, srcmaps source.Maps[any]) []source.SyntaxError {
	for i, id := range s.Targets {
		ith := env.Variable(id)
		//
		if ith.IsParameter() {
			return srcmaps.SyntaxErrors(s, fmt.Sprintf("cannot write parameter %s", ith.Name))
		}
		//
		for j := i + 1; j < len(s.Targets); j++ {
			if s.Targets[i] == s.Targets[j] {
				return srcmaps.SyntaxErrors(s, fmt.Sprintf("conflicting write to %s", ith.Name))
			}
		}
	}
	//
	return nil
}
