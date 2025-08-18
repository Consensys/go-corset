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

import "github.com/consensys/go-corset/pkg/util/field"

// Program represents a complete set of functions and related declarations
// defining a program.
type Program[F field.Element[F], T Instruction[T]] interface {
	// Function returns the ith function in this program.
	Function(uint) Function[F, T]
	// Functions returns the set of functions defined in this program.
	Functions() []*Function[F, T]
}

// NewProgram constructs a new program using a given level of instruction.
func NewProgram[F field.Element[F], T Instruction[T]](components ...*Function[F, T]) Program[F, T] {
	fns := make([]*Function[F, T], len(components))
	copy(fns, components)

	return &program[F, T]{fns}
}

// ============================================================================
// Helpers
// ============================================================================

// Simple implementation of Program[T]
type program[F field.Element[F], T Instruction[T]] struct {
	functions []*Function[F, T]
}

// Function returns the ith function in this program.
func (p *program[F, T]) Function(id uint) Function[F, T] {
	return *p.functions[id]
}

// Functions returns all functions making up this program.
func (p *program[F, T]) Functions() []*Function[F, T] {
	return p.functions
}
