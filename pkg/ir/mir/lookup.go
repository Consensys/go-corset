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

type LookupGeometry struct {
	// Dimensions determines the number of required columns for each lookup
	// pairing.
	dimensions []uint
	// Widths determines the bitwidth required for each lookup pairing.
	bitwidths []uint
}

// NewLookupGeometry constructs a new geometry for a lookup with n pairings.
func NewLookupGeometry(n uint) LookupGeometry {
	return LookupGeometry{
		make([]uint, n), make([]uint, n),
	}
}

func (p *LookupGeometry) applyBitwidths(source ir.Enclosed[[]Term], mapping schema.GlobalLimbMap) {
	var (
		terms  = source.Item
		modmap = mapping.Module(source.Module)
	)
	// Sanity check
	if len(terms) != len(p.dimensions) {
		// Unreachable, as should be caught earlier in the pipeline.
		panic("misaligned lookup")
	}
	//
	for i, ith := range terms {
		if i != 0 || ith.IsDefined() {
			// FIXME: the bitwidth calculation here is incorrect when negative
			// values are encountered.
			bitwidth := ith.ValueRange(modmap).BitWidth()
			//
			p.bitwidths[i] = max(p.bitwidths[i], bitwidth)
		}
	}
}

func (p *LookupGeometry) applyDimensions(source ir.Enclosed[[][]Term], mapping schema.GlobalLimbMap) {
	var terms = source.Item
	// Sanity check
	if len(terms) != len(p.dimensions) {
		// Unreachable, as should be caught earlier in the pipeline.
		panic("misaligned lookup")
	}
	//
	for i, ith := range terms {
		p.dimensions[i] = max(p.dimensions[i], uint(len(ith)))
	}
}

func (p *LookupGeometry) check(handle string, vector ir.Enclosed[[]Term], mapping schema.GlobalLimbMap) {
	var (
		terms  = vector.Item
		modmap = mapping.Module(vector.Module)
	)
	//
	for i, ith := range terms {
		if i != 0 || ith.IsDefined() {
			// FIXME: the bitwidth calculation here is incorrect when negative
			// values are encountered.
			bitwidth := ith.ValueRange(modmap).BitWidth()
			//
			if bitwidth != 0 && bitwidth != p.bitwidths[i] {
				lisp := ith.Lisp(modmap).String(false)
				fmt.Printf("************ Lookup %s.%s mismatch for %s (was %dbits but expecting %dbits)\n",
					modmap.Name(), handle, lisp, bitwidth, p.bitwidths[i])
			}
		}
	}
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideLookup(c LookupConstraint, mapping schema.GlobalLimbMap) LookupConstraint {
	var (
		geometry = NewLookupGeometry(uint(len(c.Sources[0].Item)))
		// Split all registers in the source vectors
		sources = splitEnclosedTerms(c.Sources, mapping)
		// Split all registers in the target vectors
		targets = splitEnclosedTerms(c.Targets, mapping)
	)
	// Sources first
	for _, source := range c.Sources {
		geometry.applyBitwidths(source, mapping)
	}
	// Targets second
	for _, target := range c.Targets {
		geometry.applyBitwidths(target, mapping)
	}
	// Sanity checks
	for _, vector := range c.Targets {
		geometry.check(c.Handle, vector, mapping)
	}

	for _, vector := range c.Sources {
		geometry.check(c.Handle, vector, mapping)
	}
	// FIXME: this is not really safe in the general case.  For example, this
	// could result in a mismatched number of columns.  Furthermore, its
	// possible these columns are incorrectly aligned, etc.
	rawTargets := flattenEnclosedVectors(targets)
	rawSources := flattenEnclosedVectors(sources)
	// Sources first
	for _, source := range rawSources {
		geometry.applyDimensions(source, mapping)
	}
	// Targets second
	for _, target := range rawTargets {
		geometry.applyDimensions(target, mapping)
	}
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
func determineLookupGeometry(sources []ir.Enclosed[[][]Term],
	targets []ir.Enclosed[[][]Term], mapping schema.GlobalLimbMap) LookupGeometry {
	//
	var geometry = NewLookupGeometry(uint(len(sources[0].Item)))
	// Sources first
	for _, source := range sources {
		geometry.applyDimensions(source, mapping)
	}
	// Targets second
	for _, target := range targets {
		geometry.applyDimensions(target, mapping)
	}
	// Done
	return geometry
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

func padEnclosedVectors(vectors []ir.Enclosed[[][]Term], geometry LookupGeometry) []ir.Enclosed[[]Term] {
	var nterms []ir.Enclosed[[]Term] = make([]ir.Enclosed[[]Term], len(vectors))
	//
	for i, vector := range vectors {
		ith := padTerms(vector.Item, geometry)
		nterms[i] = ir.Enclose(vector.Module, ith)
	}
	//
	return nterms
}

func padTerms(terms [][]Term, geometry LookupGeometry) []Term {
	var nterms []Term

	for i, ith := range terms {
		nterms = append(nterms, padTerm(ith, geometry.dimensions[i])...)
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
