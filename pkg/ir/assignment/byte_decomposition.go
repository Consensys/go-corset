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
package assignment

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ByteDecomposition is part of a range constraint for wide columns (e.g. u32)
// implemented using a byte decomposition.
type ByteDecomposition struct {
	// Handle for identifying this assignment
	handle string
	// Width of decomposition.
	bitwidth uint
	// The source register being decomposed
	source sc.RegisterRef
	// Target registers holding the decomposition
	targets []sc.RegisterRef
}

// NewByteDecomposition creates a new sorted permutation
func NewByteDecomposition(handle string, sourceRegister sc.RegisterRef,
	bitwidth uint, byteRegisters []sc.RegisterRef) *ByteDecomposition {
	//
	return &ByteDecomposition{handle, bitwidth, sourceRegister, byteRegisters}
}

// Compute computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *ByteDecomposition) Compute(tr trace.Trace, schema sc.AnySchema) ([]trace.ArrayColumn, error) {
	var n = uint(len(p.targets))
	// Read inputs
	sources := ReadRegisters(tr, p.source)
	// Apply native function
	data := byteDecompositionNativeFunction(n, sources)
	// Write outputs
	targets := WriteRegisters(schema, p.targets, data)
	//
	return targets, nil
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *ByteDecomposition) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *ByteDecomposition) Consistent(schema sc.AnySchema) []error {
	var (
		bitwidth = schema.Register(p.source).Width
		total    = uint(0)
		errors   []error
	)
	//
	for _, ref := range p.targets {
		reg := schema.Module(ref.Module()).Register(ref.Register())
		total += reg.Width
	}
	//
	if total != bitwidth {
		err := fmt.Errorf("inconsistent byte decomposition (decomposed %d bits, but expected %d)", total, bitwidth)
		errors = append(errors, err)
	}
	//
	return errors
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *ByteDecomposition) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ByteDecomposition) RegistersRead() []sc.RegisterRef {
	return []sc.RegisterRef{p.source}
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *ByteDecomposition) RegistersWritten() []sc.RegisterRef {
	return p.targets
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *ByteDecomposition) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		srcModule = schema.Module(p.source.Module())
		source    = srcModule.Register(p.source.Register())
		targets   = sexp.EmptyList()
	)
	//
	for _, t := range p.targets {
		tgtModule := schema.Module(t.Module())
		reg := tgtModule.Register(t.Register())
		targets.Append(sexp.NewList([]sexp.SExp{
			// name
			sexp.NewSymbol(reg.QualifiedName(tgtModule)),
			// type
			sexp.NewSymbol(fmt.Sprintf("u%d", reg.Width)),
		}))
	}

	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("decompose"),
			targets,
			sexp.NewSymbol(source.QualifiedName(srcModule)),
		})
}

// ============================================================================
// Native Function
// ============================================================================

func byteDecompositionNativeFunction(n uint, sources []field.FrArray) []field.FrArray {
	var (
		source  = sources[0]
		targets = make([]field.FrArray, n)
		height  = source.Len()
	)
	// Sanity check
	if len(sources) != 1 {
		panic("too many source columns for byte decomposition")
	}
	// Initialise columns
	for i := range n {
		// Construct a byte array for ith byte
		targets[i] = field.NewFrArray(height, 8)
	}
	// Decompose each row of each column
	for i := range height {
		ith := decomposeIntoBytes(source.Get(i), n)
		for j := uint(0); j < n; j++ {
			targets[j].Set(i, ith[j])
		}
	}
	//
	return targets
}

// Decompose a given element into n bytes in little endian form.  For example,
// decomposing 41b into 2 bytes gives [0x1b,0x04].
func decomposeIntoBytes(val fr.Element, n uint) []fr.Element {
	// Construct return array
	elements := make([]fr.Element, n)
	// Determine bytes of this value (in big endian form).
	bytes := val.Bytes()
	m := uint(len(bytes) - 1)
	// Convert each byte into a field element
	for i := uint(0); i < n; i++ {
		j := m - i
		ith := fr.NewElement(uint64(bytes[j]))
		elements[i] = ith
	}

	// Done
	return elements
}
