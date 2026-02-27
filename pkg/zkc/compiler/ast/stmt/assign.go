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

import (
	"errors"
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Assign represents a generic assignment of the following form:
//
// tn, .., t0 := e
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, e is the source expression.  For example,
// consider this case:
//
// c, r0 := r1 + 1
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 + 1 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to c.  Thus, in this
// particular example, c represents a carry flag.
type Assign[E any] struct {
	// Target registers for assignment
	Targets []variable.Id
	// Source expresion for assignment
	Source expr.Expr
}

// Buses implementation for Instruction interface
func (p *Assign[E]) Buses() []E {
	panic("todo")
}

// Uses implementation for Instruction interface.
func (p *Assign[E]) Uses() []variable.Id {
	return expr.Uses(p.Source)
}

// Definitions implementation for Instruction interface.
func (p *Assign[E]) Definitions() []variable.Id {
	return p.Targets
}

func (p *Assign[E]) String(mapping variable.Map) string {
	var builder strings.Builder
	//
	builder.WriteString(variablesToString(array.Reverse(p.Targets), mapping))
	builder.WriteString(" = ")
	builder.WriteString(p.Source.String(mapping))
	//
	return builder.String()
}

// Validate implementation for Instruction interface.
func (p *Assign[E]) Validate(env variable.Map) error {
	var (
		lhs_bits         = sumTargetBits(p.Targets, env)
		rhs_bits, signed = expr.BitWidth(p.Source, env)
	)
	// check
	if lhs_bits < rhs_bits {
		return fmt.Errorf("bit overflow (u%d into u%d)", rhs_bits, lhs_bits)
	} else if signed {
		// Sign bit required, so check there is one.
		if err := checkSignBit(p.Targets, env); err != nil {
			return err
		}
	}
	//
	return checkTargetRegisters(p.Targets, env)
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
func checkSignBit(targets []variable.Id, env variable.Map) error {
	var n = len(targets) - 1
	// Sanity check targets
	if n < 0 {
		return errors.New("malformed assignment")
	}
	// Determine width of sign bit
	signBitWidth := env.Variable(targets[n]).BitWidth()
	// Check it is a single bit
	if signBitWidth == 1 {
		return nil
	}
	// Problem, no alignment.
	return fmt.Errorf("missing sign bit (found u%d most significant bits)", signBitWidth)
}

// CheckTargetRegisters performs some simple checks on a set of target registers
// being written.  Firstly, they cannot be input registers (as this are always
// constant).  Secondly, we cannot write to the same register more than once
// (i.e. a conflicting write).
func checkTargetRegisters(targets []variable.Id, env variable.Map) error {
	for i, id := range targets {
		ith := env.Variable(id)
		//
		if ith.IsParameter() {
			return fmt.Errorf("cannot write parameter %s", ith.Name)
		}
		//
		for j := i + 1; j < len(targets); j++ {
			if targets[i] == targets[j] {
				return fmt.Errorf("conflicting write to %s", ith.Name)
			}
		}
	}
	//
	return nil
}

// variablesToString returns a string representation for zero or more registers
// separated by a comma.
func variablesToString(rids []variable.Id, env variable.Map) string {
	var builder strings.Builder
	//
	for i := 0; i < len(rids); i++ {
		var rid = rids[i]
		//
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(env.Variable(rid).Name)
	}
	//
	return builder.String()
}
