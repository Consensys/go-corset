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
	"fmt"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideLookup(c LookupConstraint, mapping schema.LimbsMap) LookupConstraint {
	var (
		// Determine overall geometry for this lookup.
		geometry = constraint.NewLookupGeometry(c, mapping)
		// Split all registers in the source vectors
		sources = mapLookupVectors(c.Sources, mapping)
		// Split all registers in the target vectors
		targets = mapLookupVectors(c.Targets, mapping)
	)
	//
	rawTargets := splitLookupVectors(targets)
	rawSources := splitLookupVectors(sources)
	//
	alignLookupVectors(rawTargets, geometry, mapping)
	alignLookupVectors(rawSources, geometry, mapping)
	//
	targets = padLookupVectors(rawTargets, geometry)
	sources = padLookupVectors(rawSources, geometry)
	//
	return constraint.NewLookupConstraint(c.Handle, targets, sources)
}

// Mapping lookup vectors essentially means applying the limb mapping to all
// registers used within the lookup vector.  For example, consider a simple
// lookup like "lookup (X) (Y)" where X=>X'1::X'0 and Y=>Y.  Then, after
// mapping, we have "lookup (X'1::X'0) (Y)".  Observe that mapping does not
// create more source/target pairings.  Rather, it splits registers within the
// existing pairings only.  Later stages will subdivide and pad the
// source/target pairings as necessary.
func mapLookupVectors(vectors []ir.Enclosed[[]Term], mapping schema.LimbsMap) []ir.Enclosed[[]Term] {
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

func splitLookupVectors(vectors []ir.Enclosed[[]Term]) []ir.Enclosed[[][]Term] {
	var nterms = make([]ir.Enclosed[[][]Term], len(vectors))
	//
	for i, vector := range vectors {
		var terms = splitLookupVector(vector.Item)
		//
		nterms[i] = ir.Enclose(vector.Module, terms)
	}
	//
	return nterms
}

func splitLookupVector(terms []Term) [][]Term {
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

func alignLookupVectors(vectors []ir.Enclosed[[][]Term], geometry constraint.LookupGeometry,
	mapping schema.LimbsMap) {
	//
	for _, vector := range vectors {
		alignLookupVector(vector, geometry, mapping)
	}
}

func alignLookupVector(vector ir.Enclosed[[][]Term], geometry constraint.LookupGeometry, mapping schema.LimbsMap) {
	var modmap = mapping.Module(vector.Module)
	//
	for i, limbs := range vector.Item {
		// FIXME: somewhere we should check that the selector fits into a single
		// column!
		alignLookupLimbs(i == 0, limbs, geometry.LimbWidths(uint(i)), modmap)
	}
}

func alignLookupLimbs(selector bool, limbs []Term, geometry []uint, mapping schema.RegisterLimbsMap) {
	var (
		n       = len(geometry) - 1
		m       = len(limbs) - 1
		limbMap = mapping.LimbsMap()
	)
	// For now, this is just a check that we have proper alignment.
	for i, limb := range limbs {
		if !selector || limb.IsDefined() {
			// Determine value range of limb
			valrange := limb.ValueRange(limbMap)
			// Determine bitwidth for that range
			bitwidth, signed := valrange.BitWidth()
			// Sanity check for irregular lookups
			if signed {
				panic(fmt.Sprintf("signed lookup encountered (%s)", limb.Lisp(limbMap).String(true)))
			} else if i != n && bitwidth > geometry[i] {
				panic(fmt.Sprintf("irregular lookup detected (u%d v u%d)", bitwidth, geometry[i]))
			} else if i != m && bitwidth != geometry[i] {
				panic(fmt.Sprintf("irregular lookup detected (u%d v u%d)", bitwidth, geometry[i]))
			}
		}
	}
}

func padLookupVectors(vectors []ir.Enclosed[[][]Term], geometry constraint.LookupGeometry) []ir.Enclosed[[]Term] {
	var nterms []ir.Enclosed[[]Term] = make([]ir.Enclosed[[]Term], len(vectors))
	//
	for i, vector := range vectors {
		ith := padTerms(vector.Item, geometry)
		nterms[i] = ir.Enclose(vector.Module, ith)
	}
	//
	return nterms
}

func padTerms(terms [][]Term, geometry constraint.LookupGeometry) []Term {
	var nterms []Term

	for i, ith := range terms {
		n := len(geometry.LimbWidths(uint(i)))
		nterms = append(nterms, padTerm(ith, n)...)
	}

	return nterms
}

func padTerm(terms []Term, geometry int) []Term {
	for len(terms) < geometry {
		// Pad out with zeros
		terms = append(terms, ir.Const64[Term](0))
	}
	//
	return terms
}
