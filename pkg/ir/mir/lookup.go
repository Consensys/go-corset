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
package mir

import (
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideLookup(c LookupConstraint, mapping schema.GlobalLimbMap) LookupConstraint {
	var (
		// Split all registers in the source vectors
		sources = splitEnclosedTerms(c.Sources, mapping)
		// Split all registers in the target vectors
		targets = splitEnclosedTerms(c.Targets, mapping)
	)
	// FIXME: this is not really safe in the general case.  For example, this
	// could result in a mismatched number of columns.  Furthermore, its
	// possible these columns are incorrectly aligned, etc.
	rawTargets := flattenEnclosedVectors(targets)
	rawSources := flattenEnclosedVectors(sources)
	// Determine overall geometry for this lookup.  If this cannot be done,
	// it will panic.  The assumption is that such an error would already be
	// caught earlier in the pipeline.
	geometry := determineLookupGeometry(rawSources, rawTargets)
	//
	targets = padEnclosedVectors(rawTargets, geometry)
	sources = padEnclosedVectors(rawSources, geometry)
	//
	return constraint.NewLookupConstraint(c.Handle, targets, sources)
}

// The "geometry" of a lookup is the maximum width (in columns) of each
// source-target pairing in the lookup.  For example, consider a lookup where (X
// Y) looksup into (A B).  Suppose X and Y split into 1 and 2 columns
// (respectively), whilst A nd B split into 3 and 2 columns (respectively).
// Then, the geometry of the lookup is [3,2], as we need three columns in the
// subdivided lookup to represent the column in the original lookup, and so on.
func determineLookupGeometry(sources []ir.Enclosed[[][]Term], targets []ir.Enclosed[[][]Term]) []uint {
	var geometry []uint = make([]uint, len(sources[0].Item))
	// Sources first
	for _, source := range sources {
		updateLookupGeometry(source, geometry)
	}
	// Targets second
	for _, target := range targets {
		updateLookupGeometry(target, geometry)
	}
	// Done
	return geometry
}

func updateLookupGeometry(source ir.Enclosed[[][]Term], geometry []uint) {
	var terms = source.Item
	// Sanity check
	if len(terms) != len(geometry) {
		// Unreachable, as should be caught earlier in the pipeline.
		panic("misaligned lookup")
	}
	//
	for i, t := range terms {
		geometry[i] = max(geometry[i], uint(len(t)))
	}
}

func splitEnclosedTerms(vectors []ir.Enclosed[[]Term], mapping schema.GlobalLimbMap) []ir.Enclosed[[]Term] {
	var nterms = make([]ir.Enclosed[[]Term], len(vectors))
	//
	for i, vector := range vectors {
		var (
			modmap = mapping.Module(vector.Module)
			terms  = splitTerms(vector.Item, modmap)
		)
		//
		nterms[i] = ir.Enclose(vector.Module, terms)
	}
	//
	return nterms
}

func flattenEnclosedVectors(vectors []ir.Enclosed[[]Term]) []ir.Enclosed[[][]Term] {
	var nterms = make([]ir.Enclosed[[][]Term], len(vectors))
	//
	for i, vector := range vectors {
		var terms = flattenTerms(vector.Item)
		//
		nterms[i] = ir.Enclose(vector.Module, terms)
	}
	//
	return nterms
}

func flattenTerms(terms []Term) [][]Term {
	var nterms [][]Term = make([][]Term, len(terms))
	//
	for i, t := range terms {
		if va, ok := t.(*VectorAccess); ok {
			for _, v := range va.Vars {
				nterms[i] = append(nterms[i], v)
			}
		} else {
			nterms[i] = append(nterms[i], t)
		}
	}
	//
	return nterms
}

func padEnclosedVectors(vectors []ir.Enclosed[[][]Term], geometry []uint) []ir.Enclosed[[]Term] {
	var nterms []ir.Enclosed[[]Term] = make([]ir.Enclosed[[]Term], len(vectors))
	//
	for i, vector := range vectors {
		ith := padTerms(vector.Item, geometry)
		nterms[i] = ir.Enclose(vector.Module, ith)
	}
	//
	return nterms
}

func padTerms(terms [][]Term, geometry []uint) []Term {
	var nterms []Term

	for i, ith := range terms {
		nterms = append(nterms, padTerm(ith, geometry[i])...)
	}

	return nterms
}

func padTerm(terms []Term, geometry uint) []Term {
	for uint(len(terms)) < geometry {
		// Pad out with zeros
		terms = append(terms, ir.Const64[Term](0))
	}
	//
	return terms
}
