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

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// SortedPermutation declares one or more columns as sorted permutations of
// existing columns.
type SortedPermutation struct {
	// Context where in which source and target columns are evaluated.
	ColumnContext sc.ModuleId
	// Target columns declared by this sorted permutation (in the order
	// of declaration).
	Targets []sc.RegisterId
	// Signs determines the sorting direction for each target column.
	Signs []bool
	// Source columns which define the new (sorted) columns.
	Sources []sc.RegisterId
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
	return &SortedPermutation{context, targets, signs, sources}
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *SortedPermutation) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// Compute computes the values of columns defined by this assignment. This
// requires copying the data in the source columns, and sorting that data
// according to the permutation criteria.
func (p *SortedPermutation) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	var ( // Calculate how many bytes required.
		scModule = schema.Module(p.ColumnContext)
		trModule = trace.Module(p.ColumnContext)
		data     = make([]field.FrArray, len(p.Sources))
	)
	// Construct target columns
	for i := range p.Sources {
		src := p.Sources[i]
		// Read column data
		src_data := trModule.Column(src.Unwrap()).Data()
		// Clone it to initialise permutation.
		data[i] = src_data.Clone()
	}
	// Sort target columns
	util.PermutationSort(data, p.Signs)
	// Physically construct the columns
	cols := make([]tr.ArrayColumn, len(p.Sources))
	//
	for i := range p.Sources {
		dstColName := scModule.Register(p.Targets[i]).Name
		srcCol := trModule.Column(p.Sources[i].Unwrap())
		cols[i] = tr.NewArrayColumn(dstColName, data[i], srcCol.Padding())
	}
	//
	return cols, nil
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *SortedPermutation) Consistent(schema sc.AnySchema) []error {
	var (
		module = schema.Module(p.Module())
		errors []error
	)
	// // Sanity check source types
	for i := range p.Sources {
		source := module.Register(p.Sources[i])
		target := module.Register(p.Targets[i])
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

// Module returns the module which encloses this sorted permutation.
func (p *SortedPermutation) Module() sc.ModuleId {
	return p.ColumnContext
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *SortedPermutation) RegistersRead() []sc.RegisterId {
	return p.Sources
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *SortedPermutation) RegistersWritten() []sc.RegisterId {
	return p.Targets
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *SortedPermutation) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		module  = schema.Module(p.Module())
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)

	for _, target := range p.Targets {
		ith := module.Register(target)
		name := sexp.NewSymbol(ith.QualifiedName(module))
		datatype := sexp.NewSymbol(fmt.Sprintf("u%d", ith.Width))
		def := sexp.NewList([]sexp.SExp{name, datatype})
		targets.Append(def)
	}

	for i, s := range p.Sources {
		ith := module.Register(s)
		ith_name := ith.QualifiedName(module)
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
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Assignment(&SortedPermutation{}))
}
