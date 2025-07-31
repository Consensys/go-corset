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
	bls12_377 "github.com/consensys/go-corset/pkg/util/field/bls12-377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// LexicographicSort provides the necessary computation for filling out columns
// added to enforce lexicographic sorting constraints between one or more source
// columns.  Specifically, a delta column is required along with one selector
// column (binary) for each source column.
type LexicographicSort struct {
	// The target columns to be filled.  The first entry is for the delta
	// column, and the remaining n entries are for the selector columns.
	targets []sc.RegisterRef
	// Source columns being sorted
	sources  []sc.RegisterRef
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
func NewLexicographicSort(targets []sc.RegisterRef, signs []bool, sources []sc.RegisterRef,
	bitwidth uint) *LexicographicSort {
	//
	return &LexicographicSort{targets, sources, signs, bitwidth}
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *LexicographicSort) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute computes the values of columns defined as needed to support the
// LexicographicSortingGadget. That includes the delta column, and the bit
// selectors.
func (p *LexicographicSort) Compute(trace tr.Trace[bls12_377.Element], schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	var (
		// Exact number of (signed) columns involved in the sort
		nbits = len(p.signs)
		// Byte width records the largest width of any column.
		bit_width = uint(0)
	)
	// Compute maximum bitwidth of all source columns, as this determines the
	// width required for the delta column.
	for i := 0; i < nbits; i++ {
		bit_width = max(bit_width, schema.Register(p.sources[i]).Width)
	}
	// Read input columns
	inputs := ReadRegisters(trace, p.sources...)
	// Apply native function
	data := lexicographicSortNativeFunction(bit_width, inputs, p.signs)
	// Write out registers
	outputs := WriteRegisters(schema, p.targets, data)
	//
	return outputs, nil
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *LexicographicSort) Consistent(schema sc.AnySchema) []error {
	var (
		errors   []error
		bitwidth = uint(0)
	)
	// Sanity check source types
	for i := range p.sources {
		source := schema.Register(p.sources[i])
		// i+1 because first target is selector
		target := schema.Register(p.targets[i+1])
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

// RegistersExpanded identifies registers expanded by this assignment.
func (p *LexicographicSort) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *LexicographicSort) RegistersRead() []sc.RegisterRef {
	return p.sources
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *LexicographicSort) RegistersWritten() []sc.RegisterRef {
	return p.targets
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *LexicographicSort) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)

	for i := range p.targets {
		ith := schema.Register(p.targets[i])
		ith_module := schema.Module(p.targets[i].Module())
		ith_name := ith.QualifiedName(ith_module)
		targets.Append(sexp.NewSymbol(ith_name))
	}

	for i := range p.sources {
		ith := schema.Register(p.sources[i])
		ith_module := schema.Module(p.sources[i].Module())
		ith_name := ith.QualifiedName(ith_module)
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

// ============================================================================
// Native Computation
// ============================================================================

func lexicographicSortNativeFunction(bitwidth uint, sources []field.FrArray, signs []bool) []field.FrArray {
	var (
		nrows = sources[0].Len()
		// Number of bit columns required (one for each column being sorted).
		nbits = len(signs)
		// target[0] is for delta column, followed by one bit columns for each
		// column being sorted.
		targets = make([]field.FrArray, 1+nbits)
		//
		one = fr.One()
	)
	// Intialise delta column
	//
	// FIXME: using an index array here ensures the underlying data is
	// represented using a full field element, rather than e.g. some smaller
	// number of bytes.  This is needed to handle reject tests which can produce
	// values outside the range of the computed register, but which we still
	// want to check are actually rejected (i.e. since they are simulating what
	// an attacker might do).
	targets[0] = field.NewFrIndexArray(nrows, bitwidth)
	// Initialise bit columns
	for i := range signs {
		// Construct a byte array for ith byte
		targets[i+1] = field.NewFrArray(nrows, 1)
	}
	//
	for i := uint(1); i < nrows; i++ {
		set := false
		// Initialise delta to zero
		targets[0].Set(i, zero)
		// Decide which row is the winner (if any)
		for j := 0; j < nbits; j++ {
			prev := sources[j].Get(i - 1)
			curr := sources[j].Get(i)

			if !set && prev.Cmp(&curr) != 0 {
				var diff fr.Element

				targets[j+1].Set(i, one)
				// Compute curr - prev
				if !signs[j] {
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
				targets[0].Set(i, diff)
				//
				set = true
			} else {
				targets[j+1].Set(i, zero)
			}
		}
	}
	//
	return targets
}
