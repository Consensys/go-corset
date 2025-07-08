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
	"github.com/consensys/go-corset/pkg/util/field"
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
func (p *SortedPermutation) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
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
func (p *SortedPermutation) Subdivide(mapping schema.RegisterMappings) sc.Assignment {
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

func sortedPermutationNativeFunction(sources []field.FrArray, signs []bool) []field.FrArray {
	// Clone target columns first
	targets := cloneNativeFunction(sources)
	// Sort target columns (in place)
	field.PermutationSort(targets, signs)
	//
	return targets
}

func cloneNativeFunction(sources []field.FrArray) []field.FrArray {
	var targets = make([]field.FrArray, len(sources))
	// Clone target columns
	for i, src := range sources {
		// Clone it to initialise permutation.
		targets[i] = src.Clone()
	}
	//
	return targets
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Assignment(&SortedPermutation{}))
}
