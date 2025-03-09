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
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// LexicographicSort provides the necessary computation for filling out columns
// added to enforce lexicographic sorting constraints between one or more source
// columns.  Specifically, a delta column is required along with one selector
// column (binary) for each source column.
type LexicographicSort struct {
	// Context in which source and target columns to be located.  All target and
	// source columns should be contained within this.
	context tr.Context
	// The target columns to be filled.  The first entry is for the delta
	// column, and the remaining n entries are for the selector columns.
	targets []sc.Column
	// Source columns being sorted
	sources  []uint
	signs    []bool
	bitwidth uint
}

// NewLexicographicSort constructs a new LexicographicSorting assignment.
func NewLexicographicSort(prefix string, context tr.Context,
	sources []uint, signs []bool, bitwidth uint) *LexicographicSort {
	//
	targets := make([]sc.Column, len(signs)+1)
	// Create delta column
	targets[0] = sc.NewColumn(context, fmt.Sprintf("%s:delta", prefix), sc.NewUintType(bitwidth))
	// Create selector columns
	for i := range signs {
		ithName := fmt.Sprintf("%s:%d", prefix, i)
		targets[1+i] = sc.NewColumn(context, ithName, sc.NewUintType(1))
	}

	return &LexicographicSort{context, targets, sources, signs, bitwidth}
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this declaration.
func (p *LexicographicSort) Context() tr.Context {
	return p.context
}

// Columns returns the columns declared by this assignment.
func (p *LexicographicSort) Columns() iter.Iterator[sc.Column] {
	return iter.NewArrayIterator(p.targets)
}

// IsComputed Determines whether or not this declaration is computed (which it
// is).
func (p *LexicographicSort) IsComputed() bool {
	return true
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

// ComputeColumns computes the values of columns defined as needed to support
// the LexicographicSortingGadget. That includes the delta column, and the bit
// selectors.
func (p *LexicographicSort) ComputeColumns(trace tr.Trace) ([]tr.ArrayColumn, error) {
	zero := fr.NewElement(0)
	one := fr.NewElement(1)
	first := p.targets[0]
	// Exact number of (signed) columns involved in the sort
	nbits := len(p.signs)
	// Determine how many rows to be constrained.
	nrows := trace.Height(p.context)
	// Initialise new data columns
	cols := make([]tr.ArrayColumn, nbits+1)
	// Byte width records the largest width of any column.
	bit_width := uint(0)
	// Configure data columns
	for i := 0; i < nbits; i++ {
		target := p.targets[1+i]
		data := field.NewFrArray(nrows, 1)
		cols[i+1] = tr.NewArrayColumn(target.Context, target.Name, data, zero)
		// Update bitwidth
		source := trace.Column(p.sources[i])
		bit_width = max(bit_width, source.Data().BitWidth())
	}
	// Configure data column
	delta := field.NewFrArray(nrows, bit_width)
	cols[0] = tr.NewArrayColumn(first.Context, first.Name, delta, zero)
	//
	for i := uint(0); i < nrows; i++ {
		set := false
		// Initialise delta to zero
		delta.Set(i, zero)
		// Decide which row is the winner (if any)
		for j := 0; j < nbits; j++ {
			prev := trace.Column(p.sources[j]).Get(int(i - 1))
			curr := trace.Column(p.sources[j]).Get(int(i))

			if !set && prev.Cmp(&curr) != 0 {
				var diff fr.Element

				cols[j+1].Data().Set(i, one)
				// Compute curr - prev
				if p.signs[j] {
					diff.Set(&curr)
					delta.Set(i, *diff.Sub(&diff, &prev))
				} else {
					diff.Set(&prev)
					delta.Set(i, *diff.Sub(&diff, &curr))
				}

				set = true
			} else {
				cols[j+1].Data().Set(i, zero)
			}
		}
	}
	// Done.
	return cols, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *LexicographicSort) Dependencies() []uint {
	return p.sources
}

// CheckConsistency performs some simple checks that the given schema is
// consistent.  This provides a double check of certain key properties, such as
// that registers used for assignments are large enough, etc.
func (p *LexicographicSort) CheckConsistency(schema sc.Schema) error {
	bitwidth := uint(0)
	// Sanity check source types
	for i := range p.sources {
		source := schema.Columns().Nth(p.sources[i])
		// i+1 because first target is selector
		target := p.targets[i+1]
		// Sanit checkout
		if source.DataType.Cmp(target.DataType) != 0 {
			return fmt.Errorf("lexicographic sort has inconsistent type for column %s (was %s, expected %s)",
				source.Name, target.DataType, source.DataType)
		}
		//
		bitwidth = max(bitwidth, source.DataType.BitWidth())
	}
	// sanity check bitwidth
	if bitwidth != p.bitwidth {
		return fmt.Errorf("lexicographic sort has inconsistent bitwidth (was %d, expected %d)",
			p.bitwidth, bitwidth)
	}
	//
	return nil
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *LexicographicSort) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	sources := sexp.EmptyList()

	for i := 0; i != len(p.targets); i++ {
		ith := p.targets[i].QualifiedName(schema)
		targets.Append(sexp.NewSymbol(ith))
	}

	for i, s := range p.sources {
		ith := sc.QualifiedName(schema, s)
		//
		if i >= len(p.signs) {
			// unsigned column
		} else if p.signs[i] {
			ith = fmt.Sprintf("+%s", ith)
		} else {
			ith = fmt.Sprintf("-%s", ith)
		}
		//
		sources.Append(sexp.NewSymbol(ith))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("lexicographic-order"),
		targets,
		sources,
	})
}
