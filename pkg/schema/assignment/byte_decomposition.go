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
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ByteDecomposition is part of a range constraint for wide columns (e.g. u32)
// implemented using a byte decomposition.
type ByteDecomposition struct {
	// The source column being decomposed
	source uint
	// Target columns needed for decomposition
	targets []sc.Column
}

// NewByteDecomposition creates a new sorted permutation
func NewByteDecomposition(prefix string, context trace.Context, source uint, bitwidth uint) *ByteDecomposition {
	var n uint = bitwidth / 8
	//
	if bitwidth == 0 {
		panic("zero byte decomposition encountered")
	}
	// Account for asymetric case
	if bitwidth%8 != 0 {
		n++
	}
	// Construct target names
	targets := make([]sc.Column, n)

	for i := uint(0); i < n; i++ {
		name := fmt.Sprintf("%s:%d", prefix, i)
		utype := sc.NewUintType(min(8, bitwidth))
		targets[i] = sc.NewColumn(context, name, utype)
		bitwidth -= 8
	}
	// Done
	return &ByteDecomposition{source, targets}
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this declaration.
func (p *ByteDecomposition) Context() trace.Context {
	return p.targets[0].Context
}

// Columns returns the columns declared by this byte decomposition (in the order
// of declaration).
func (p *ByteDecomposition) Columns() iter.Iterator[sc.Column] {
	return iter.NewArrayIterator(p.targets)
}

// IsComputed Determines whether or not this declaration is computed.
func (p *ByteDecomposition) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// ComputeColumns computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *ByteDecomposition) ComputeColumns(tr trace.Trace) ([]trace.ArrayColumn, error) {
	// Calculate how many bytes required.
	n := len(p.targets)
	// Identify source column
	source := tr.Column(p.source)
	// Determine height of column
	height := tr.Height(source.Context())
	// Determine padding values
	padding := decomposeIntoBytes(source.Padding(), n)
	// Construct byte column data
	cols := make([]trace.ArrayColumn, n)
	// Initialise columns
	for i := 0; i < n; i++ {
		ith := p.targets[i]
		// Construct a byte array for ith byte
		data := field.NewFrArray(height, 8)
		// Construct a byte column for ith byte
		cols[i] = trace.NewArrayColumn(ith.Context, ith.Name, data, padding[i])
	}
	// Decompose each row of each column
	for i := uint(0); i < height; i = i + 1 {
		ith := decomposeIntoBytes(source.Get(int(i)), n)
		for j := 0; j < n; j++ {
			cols[j].Data().Set(i, ith[j])
		}
	}
	// Done
	return cols, nil
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *ByteDecomposition) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ByteDecomposition) Dependencies() []uint {
	return []uint{p.source}
}

// CheckConsistency performs some simple checks that the given schema is
// consistent.  This provides a double check of certain key properties, such as
// that registers used for assignments are large enough, etc.
func (p *ByteDecomposition) CheckConsistency(schema sc.Schema) error {
	n := schema.Columns().Nth(p.source).DataType.ByteWidth()
	//
	if uint(len(p.targets)) != n {
		return fmt.Errorf("inconsistent byte decomposition (have %d byte columns, expected %d)", len(p.targets), n)
	}
	//
	return nil
}

// Decompose a given element into n bytes in little endian form.  For example,
// decomposing 41b into 2 bytes gives [0x1b,0x04].
func decomposeIntoBytes(val fr.Element, n int) []fr.Element {
	// Construct return array
	elements := make([]fr.Element, n)

	// Determine bytes of this value (in big endian form).
	bytes := val.Bytes()
	m := len(bytes) - 1
	// Convert each byte into a field element
	for i := 0; i < n; i++ {
		j := m - i
		ith := fr.NewElement(uint64(bytes[j]))
		elements[i] = ith
	}

	// Done
	return elements
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *ByteDecomposition) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	for _, t := range p.targets {
		targets.Append(sexp.NewList([]sexp.SExp{
			// name
			sexp.NewSymbol(t.QualifiedName(schema)),
			// type
			sexp.NewSymbol(t.DataType.String()),
		}))
	}

	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("decompose"),
			targets,
			sexp.NewSymbol(sc.QualifiedName(schema, p.source)),
		})
}
