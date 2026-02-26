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
	"fmt"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

const (
	// EQ indicates an equality condition
	EQ Condition = 0
	// NEQ indicates a non-equality condition
	NEQ Condition = 1
	// LT indicates a less-than condition
	LT Condition = 2
	// GT indicates a greater-than condition
	GT Condition = 3
	// LTEQ indicates a less-than-or-equals condition
	LTEQ Condition = 4
	// GTEQ indicates a greater-than-or-equals condition
	GTEQ Condition = 5
)

// Condition represents the set of possible conditions for an if-goto.
type Condition uint8

// IfGoto describes a conditional branch which branches to a given target
// instruction if the given condition holds.
type IfGoto[E any] struct {
	// Cond indicates the condition
	Cond expr.Condition
	// Target identifies target PC
	Target uint
}

// Buses implementation for Instruction interface
func (p *IfGoto[E]) Buses() []E {
	panic("todo")
}

// Uses implementation for Instruction interface.
func (p *IfGoto[E]) Uses() []variable.Id {
	var (
		reads []variable.Id
		bits  bit.Set = p.Cond.Uses()
	)
	// Collect them all up
	for iter := bits.Iter(); iter.HasNext(); {
		next := iter.Next()
		//
		reads = append(reads, next)
	}
	//
	return reads
}

// Definitions implementation for Instruction interface.
func (p *IfGoto[E]) Definitions() []variable.Id {
	return nil
}

func (p *IfGoto[E]) String(env variable.Map) string {
	return fmt.Sprintf("if %s goto %d", p.Cond.String(env), p.Target)
}

// Validate implementation for Instruction interface.
func (p *IfGoto[E]) Validate(env variable.Map) error {
	//
	return nil
}
