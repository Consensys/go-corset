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
	"math"

	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// UniformSchema represents the simplest kind of schema which contains only
// modules of the same kind (e.g. all MIR modules).
type UniformSchema[M Module] struct {
	modules []M
}

// Sanity check
var _ Schema[Constraint] = UniformSchema[Module]{}

// NewUniformSchema constructs a new schema comprising the given modules.
func NewUniformSchema[M Module](modules []M) UniformSchema[M] {
	return UniformSchema[M]{modules}
}

// Assignments returns an iterator over the assignments of this schema
// These are the computations used to assign values to all computed columns
// in this schema.
func (p UniformSchema[M]) Assignments() iter.Iterator[Assignment] {
	return assignmentsOf(p.modules)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p UniformSchema[M]) Consistent() []error {
	var errors []error
	// Check modules
	for _, m := range p.modules {
		errors = append(errors, m.Consistent(p)...)
	}
	// Done
	return errors
}

// Constraints returns an iterator over all constraints defined in this
// schema.
func (p UniformSchema[M]) Constraints() iter.Iterator[Constraint] {
	return constraintsOf(p.modules)
}

// HasModule checks whether a module with the given name exists and, if so,
// returns its module identifier.  Otherwise, it returns false.
func (p UniformSchema[M]) HasModule(name string) (ModuleId, bool) {
	for i := range p.Width() {
		if p.Module(i).Name() == name {
			return i, true
		}
	}
	// Fail
	return math.MaxUint, false
}

// Module provides access to a given module in this schema.
func (p UniformSchema[M]) Module(module uint) Module {
	return p.modules[module]
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p UniformSchema[M]) Modules() iter.Iterator[Module] {
	arrayIter := iter.NewArrayIterator(p.modules)
	return iter.NewCastIterator[M, Module](arrayIter)
}

// RawModules provides access to the underlying modules of this schema.
func (p UniformSchema[M]) RawModules() []M {
	return p.modules
}

// Register returns the given register in this schema.
func (p UniformSchema[M]) Register(ref RegisterRef) Register {
	return p.Module(ref.Module()).Register(ref.Register())
}

// Width returns the number of modules in this schema.
func (p UniformSchema[M]) Width() uint {
	return uint(len(p.modules))
}

// Extract an iterator over all the constraints in a given array using a
// projecting iterator.
func assignmentsOf[M Module](modules []M) iter.Iterator[Assignment] {
	arrIter := iter.NewArrayIterator(modules)
	//
	return iter.NewFlattenIterator(arrIter, func(m M) iter.Iterator[Assignment] {
		return m.Assignments()
	})
}

// Extract an iterator over all the constraints in a given array using a
// projecting iterator.
func constraintsOf[M Module](modules []M) iter.Iterator[Constraint] {
	arrIter := iter.NewArrayIterator(modules)
	//
	return iter.NewFlattenIterator(arrIter, func(m M) iter.Iterator[Constraint] {
		return m.Constraints()
	})
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p UniformSchema[M]) GobEncode() (data []byte, err error) {
	var buffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&buffer)
	// Modules
	if err := gobEncoder.Encode(&p.modules); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (p *UniformSchema[M]) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobDecoder := gob.NewDecoder(buffer)
	// Modules
	if err := gobDecoder.Decode(&p.modules); err != nil {
		return err
	}
	// Success!
	return nil
}
