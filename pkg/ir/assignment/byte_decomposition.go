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
	// Context of enclosing module.
	context trace.Context
	// Width of decomposition.
	bitwidth uint
	// The source register being decomposed
	sourceRegister sc.RegisterId
	// Target registers holding the decomposition
	targetRegisters []sc.RegisterId
}

// NewByteDecomposition creates a new sorted permutation
func NewByteDecomposition(handle string, context trace.Context, sourceRegister sc.RegisterId,
	bitwidth uint, byteRegisters []sc.RegisterId) *ByteDecomposition {
	//
	return &ByteDecomposition{handle, context, bitwidth, sourceRegister, byteRegisters}
}

// Compute computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *ByteDecomposition) Compute(tr trace.Trace, schema sc.AnySchema) ([]trace.ArrayColumn, error) {
	var ( // Calculate how many bytes required.
		scModule = schema.Module(p.context.ModuleId)
		trModule = tr.Module(p.context.ModuleId)
		n        = len(p.targetRegisters)
		// Identify source column
		source = trModule.Column(p.sourceRegister.Unwrap())
		// Determine height of column
		height = tr.Height(source.Context())
		// Determine padding values
		padding = decomposeIntoBytes(source.Padding(), n)
		// Construct byte column data
		cols = make([]trace.ArrayColumn, n)
	)
	// Initialise columns
	for i := 0; i < n; i++ {
		ith := scModule.Register(p.targetRegisters[i])
		// Construct a byte array for ith byte
		data := field.NewFrArray(height, 8)
		// Construct a byte column for ith byte
		cols[i] = trace.NewArrayColumn(source.Context(), ith.Name, data, padding[i])
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
func (p *ByteDecomposition) Dependencies() []sc.RegisterId {
	return []sc.RegisterId{p.sourceRegister}
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *ByteDecomposition) Consistent(schema sc.AnySchema) []error {
	var (
		module   = schema.Module(p.Module())
		bitwidth = module.Register(p.sourceRegister).Width
		total    = uint(0)
		errors   []error
	)
	//
	for _, bid := range p.targetRegisters {
		total += module.Register(bid).Width
	}
	//
	if total != bitwidth {
		err := fmt.Errorf("inconsistent byte decomposition (decomposed %d bits, but expected %d)", total, bitwidth)
		errors = append(errors, err)
	}
	//
	return errors
}

// Module returns the enclosing register for all columns computed by this
// assignment.
func (p *ByteDecomposition) Module() uint {
	return p.context.ModuleId
}

// Registers identifies registers assigned by this assignment.
func (p *ByteDecomposition) Registers() []sc.RegisterId {
	return p.targetRegisters
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
func (p *ByteDecomposition) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		module  = schema.Module(p.context.ModuleId)
		source  = module.Register(p.sourceRegister)
		targets = sexp.EmptyList()
	)
	//
	for _, t := range p.targetRegisters {
		reg := module.Register(t)
		targets.Append(sexp.NewList([]sexp.SExp{
			// name
			sexp.NewSymbol(reg.QualifiedName(module)),
			// type
			sexp.NewSymbol(fmt.Sprintf("u%d", reg.Width)),
		}))
	}

	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("decompose"),
			targets,
			sexp.NewSymbol(source.QualifiedName(module)),
		})
}
