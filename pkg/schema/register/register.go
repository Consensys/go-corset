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
package register

import (
	"fmt"
	"math/big"
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
	Kind Type
	// Given name of this register.
	Name string
	// Width (in bits) of this register
	Width uint
	// Determines what value will be used to padd this register.
	Padding big.Int
}

// New constructs a new register of a given kind (i.e. input, output or
// computed) with the given name and bitwidth.
func New(kind Type, name string, bitwidth uint, padding big.Int) Register {
	return Register{kind, name, bitwidth, padding}
}

// NewInput constructs a new input register with the given name and
// bitwidth.
func NewInput(name string, bitwidth uint, padding big.Int) Register {
	return Register{INPUT_REGISTER, name, bitwidth, padding}
}

// NewOutput constructs a new output register with the given name and
// bitwidth.
func NewOutput(name string, bitwidth uint, padding big.Int) Register {
	return Register{OUTPUT_REGISTER, name, bitwidth, padding}
}

// NewComputed constructs a new computed register with the given name and
// bitwidth.
func NewComputed(name string, bitwidth uint, padding big.Int) Register {
	return Register{COMPUTED_REGISTER, name, bitwidth, padding}
}

// NewConst constructs a new "constant register".  That is a register which
// always holds a constant value.  Currently, only constants 0 or 1 are
// supported.
func NewConst(value uint8) Register {
	var name = fmt.Sprintf("%d", value)
	//
	switch value {
	case 0:
		return Register{ZERO_REGISTER, name, 0, *big.NewInt(0)}
	default:
		panic(fmt.Sprintf("unsupported constant register (%d)", value))
	}
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

// IsComputed determines whether or not this is a computed register.  Observer
// that "zero" registers are included in this, since they are neither input nor
// output registers.
func (p *Register) IsComputed() bool {
	return p.Kind == COMPUTED_REGISTER || p.IsConst()
}

// IsConst determines whether or not this is a constant "zero" or "one" register
func (p *Register) IsConst() bool {
	return p.Kind == ZERO_REGISTER || p.Kind == ONE_REGISTER
}

// ConstValue determines the constant value for a given constant register.
func (p *Register) ConstValue() uint8 {
	switch p.Kind {
	case ZERO_REGISTER:
		return 0
	case ONE_REGISTER:
		return 1
	default:
		panic("register not constant")
	}
}

// MaxValue returns the largest value expressible in this register (i.e. Bound() -
// 1).  For example, the max value of an 8bit register is 255.
func (p *Register) MaxValue() *big.Int {
	max := p.Bound()
	max.Sub(max, &one)
	//
	return max
}

// QualifiedName returns the fully qualified name of this register
func (p Register) QualifiedName(mod Map) string {
	var (
		name    = p.Name
		modName = mod.Name().String()
	)
	//
	if modName != "" {
		return fmt.Sprintf("%s:%s", modName, name)
	}
	//
	return name
}

func (p Register) String() string {
	return fmt.Sprintf("%s:u%d:0x%s", p.Name, p.Width, p.Padding.Text(16))
}

// ============================================================================
// Helpers
// ============================================================================

// ToString provides a simplistic default string implementation for a
// RegisterId.  This is useful primarily for debugging where we want to e.g.
// print a constraint but don't have access to an appropriate mapping, etc.
func ToString(rid Id) string {
	return fmt.Sprintf("#%d", rid.Unwrap())
}

// WidthOfRegisters returns the combined bitwidth of the given
// registers.  For example, suppose we have three registers: x:u8, y:u8, z:u11.
// Then the combined width is 8+8+11=27.
func WidthOfRegisters(regs []Register, rids []Id) uint {
	var (
		width uint
	)
	//
	for _, rid := range rids {
		width += regs[rid.Unwrap()].Width
	}
	//
	return width
}
