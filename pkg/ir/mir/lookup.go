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
		// Determine overall geometry for this lookup.  If this cannot be done,
		// it will panic.  The assumption is that such an error would already be
		// caught earlier in the pipeline.
		geometry = determineLookupGeometry(c.Sources, c.Targets, mapping)
		// Split all registers in the source vectors
		sources = splitEnclosedTerms(c.Sources, mapping)
		// Split all registers in the target vectors
		targets = splitEnclosedTerms(c.Targets, mapping)
	)
	// FIXME: this is not really safe in the general case.  For example, this
	// could result in a mismatched number of columns.  Furthermore, its
	// possible these columns are incorrectly aligned, etc.
	targets = flattenEnclosedVectors(targets, geometry)
	sources = flattenEnclosedVectors(sources, geometry)
	//
	return constraint.NewLookupConstraint(c.Handle, targets, sources)
}

// The "geometry" of a lookup is the maximum width of each source-target pairing
// in the lookup.  For example, consider a lookup where (X Y) looksup into (A
// B).  Then, the geometry is (xA yB), where xA is the max width of X and A,
// whilst yB is the max width of Y and B.  We must be able to determine a fixed
// geometry, otherwise we cannot proceed to safely split the lookup.
func determineLookupGeometry(sources []ir.Enclosed[[]Term], targets []ir.Enclosed[[]Term], mapping schema.GlobalLimbMap) []uint {
	var geometry []uint = make([]uint, len(sources[0].Item))
	// Sources first
	for _, source := range sources {
		updateLookupGeometry(source, geometry, mapping)
	}
	// Targets second
	for _, target := range targets {
		updateLookupGeometry(target, geometry, mapping)
	}
	// Done
	return geometry
}

func updateLookupGeometry(source ir.Enclosed[[]Term], geometry []uint, mapping schema.GlobalLimbMap) {
	// var (
	// 	mod   = mapping.Module(source.Module)
	// 	terms = source.Item
	// )
	// // Sanity check
	// if len(terms) != len(geometry) {
	// 	// Unreachable, as should be caught earlier in the pipeline.
	// 	panic("misaligned lookup")
	// }
	// //
	// for i, t := range terms {
	// 	// FIXME: this is currently completely
	// 	bitwidth := t.ValueRange(mod).BitWidth()
	// 	geometry[i] = max(geometry[i], bitwidth)
	// 	// Sanity check
	// 	if bitwidth == math.MaxUint {
	// 		panic(fmt.Sprintf("unknown bitwidth for term \"%s\"", t.Lisp(mod)))
	// 	}
	// }
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

func flattenEnclosedVectors(vectors []ir.Enclosed[[]Term], geometry []uint) []ir.Enclosed[[]Term] {
	var nterms = make([]ir.Enclosed[[]Term], len(vectors))
	//
	for i, vector := range vectors {
		var (
			terms = flattenTerms(vector.Item)
		)
		//
		nterms[i] = ir.Enclose(vector.Module, terms)
	}
	//
	return nterms
}

func flattenTerms(terms []Term) []Term {
	var nterms []Term
	//
	for _, t := range terms {
		if va, ok := t.(*VectorAccess); ok {
			for _, v := range va.Vars {
				nterms = append(nterms, v)
			}
		} else {
			nterms = append(nterms, t)
		}
	}
	//
	return nterms
}
