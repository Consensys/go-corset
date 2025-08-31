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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
)

// IoState captures the mapping from inputs (i.e. parameters) to outputs (i.e.
// returns) for a particular instance of a given function.
type IoState struct {
	ninputs uint
	state   []byte
}

// LessEq comparator for the I/O registers of a particular function instance.
// Observe that, since functions are always deterministic, this only considers
// the inputs (as the outputs follow directly from this).
func (p IoState) LessEq(other IoState) bool {
	panic("todo")
}

// Executor provides a mechanism for executing a program efficiently and
// generating a suitable top-level trace.  Executor implements the io.Map
// interface.
type Executor[F field.Element[F], T io.Instruction[T]] struct {
	program io.Program[F, T]
	states  []set.AnySortedSet[IoState]
}

// NewExectutor constructs a new executor.
func NewExecutor[F field.Element[F], T io.Instruction[T]](program io.Program[F, T]) Executor[F, T] {
	return Executor[F, T]{program}
}

func (p *Executor[F, T]) Read(bus uint, address []big.Int) []big.Int {
	panic("todo")
}

func (p *Executor[F, T]) Write(bus uint, address []big.Int, values []big.Int) {
	panic("todo")
}
