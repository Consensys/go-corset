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
package macro

import (
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
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
type Assign struct {
	// Target registers for assignment
	Targets []io.RegisterId
	// Source expresion for assignment
	Source Expr
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *Assign) Execute(state io.State) uint {
	panic("todo")
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Assign) Lower(pc uint) micro.Instruction {
	panic("todo")
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Assign) RegistersRead() []io.RegisterId {
	var (
		reads []io.RegisterId
		bits  = p.Source.RegistersRead()
	)
	for iter := bits.Iter(); iter.HasNext(); {
		next := iter.Next()
		//
		reads = append(reads, schema.NewRegisterId(next))
	}
	//
	return reads
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Assign) RegistersWritten() []io.RegisterId {
	return p.Targets
}

func (p *Assign) String(fn schema.RegisterMap) string {
	panic("todo")
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Assign) Validate(fieldWidth uint, fn schema.RegisterMap) error {
	panic("todo")
}
