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
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// LexicographicSort provides the necessary computation for filling out columns
// added to enforce lexicographic sorting constraints between one or more source
// columns.  Specifically, a delta column is required along with one selector
// column (binary) for each source column.
type LexicographicSort struct {
	// Context in which source and target columns to be located.  All target and
	// source columns should be contained within this.
	context sc.ModuleId
	// The target columns to be filled.  The first entry is for the delta
	// column, and the remaining n entries are for the selector columns.
	targets []sc.RegisterId
	// Source columns being sorted
	sources  []sc.RegisterId
	signs    []bool
	bitwidth uint
}

// LexicographicSortRegisters is a helper for allocated the registers needed for
// a lexicographic sort.
func LexicographicSortRegisters(n uint, prefix string, bitwidth uint) []sc.Register {
	//
	targets := make([]sc.Register, n+1)
	// Create delta column
	targets[0] = sc.NewComputedRegister(fmt.Sprintf("%s:delta", prefix), bitwidth)
	// Create selector columns
	for i := range n {
		ithName := fmt.Sprintf("%s:mux:%d", prefix, i)
		targets[1+i] = sc.NewComputedRegister(ithName, 1)
	}
	//
	return targets
}

// NewLexicographicSort constructs a new LexicographicSorting assignment.
func NewLexicographicSort(context sc.ModuleId,
	targets []sc.RegisterId, signs []bool, sources []sc.RegisterId, bitwidth uint) *LexicographicSort {
	//
	return &LexicographicSort{context, targets, sources, signs, bitwidth}
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *LexicographicSort) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// Compute computes the values of columns defined as needed to support the
// LexicographicSortingGadget. That includes the delta column, and the bit
// selectors.
func (p *LexicographicSort) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	var (
		scModule = schema.Module(p.context)
		trModule = trace.Module(p.context)
		zero     = fr.NewElement(0)
		one      = fr.NewElement(1)
		first    = scModule.Register(p.targets[0])
		// Exact number of (signed) columns involved in the sort
		nbits = len(p.signs)
		// Determine how many rows to be constrained.
		nrows = trace.Module(p.context).Height()
		// Initialise new data columns
		cols = make([]tr.ArrayColumn, nbits+1)
		// Byte width records the largest width of any column.
		bit_width = uint(0)
	)
	// Configure data columns
	for i := 0; i < nbits; i++ {
		target := scModule.Register(p.targets[1+i])
		data := field.NewFrArray(nrows, 1)
		cols[i+1] = tr.NewArrayColumn(target.Name, data, zero)
		// Update bitwidth
		source := trModule.Column(p.sources[i].Unwrap())
		bit_width = max(bit_width, source.Data().BitWidth())
	}
	// Configure data column
	delta := field.NewFrArray(nrows, bit_width)
	cols[0] = tr.NewArrayColumn(first.Name, delta, zero)
	//
	for i := uint(0); i < nrows; i++ {
		set := false
		// Initialise delta to zero
		delta.Set(i, zero)
		// Decide which row is the winner (if any)
		for j := 0; j < nbits; j++ {
			prev := trModule.Column(p.sources[j].Unwrap()).Get(int(i - 1))
			curr := trModule.Column(p.sources[j].Unwrap()).Get(int(i))

			if !set && prev.Cmp(&curr) != 0 {
				var diff fr.Element

				cols[j+1].Data().Set(i, one)
				// Compute curr - prev
				if !p.signs[j] {
					// Swap
					prev, curr = curr, prev
				}
				// Sanity check whether computation is valid.  In cases where a
				// selector is used, then negative (i.e. invalid) values can
				// legitimately arise when the selector is not enabled.  In such
				// cases, we just need any valid filler value.
				if curr.Cmp(&prev) < 0 {
					// Computation is invalid, so use filler of zero.
					diff.Set(&zero)
				} else {
					diff.Sub(&curr, &prev)
				}
				//
				delta.Set(i, diff)
				//
				set = true
			} else {
				cols[j+1].Data().Set(i, zero)
			}
		}
	}
	// Done.
	return cols, nil
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *LexicographicSort) Consistent(schema sc.AnySchema) []error {
	var (
		errors   []error
		bitwidth = uint(0)
		module   = schema.Module(p.context)
	)
	// Sanity check source types
	for i := range p.sources {
		source := module.Register(p.sources[i])
		// i+1 because first target is selector
		target := module.Register(p.targets[i+1])
		// Sanit checkout
		if source.Width != target.Width {
			errors = append(errors,
				fmt.Errorf("lexicographic sort has inconsistent type for column %s (was u%d, expected u%d)",
					source.Name, target.Width, source.Width))
		}
		//
		bitwidth = max(bitwidth, source.Width)
	}
	// sanity check bitwidth
	if bitwidth != p.bitwidth {
		errors = append(errors,
			fmt.Errorf("lexicographic sort has inconsistent bitwidth (was u%d, expected u%d)", p.bitwidth, bitwidth))
	}
	//
	return errors
}

// Module returns the module which encloses this sorted permutation.
func (p *LexicographicSort) Module() sc.ModuleId {
	return p.context
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *LexicographicSort) RegistersRead() []sc.RegisterId {
	return p.sources
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *LexicographicSort) RegistersWritten() []sc.RegisterId {
	return p.targets
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *LexicographicSort) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		module  = schema.Module(p.context)
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)

	for i := range p.targets {
		ith := module.Register(p.targets[i])
		ith_name := ith.QualifiedName(module)
		targets.Append(sexp.NewSymbol(ith_name))
	}

	for i := range p.sources {
		ith := module.Register(p.sources[i])
		ith_name := ith.QualifiedName(module)
		//
		if i >= len(p.signs) {
			// unsigned column
		} else if p.signs[i] {
			ith_name = fmt.Sprintf("+%s", ith_name)
		} else {
			ith_name = fmt.Sprintf("-%s", ith_name)
		}
		//
		sources.Append(sexp.NewSymbol(ith_name))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("lexicographic-order"),
		targets,
		sources,
	})
}
