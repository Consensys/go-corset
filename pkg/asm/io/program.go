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

// Program represents a complete set of functions and related declarations
// defining a program.
type Program[T any] interface {
	// Function returns the ith function in this program.
	Function(uint) Function[T]
	// Functions returns the set of functions defined in this program.
	Functions() []Function[T]
}

// NewProgram constructs a new program using a given level of instruction.
func NewProgram[T any](components ...Function[T]) Program[T] {
	return &program[T]{components}
}

// ============================================================================
// Helpers
// ============================================================================

// Simple implementation of Program[T]
type program[T any] struct {
	functions []Function[T]
}

// Function returns the ith function in this program.
func (p *program[T]) Function(id uint) Function[T] {
	return p.functions[id]
}

// Functions returns all functions making up this program.
func (p *program[T]) Functions() []Function[T] {
	return p.functions
}
