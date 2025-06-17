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
package io

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Assignment represents a wrapper around an instruction in order for it to
// conform to the schema.Assignment interface.
type Assignment[T Instruction[T]] Function[T]

// Bounds implementation for schema.Assignment interface.
func (p Assignment[T]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute implementation for schema.Assignment interface.
func (p Assignment[T]) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	panic("todo")
}

// Consistent implementation for schema.Assignment interface.
func (p Assignment[T]) Consistent(sc.AnySchema) []error {
	return nil
}

// Lisp implementation for schema.Assignment interface.
func (p Assignment[T]) Lisp(schema sc.AnySchema) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("function"),
		sexp.NewSymbol(p.name),
	})
}

// RegistersRead implementation for schema.Assignment interface.
func (p Assignment[T]) RegistersRead() []sc.RegisterRef {
	var regs []sc.RegisterRef
	//
	for i, reg := range p.registers {
		if reg.IsInputOutput() {
			rid := sc.NewRegisterId(uint(i))
			regs = append(regs, sc.NewRegisterRef(p.id, rid))
		}
	}
	//
	return regs
}

// RegistersWritten implementation for schema.Assignment interface.
func (p Assignment[T]) RegistersWritten() []sc.RegisterRef {
	var regs []sc.RegisterRef
	//
	for i := range p.registers {
		// Trace expansion writes to all registers, including input/outputs.
		// This is because it may expand the I/O registers.
		rid := sc.NewRegisterId(uint(i))
		regs = append(regs, sc.NewRegisterRef(p.id, rid))
	}
	//
	return regs
}
