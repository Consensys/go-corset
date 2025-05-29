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
	COMPUTED_REGISTER = uint8(2)
)

// Register represents a specific register in the schema that, eventually, will
// be mapped to one (or more) columns in the trace.  Observe that multiple
// registers can end up being mapped to the same column via "register
// allocation".  Likewise, a single register can end up being mapped across
// multiple columns as a result of subdivision to ensure field agnosticity.
// Hence, why they are referred to as registers rather than columns --- they are
// similar, but not identical, concepts.
type Register struct {
	// Kind of register (input / output)
	Kind uint8
	// Given name of this register.
	Name string
	// Width (in bits) of this register
	Width uint
}

// NewRegister constructs a new register of a given kind (i.e. input, output or
// computed) with the given name and bitwidth.
func NewRegister(kind uint8, name string, bitwidth uint) Register {
	return Register{kind, name, bitwidth}
}

// NewInputRegister constructs a new input register with the given name and
// bitwidth.
func NewInputRegister(name string, bitwidth uint) Register {
	return Register{INPUT_REGISTER, name, bitwidth}
}

// NewOutputRegister constructs a new output register with the given name and
// bitwidth.
func NewOutputRegister(name string, bitwidth uint) Register {
	return Register{OUTPUT_REGISTER, name, bitwidth}
}

// NewComputedRegister constructs a new computed register with the given name and
// bitwidth.
func NewComputedRegister(name string, bitwidth uint) Register {
	return Register{COMPUTED_REGISTER, name, bitwidth}
}

// Bound returns the first value which cannot be represented by the given
// bitwidth.  For example, the bound of an 8bit register is 256.
func (p *Register) Bound() *big.Int {
	var (
		bound = big.NewInt(2)
		width = big.NewInt(int64(p.Width))
	)
	// Compute 2^n
	return bound.Exp(bound, width, nil)
}

// IsInput determines whether or not this is an input register
func (p *Register) IsInput() bool {
	return p.Kind == INPUT_REGISTER
}

// IsInputOutput determines whether or not this is an input or output register
func (p *Register) IsInputOutput() bool {
	return p.IsInput() || p.IsOutput()
}

// IsOutput determines whether or not this is an output register
func (p *Register) IsOutput() bool {
	return p.Kind == OUTPUT_REGISTER
}

// IsComputed determines whether or not this is a computed register
func (p *Register) IsComputed() bool {
	return p.Kind == COMPUTED_REGISTER
}

// MaxValue returns the largest value expressible in this register (i.e. Bound() -
// 1).  For example, the max value of an 8bit register is 255.
func (p *Register) MaxValue() *big.Int {
	max := p.Bound()
	max.Sub(max, &one)
	//
	return max
}

var one = *big.NewInt(1)

// QualifiedName returns the fully qualified name of this register
func (p Register) QualifiedName(mod Module) string {
	if mod.Name() != "" {
		return fmt.Sprintf("%s:%s", mod.Name(), p.Name)
	}
	//
	return p.Name
}

func (p Register) String() string {
	return fmt.Sprintf("%s:u%d", p.Name, p.Width)
}
