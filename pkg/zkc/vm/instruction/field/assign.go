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
package field

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/base"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
)

// Assign from a given source expression to a given set of target
// registers.
type Assign[F field.Element[F]] struct {
	// Target register for assignment
	Target register.Id
	// Source registers for assignment
	Source air.Term[F]
}

// ============================================================================
// Field Instructions
// ============================================================================

// OpCode implementation for Instruction interface
func (p *Assign[F]) OpCode() opcode.OpCode {
	panic("todo")
}

// IsField implementation for instruction.Field interface
func (p *Assign[F]) IsField() bool {
	return true
}

// Uses implementation for Instruction interface.
func (p *Assign[F]) Uses() []register.Id {
	panic("todo")
}

// Definitions implementation for Instruction interface.
func (p *Assign[F]) Definitions() []register.Id {
	return []register.Id{p.Target}
}

func (p *Assign[F]) String(mapping base.SystemMap) string {
	var (
		lhs = base.RegistersToString(mapping, p.Target)
		rhs = p.Source.Lisp(false, mapping).String(false)
	)
	//
	return fmt.Sprintf("%s = %s", lhs, rhs)
}

// Validate implementation for Instruction interface.
func (p *Assign[F]) Validate(_ field.Config, _ base.SystemMap) []error {
	return nil
}

// MicroValidate implementation for Instruction interface.
func (p *Assign[F]) MicroValidate(_ uint, _ field.Config, _ base.SystemMap) []error {
	return nil
}
