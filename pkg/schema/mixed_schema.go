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
	"math"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// Subdivide a mixed schema of field agnostic modules according to the given
// bandwidth and maximum register width requirements.  See discussion of
// FieldAgnosticModule for more on this process.
func Subdivide[M1 FieldAgnosticModule[M1], M2 FieldAgnosticModule[M2]](
	bandwidth uint,
	maxRegisterWidth uint,
	schema MixedSchema[M1, M2]) MixedSchema[M1, M2] {
	//
	var (
		left  []M1 = make([]M1, len(schema.left))
		right []M2 = make([]M2, len(schema.right))
	)
	// Sanity checks
	if bandwidth < maxRegisterWidth {
		panic(
			fmt.Sprintf("field width (%dbits) smaller than register width (%dbits)", bandwidth, maxRegisterWidth))
	}
	// Subdivide the left
	for i, m := range schema.left {
		left[i] = m.Subdivide(bandwidth, maxRegisterWidth)
	}
	// Subdivide the right
	for i, m := range schema.right {
		right[i] = m.Subdivide(bandwidth, maxRegisterWidth)
	}
	// Done
	return MixedSchema[M1, M2]{left, right}
}

// MixedSchema represents a schema comprised of exactly two kinds of concrete
// module.  This are split into those on the "left" (all of one kind) and those
// on the "right" (all of the other kind).  This can be useful, for example, for
// packaging together modules from different layers, such as assembly and legacy
// (i.e. low-level) modules mixed together.
type MixedSchema[M1 Module, M2 Module] struct {
	left  []M1
	right []M2
}

var _ Schema[Constraint] = MixedSchema[Module, Module]{}

// NewMixedSchema constructs a new schema composed of two distinct sets of
// modules, referred to as the "left" and the "right".  Those on the left are
// allocated lower module indices, whilst the indices of those on the right
// begin immediately following the left.
func NewMixedSchema[M1 Module, M2 Module](leftModules []M1, rightModules []M2) MixedSchema[M1, M2] {
	return MixedSchema[M1, M2]{leftModules, rightModules}
}

// Assignments returns an iterator over the assignments of this schema
// These are the computations used to assign values to all computed columns
// in this schema.
func (p MixedSchema[M1, M2]) Assignments() iter.Iterator[Assignment] {
	leftIter := assignmentsOf(p.left)
	rightIter := assignmentsOf(p.right)
	//
	return iter.NewAppendIterator(leftIter, rightIter)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p MixedSchema[M1, M2]) Consistent() []error {
	var errors []error
	// Check left
	for _, m := range p.left {
		errors = append(errors, m.Consistent(p)...)
	}
	// Check right
	for _, m := range p.right {
		errors = append(errors, m.Consistent(p)...)
	}
	// Done
	return errors
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p MixedSchema[M1, M2]) Constraints() iter.Iterator[Constraint] {
	leftIter := constraintsOf(p.left)
	rightIter := constraintsOf(p.right)
	//
	return iter.NewAppendIterator(leftIter, rightIter)
}

// Expand a given trace according to this schema by determining appropriate
// values for all computed columns within the schema.
func (p MixedSchema[M1, M2]) Expand(trace.Trace) (trace.Trace, []error) {
	panic("todo")
}

// HasModule checks whether a module with the given name exists and, if so,
// returns its module identifier.  Otherwise, it returns false.
func (p MixedSchema[M1, M2]) HasModule(name string) (ModuleId, bool) {
	for i := range p.Width() {
		if p.Module(i).Name() == name {
			return i, true
		}
	}
	// Fail
	return math.MaxUint, false
}

// Module returns a given module in this schema.
func (p MixedSchema[M1, M2]) Module(module uint) Module {
	var (
		n = uint(len(p.left))
	)
	//
	if module < n {
		return p.left[module]
	}
	//
	return p.right[module-n]
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p MixedSchema[M1, M2]) Modules() iter.Iterator[Module] {
	leftIter := iter.NewArrayIterator(p.left)
	rightIter := iter.NewArrayIterator(p.right)
	leftCastingIter := iter.NewCastIterator[M1, Module](leftIter)
	rightCastingIter := iter.NewCastIterator[M2, Module](rightIter)
	//
	return iter.NewAppendIterator(leftCastingIter, rightCastingIter)
}

// LeftModules returns those modules which form the "left" part of this mixed
// schema.
func (p MixedSchema[M1, M2]) LeftModules() []M1 {
	return p.left
}

// Register returns the given register in this schema.
func (p MixedSchema[M1, M2]) Register(ref RegisterRef) Register {
	return p.Module(ref.Module()).Register(ref.Register())
}

// RightModules returns those modules which form the "right" part of this mixed
// schema.
func (p MixedSchema[M1, M2]) RightModules() []M2 {
	return p.right
}

// Width returns the number of modules in this schema.
func (p MixedSchema[M1, M2]) Width() uint {
	return uint(len(p.left) + len(p.right))
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p MixedSchema[M1, M2]) GobEncode() (data []byte, err error) {
	var buffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&buffer)
	// Left modules
	if err := gobEncoder.Encode(p.left); err != nil {
		return nil, err
	}
	// Right modules
	if err := gobEncoder.Encode(p.right); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (p *MixedSchema[M1, M2]) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobDecoder := gob.NewDecoder(buffer)
	// Left modules
	if err := gobDecoder.Decode(&p.left); err != nil {
		return err
	}
	// Right modules
	if err := gobDecoder.Decode(&p.right); err != nil {
		return err
	}
	// Success!
	return nil
}
