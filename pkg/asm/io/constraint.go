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
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ConstraintFailure provides structural information about a failing vanishing constraint.
type ConstraintFailure struct {
	// Module where constraint failed
	Context schema.ModuleId
	// Row on which the constraint failed
	Row uint
	// Message
	Msg string
}

// Message provides a suitable error message
func (p *ConstraintFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("constraint failure (row %d): %s", p.Row, p.Msg)
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *ConstraintFailure) RequiredCells(tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return set.NewAnySortedSet[trace.CellRef]()
}

func (p *ConstraintFailure) String() string {
	return p.Message()
}

// Constraint represents a wrapper around an instruction in order for it to
// conform to the constraint interface.
type Constraint[T Instruction[T]] Function[T]

// Accepts implementation for schema.Constraint interface.
func (p Constraint[T]) Accepts(trace tr.Trace) (bit.Set, sc.Failure) {
	// Extract relevant part of the trace
	var (
		coverage bit.Set
		trModule       = trace.Module(p.id)
		state    State = EmptyState(RETURN, p.registers, nil)
	)
	//
	for i := range trModule.Height() {
		// Initialise or check state
		if state.Terminated() {
			// Extract state at start of function instance.
			state = extractState(int(i), state, trModule)
			// Reset to function start
			state.Goto(0)
		}
		// Execute instruction
		pc := p.code[state.Pc()].Execute(state)
		// Sanity check state
		if err := checkState(int(i), state, p.id, trModule); err != nil {
			return coverage, err
		}
		// Execute instruction
		state.Goto(pc)
	}
	// Sanity check frame is complete.
	if !state.Terminated() {
		msg := fmt.Sprintf("function terminated unexpectedly (pc=%d)", state.Pc())
		return coverage, &ConstraintFailure{p.id, trModule.Height() - 1, msg}
	}
	// Success
	return coverage, nil
}

// Bounds implementation for schema.Constraint interface.
func (p Constraint[T]) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Consistent implementation for schema.Constraint interface.
func (p Constraint[T]) Consistent(sc.AnySchema) []error {
	return nil
}

// Contexts implementation for schema.Constraint interface.
func (p Constraint[T]) Contexts() []sc.ModuleId {
	return []sc.ModuleId{p.id}
}

// Name implementation for schema.Constraint interface.
func (p Constraint[T]) Name() string {
	return p.name
}

// Lisp implementation for schema.Constraint interface.
func (p Constraint[T]) Lisp(schema sc.AnySchema) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("function"),
		sexp.NewSymbol(p.name),
	})
}

// Substitute implementation for schema.Constraint interface.
func (p Constraint[T]) Substitute(map[string]fr.Element) {
	// Do nothing since assembly instructions do not (at the time of writing)
	// employ labelled constants.
}

func extractState(row int, state State, trace tr.Module) State {
	//
	for i := range state.registers {
		var (
			rid   = sc.NewRegisterId(uint(i))
			col   = trace.Column(uint(i))
			frVal = col.Get(row)
			biVal big.Int
		)
		// Convert field element to big int
		frVal.BigInt(&biVal)
		// Assign corresponding register
		state.Store(rid, biVal)
	}
	//
	return state
}

func checkState(row int, state State, mid sc.ModuleId, trace tr.Module) sc.Failure {
	// Check each regsiter in turn
	for i := range trace.Width() {
		var (
			rid   = sc.NewRegisterId(i)
			col   = trace.Column(i)
			frVal = col.Get(row)
			biVal big.Int
			stVal = state.Load(rid)
		)
		// Convert field element to big int
		frVal.BigInt(&biVal)
		// Sanity check they match
		if biVal.Cmp(stVal) != 0 {
			msg := fmt.Sprintf("invalid register state (%s holds 0x%s, expected 0x%s)",
				state.registers[i].Name, biVal.Text(16), stVal.Text(16))
			return &ConstraintFailure{mid, uint(row), msg}
		}
	}
	// Success
	return nil
}
