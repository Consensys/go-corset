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
package schema

import (
	"fmt"
	"math/big"
)

// ============================================================================

const (
	// INPUT_REGISTER signals a register used for holding the input values of a
	// function.
	INPUT_REGISTER = uint8(0)
	// OUTPUT_REGISTER signals a register used for holding the output values of
	// a function.
	OUTPUT_REGISTER = uint8(1)
	// COMPUTED_REGISTER signals a register whose values are computed from one
	// (or more) assignments during trace expansion.
	COMPUTED_REGISTER = uint8(1)
)

// Column represents a specific column in the schema that, ultimately, will
// correspond 1:1 with a column in the trace.
type Column struct {
	// Kind of register (input / output)
	Kind uint8
	// Given name of this register.
	Name string
	// Width (in bits) of this register
	Width uint
	// Returns the Name of this column
}

func NewColumn(kind uint8, name string, bitwidth uint) Column {
	return Column{kind, name, bitwidth}
}

func NewInputColumn(name string, bitwidth uint) Column {
	return Column{INPUT_REGISTER, name, bitwidth}
}

func NewOutputColumn(name string, bitwidth uint) Column {
	return Column{OUTPUT_REGISTER, name, bitwidth}
}

func NewComputedColumn(name string, bitwidth uint) Column {
	return Column{COMPUTED_REGISTER, name, bitwidth}
}

// Bound returns the first value which cannot be represented by the given
// bitwidth.  For example, the bound of an 8bit register is 256.
func (p *Column) Bound() *big.Int {
	var (
		bound = big.NewInt(2)
		width = big.NewInt(int64(p.Width))
	)
	// Compute 2^n
	return bound.Exp(bound, width, nil)
}

// IsInput determines whether or not this is an input register
func (p *Column) IsInput() bool {
	return p.Kind == INPUT_REGISTER
}

// IsOutput determines whether or not this is an output register
func (p *Column) IsOutput() bool {
	return p.Kind == OUTPUT_REGISTER
}

// IsComputed determines whether or not this is a computed register
func (p *Column) IsComputed() bool {
	return p.Kind == COMPUTED_REGISTER
}

// MaxValue returns the largest value expressible in this register (i.e. Bound() -
// 1).  For example, the max value of an 8bit register is 255.
func (p *Column) MaxValue() *big.Int {
	max := p.Bound()
	max.Sub(max, &one)
	//
	return max
}

var one = *big.NewInt(1)

// QualifiedName returns the fully qualified name of this column
func (p Column) QualifiedName(mod Module) string {
	if mod.Name() != "" {
		return fmt.Sprintf("%s:%s", mod.Name, p.Name)
	}
	//
	return p.Name
}

func (p Column) String() string {
	return fmt.Sprintf("%s:u%d", p.Name, p.Width)
}
