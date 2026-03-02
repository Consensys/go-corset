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
package instruction

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Mul represents an instruction of the following form:
//
// tn, .., t0 := r0 * ... * rn * c
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, r0 .. rn are the source registers and c is a
// constant (which can be 1).  For example, consider this case:
//
// c, r0 := r1 * 2
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 * 2 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to c.  Thus, in this
// particular example, c represents a carry flag.
type Mul[W word.Word[W]] struct {
	// Target registers for assignment
	Targets []register.Id
	// Source registers for assignment
	Sources []register.Id
	// Constant for assignment
	Constant W
}

// NewMul constructs a new addition instruction
func NewMul[W word.Word[W]](targets []register.Id, sources []register.Id, constant W) *Mul[W] {
	return &Mul[W]{targets, sources, constant}
}

// Uses implementation for Instruction interface
func (p *Mul[W]) Uses() []register.Id {
	return p.Sources
}

// Definitions implementation for Instruction interface
func (p *Mul[W]) Definitions() []register.Id {
	return p.Targets
}

func (p *Mul[W]) String(mapping register.Map) string {
	var builder strings.Builder
	//
	builder.WriteString(registersToString(array.Reverse(p.Targets), mapping))
	builder.WriteString(" = ")
	builder.WriteString(expressionToString("*", p.Sources, p.Constant, mapping))
	//
	return builder.String()
}

// Validate implementation for Instruction interface.
func (p *Mul[W]) Validate(config field.Config, env register.Map) []error {
	var errors []error
	// (1) validate left-hand side fits within bandwidth; target registers fit
	// within register width; target registers have valid identifiers;
	errors = append(errors, checkTargetRegisters(config, p.Targets, env)...)
	// (2) validate right-hand side within bandwidth;
	if width := p.rhsBitwidth(env); width > config.BandWidth {
		errors = append(errors, fmt.Errorf("right-hand side exceeds target bandwidth (u%d > u%d)", width, config.BandWidth))
	}
	//
	return errors
}

// MicroValidate implementation for MicroInstruction interface.
func (p *Mul[W]) MicroValidate(_ uint, field field.Config, env register.Map) []error {
	return p.Validate(field, env)
}

func (p *Mul[W]) rhsBitwidth(env register.Map) uint {
	var rhs big.Int
	//
	rhs.SetUint64(1)
	//
	for _, target := range p.Sources {
		ith := env.Register(target)
		rhs.Mul(&rhs, ith.MaxValue())
	}
	// Include constant (if relevant)
	rhs.Mul(&rhs, p.Constant.BigInt())
	//
	return uint(rhs.BitLen())
}
