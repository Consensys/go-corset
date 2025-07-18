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
	"bytes"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/trace"
)

// RegisterMap provides a generic interface for entities which hold information
// about registers.
type RegisterMap interface {
	// Name returns the name given to the enclosing entity (i.e. module or
	// function).
	Name() string
	// HasRegister checks whether a register with the given name exists and, if
	// so, returns its register identifier.  Otherwise, it returns false.
	HasRegister(name string) (RegisterId, bool)
	// Access a given register in this module.
	Register(RegisterId) Register
	// Registers providers access to the underlying registers of this map.
	Registers() []Register
}

// RegisterId captures the notion of a register index.  That is, for each
// module, every register is allocated a given index starting from 0.  The
// purpose of the wrapper is avoid confusion between uint values and things
// which are expected to identify Columns.
type RegisterId = trace.ColumnId

// NewRegisterId constructs a new register ID from a given raw index.
func NewRegisterId(index uint) RegisterId {
	return trace.NewColumnId(index)
}

// NewUnusedRegisterId constructs something akin to a null reference.  This is
// used in some situations where we may (or may not) want to refer to a specific
// register.
func NewUnusedRegisterId() RegisterId {
	return trace.NewUnusedColumnId()
}

// RegisterRef abstracts a complete (i.e. global) register identifier.
type RegisterRef = trace.ColumnRef

// NewRegisterRef constructs a new register reference from the given module and
// register identifiers.
func NewRegisterRef(mid ModuleId, rid RegisterId) RegisterRef {
	return trace.NewColumnRef(mid, rid)
}

// RegisterType captures the type of a given register, such as whether it
// represents an input column, and output column or a computed register, etc.
type RegisterType struct {
	kind uint8
}

var (
	// INPUT_REGISTER signals a register used for holding the input values of a
	// function.
	INPUT_REGISTER = RegisterType{uint8(0)}
	// OUTPUT_REGISTER signals a register used for holding the output values of
	// a function.
	OUTPUT_REGISTER = RegisterType{uint8(1)}
	// COMPUTED_REGISTER signals a register whose values are computed from one
	// (or more) assignments during trace expansion.
	COMPUTED_REGISTER = RegisterType{uint8(2)}
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
	Kind RegisterType
	// Given name of this register.
	Name string
	// Width (in bits) of this register
	Width uint
}

// NewRegister constructs a new register of a given kind (i.e. input, output or
// computed) with the given name and bitwidth.
func NewRegister(kind RegisterType, name string, bitwidth uint) Register {
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
func (p Register) QualifiedName(mod RegisterMap) string {
	if mod.Name() != "" {
		return fmt.Sprintf("%s:%s", mod.Name(), p.Name)
	}
	//
	return p.Name
}

func (p Register) String() string {
	return fmt.Sprintf("%s:u%d", p.Name, p.Width)
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// WidthOfRegisters returns the combined bitwidth of the given
// registers.  For example, suppose we have three registers: x:u8, y:u8, z:u11.
// Then the combined width is 8+8+11=27.
func WidthOfRegisters(regs []Register, rids []RegisterId) uint {
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

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p RegisterType) GobEncode() (data []byte, err error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	//
	if err := gobEncoder.Encode(&p.kind); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (p *RegisterType) GobDecode(data []byte) error {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
	)
	//
	if err := gobDecoder.Decode(&p.kind); err != nil {
		return err
	}
	// Success!
	return nil
}
