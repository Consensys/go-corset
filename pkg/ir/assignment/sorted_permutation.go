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
	"encoding/gob"
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// SortedPermutation declares one or more columns as sorted permutations of
// existing columns.
type SortedPermutation struct {
	// Target columns declared by this sorted permutation (in the order
	// of declaration).
	Targets []sc.RegisterRef
	// Signs determines the sorting direction for each target column.
	Signs []bool
	// Source columns which define the new (sorted) columns.
	Sources []sc.RegisterRef
}

// NewSortedPermutation creates a new sorted permutation
func NewSortedPermutation(context sc.ModuleId, targets []sc.RegisterId, signs []bool,
	sources []sc.RegisterId) *SortedPermutation {
	//
	if len(targets) != len(sources) {
		panic("target and source column have differing lengths!")
	} else if len(signs) == 0 || len(signs) > len(targets) {
		panic("invalid sort directions")
	}
	//
	return &SortedPermutation{toRegisterRefs(context, targets), signs, toRegisterRefs(context, sources)}
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *SortedPermutation) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute computes the values of columns defined by this assignment. This
// requires copying the data in the source columns, and sorting that data
// according to the permutation criteria.
func (p *SortedPermutation) Compute(trace tr.Trace[bls12_377.Element], schema sc.AnySchema) ([]tr.ArrayColumn[bls12_377.Element], error) {
	// Read inputs
	sources := ReadRegisters(trace, p.Sources...)
	// Apply native function
	data := sortedPermutationNativeFunction(sources, p.Signs)
	// Write outputs
	targets := WriteRegisters(schema, p.Targets, data)
	//
	return targets, nil
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *SortedPermutation) Consistent(schema sc.AnySchema) []error {
	var errors []error
	// // Sanity check source types
	for i := range p.Sources {
		source := schema.Register(p.Sources[i])
		target := schema.Register(p.Targets[i])
		// Sanit checkout
		if source.Width != target.Width {
			err := fmt.Errorf("sorted permutation has inconsistent type for column %s => %s (was u%d, expected u%d)",
				source.Name, target.Name, target.Width, source.Width)
			errors = append(errors, err)
		}
	}
	//
	return errors
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *SortedPermutation) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *SortedPermutation) RegistersRead() []sc.RegisterRef {
	return p.Sources
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *SortedPermutation) RegistersWritten() []sc.RegisterRef {
	return p.Targets
}

// Subdivide implementation for the FieldAgnostic interface.
func (p *SortedPermutation) Subdivide(mapping schema.LimbsMap) sc.Assignment {
	return p
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *SortedPermutation) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)

	for _, t := range p.Targets {
		ith := schema.Register(t)
		name := sexp.NewSymbol(ith.QualifiedName(schema.Module(t.Module())))
		datatype := sexp.NewSymbol(fmt.Sprintf("u%d", ith.Width))
		def := sexp.NewList([]sexp.SExp{name, datatype})
		targets.Append(def)
	}

	for i, s := range p.Sources {
		ith := schema.Register(s)
		ith_name := ith.QualifiedName(schema.Module(s.Module()))
		//
		if i >= len(p.Signs) {

		} else if p.Signs[i] {
			ith_name = fmt.Sprintf("+%s", ith_name)
		} else {
			ith_name = fmt.Sprintf("-%s", ith_name)
		}
		//
		sources.Append(sexp.NewSymbol(ith_name))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("sort"),
		targets,
		sources,
	})
}

// ============================================================================
// Native Function
// ============================================================================

func sortedPermutationNativeFunction[F field.Element[F]](sources []array.Array[F], signs []bool) []array.Array[F] {
	// // Clone target columns first
	// targets := cloneNativeFunction(sources)
	// // Sort target columns (in place)
	// permutationSort(targets, signs)
	// //
	// return targets
	panic("todo")
}

func cloneNativeFunction[F field.Element[F]](sources []array.Array[F]) []field.FrArray {
	// var targets = make([]field.FrArray, len(sources))
	// // Clone target columns
	// for i, src := range sources {
	// 	// Clone it to initialise permutation.
	// 	targets[i] = src.Clone()
	// }
	// //
	// return targets
	panic("todo")
}

// PermutationSort sorts an array of columns in row-wise fashion.  For
// example, suppose consider [ [0,4,3,3], [1,2,4,3] ].  We can imagine
// that this is first transformed into an array of rows (i.e.
// [[0,1],[4,2],[3,4],[3,3]]) and then sorted lexicographically (to
// give [[0,1],[3,3],[3,4],[4,2]]).  This is then projected back into
// the original column-wise formulation, to give: [[0,3,3,4],
// [1,3,4,2]].
//
// A further complication is that the direction of sorting for each
// columns is determined by its sign.
//
// NOTE: the current implementation is not intended to be particularly
// efficient.  In particular, would be better to do the sort directly
// on the columns array without projecting into the row-wise form.
// func permutationSort[T FrArray](cols []T, signs []bool) {
// 	n := cols[0].Len()
// 	m := len(cols)
// 	// Rotate input matrix
// 	rows := rotate(cols, m, n)
// 	// Perform the permutation sort
// 	slices.SortFunc(rows, func(l []fr.Element, r []fr.Element) int {
// 		return permutationSortFunc(l, r, signs)
// 	})
// 	// Project back
// 	for i := uint(0); i < n; i++ {
// 		row := rows[i]
// 		for j := 0; j < m; j++ {
// 			cols[j].Set(i, row[j])
// 		}
// 	}
// }

func permutationSortFunc[F field.Element[F]](lhs []F, rhs []F, signs []bool) int {
	for i := 0; i < len(signs); i++ {
		// Compare ith elements
		c := lhs[i].Cmp(rhs[i])
		// Check whether same
		if c != 0 {
			if signs[i] {
				// Positive
				return c
			}
			// Negative
			return -c
		}
	}
	// Identical
	return 0
}

// Clone and rotate a 2-dimensional array assuming a given geometry.
func rotate[F field.Element[F], T array.MutArray[F]](src []T, ncols int, nrows uint) [][]F {
	// Copy outer arrays
	dst := make([][]F, nrows)
	// Copy inner arrays
	for i := uint(0); i < nrows; i++ {
		row := make([]F, ncols)
		for j := 0; j < ncols; j++ {
			row[j] = src[j].Get(i)
		}

		dst[i] = row
	}
	//
	return dst
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Assignment(&SortedPermutation{}))
}
