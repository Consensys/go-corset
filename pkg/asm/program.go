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
	"bytes"
	"encoding/gob"
	"math"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/program"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
)

// MixedProgram represents the composition of an assembly program along with
// zero or more legacy (i.e. external) modules.
type MixedProgram[F field.Element[F], T io.Instruction[T], M schema.Module[F]] struct {
	program io.Program[T]
	// External module declarations.
	externs []M
}

// NewMixedProgram constructs a new program using a given level of instruction.
func NewMixedProgram[F field.Element[F], T io.Instruction[T], M schema.Module[F]](program io.Program[T], externs ...M,
) MixedProgram[F, T, M] {
	return MixedProgram[F, T, M]{program, externs}
}

// Externs returns the set of external modules
func (p *MixedProgram[F, T, M]) Externs() []M {
	return p.externs
}

// Function returns the ith function in this program.
func (p *MixedProgram[F, T, M]) Function(id uint) io.Function[T] {
	return p.program.Function(id)
}

// Functions returns all functions making up this program.
func (p *MixedProgram[F, T, Map]) Functions() []*io.Function[T] {
	return p.program.Functions()
}

// ============================================================================
// Schema interface
// ============================================================================

// Assignments returns an iterator over the assignments of this schema
// These are the computations used to assign values to all computed columns
// in this schema.
func (p *MixedProgram[F, T, M]) Assignments() iter.Iterator[schema.Assignment[F]] {
	return iter.NewFlattenIterator(p.Modules(), func(m schema.Module[F]) iter.Iterator[schema.Assignment[F]] {
		return m.Assignments()
	})
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p *MixedProgram[F, T, M]) Consistent(fieldWidth uint) []error {
	var errors []error
	// Check left
	for _, m := range p.program.Functions() {
		errors = append(errors, m.Validate(fieldWidth)...)
	}
	// Check right
	for _, m := range p.externs {
		errors = append(errors, m.Consistent(fieldWidth, p)...)
	}
	// Done
	return errors
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p *MixedProgram[F, T, M]) Constraints() iter.Iterator[schema.Constraint[F]] {
	return iter.NewFlattenIterator(p.Modules(), func(m schema.Module[F]) iter.Iterator[schema.Constraint[F]] {
		return m.Constraints()
	})
}

// HasModule checks whether a module with the given name exists and, if so,
// returns its module identifier.  Otherwise, it returns false.
func (p *MixedProgram[F, T, M]) HasModule(name module.Name) (schema.ModuleId, bool) {
	for i := range p.Width() {
		if p.Module(i).Name() == name {
			return i, true
		}
	}
	// Fail
	return math.MaxUint, false
}

// Module returns a given module in this schema.
func (p *MixedProgram[F, T, M]) Module(module uint) schema.Module[F] {
	var (
		n = uint(len(p.program.Functions()))
	)
	//
	if module < n {
		return program.NewModule[F](module, p.program.Function(module))
	}
	//
	return p.externs[module-n]
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *MixedProgram[F, T, M]) Modules() iter.Iterator[schema.Module[F]] {
	// Map all functions into modules
	modules := array.Map(p.program.Functions(), func(mid uint, fn *io.Function[T]) schema.Module[F] {
		return program.NewModule[F](mid, *fn)
	})
	// Construct appropriate iterators
	leftIter := iter.NewArrayIterator(modules)
	rightIter := iter.NewArrayIterator(p.externs)
	rightCastingIter := iter.NewCastIterator[M, schema.Module[F]](rightIter)
	// Done
	return iter.NewAppendIterator(leftIter, rightCastingIter)
}

// Register returns the given register in this schema.
func (p *MixedProgram[F, T, M]) Register(ref register.Ref) Register {
	return p.Module(ref.Module()).Register(ref.Register())
}

// Width returns the number of modules in this schema.
func (p *MixedProgram[F, T, M]) Width() uint {
	return uint(len(p.program.Functions()) + len(p.externs))
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p *MixedProgram[F, T, M]) GobEncode() (data []byte, err error) {
	var buffer bytes.Buffer
	//
	gobEncoder := gob.NewEncoder(&buffer)
	// Left modules
	if err := gobEncoder.Encode(&p.program); err != nil {
		return nil, err
	}
	// Right modules
	if err := gobEncoder.Encode(p.externs); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (p *MixedProgram[F, T, M]) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobDecoder := gob.NewDecoder(buffer)
	// Left modules
	if err := gobDecoder.Decode(&p.program); err != nil {
		return err
	}
	// Right modules
	if err := gobDecoder.Decode(&p.externs); err != nil {
		return err
	}
	// Success!
	return nil
}
