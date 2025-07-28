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
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideLookup(c LookupConstraint, mapping schema.LimbsMap) LookupConstraint {
	var (
		// Determine overall geometry for this lookup.
		geometry = lookup.NewGeometry(c, mapping)
		// Split all registers in the source vectors
		sources = mapLookupVectors(c.Sources, mapping)
		// Split all registers in the target vectors
		targets = mapLookupVectors(c.Targets, mapping)
	)
	//
	targets = splitLookupVectors(geometry, targets, mapping)
	sources = splitLookupVectors(geometry, sources, mapping)
	//
	return lookup.NewConstraint(c.Handle, targets, sources)
}

// Mapping lookup vectors essentially means applying the limb mapping to all
// registers used within the lookup vector.  For example, consider a simple
// lookup like "lookup (X) (Y)" where X=>X'1::X'0 and Y=>Y.  Then, after
// mapping, we have "lookup (X'1::X'0) (Y)".  Observe that mapping does not
// create more source/target pairings.  Rather, it splits registers within the
// existing pairings only.  Later stages will subdivide and pad the
// source/target pairings as necessary.
func mapLookupVectors(vectors []lookup.Vector[Term], mapping schema.LimbsMap) []lookup.Vector[Term] {
	var nterms = make([]lookup.Vector[Term], len(vectors))
	//
	for i, vector := range vectors {
		var (
			modmap = mapping.Module(vector.Module)
			terms  = splitTerms(vector.Terms, modmap)
		)
		// TODO: what about the selector itself?
		nterms[i] = lookup.NewVector(vector.Module, vector.Selector, terms...)
	}
	//
	return nterms
}

// Splitting lookup vectors means splitting source/target pairings into one or
// more terms.  For example, consider the lookup "lookup (X'1::X'0) (Y)" where
// X'1, X'0, and Y are all u16.  The geometry of this lookup is [u32]. Assume a
// field bandwidth which cannot hold a u32 without overflow. Then, after
// splitting, we would expect "lookup (X'0, X'1) (Y, 0)". Here, the geometry has
// now changed to [u16,u16] to accommodate the field bandwidth.  Furthermore,
// notice padding has been applied to ensure we have a matching number of
// columns on the left- and right-hand sides.
func splitLookupVectors(geometry lookup.Geometry, vectors []lookup.Vector[Term],
	mapping schema.LimbsMap) []lookup.Vector[Term] {
	//
	var nterms = make([]lookup.Vector[Term], len(vectors))
	//
	for i, vector := range vectors {
		nterms[i] = splitLookupVector(geometry, vector, mapping)
	}
	//
	return nterms
}

func splitLookupVector(geometry lookup.Geometry, vector lookup.Vector[Term],
	mapping schema.LimbsMap) lookup.Vector[Term] {
	//
	var (
		limbs  [][]Term = make([][]Term, vector.Len())
		modmap          = mapping.Module(vector.Module)
	)
	// Initial split
	for i, t := range vector.Terms {
		// Determine value range of ith term
		valrange := t.ValueRange(modmap.LimbsMap())
		// Determine bitwidth for that range
		bitwidth, signed := valrange.BitWidth()
		// Sanity check signed lookups
		if signed {
			panic(fmt.Sprintf("signed lookup encountered (%s)", t.Lisp(modmap).String(true)))
		}
		// Check whether value range exceeds available bandwidth
		if bitwidth > geometry.BandWidth() {
			// Yes, therefore need to split
			//nolint
			if va, ok := t.(*VectorAccess); ok {
				for _, v := range va.Vars {
					limbs[i] = append(limbs[i], v)
				}
			} else {
				// TODO: fix this
				panic("cannot (yet) split lookup term")
			}
		} else {
			// bandwidth is not exceeded, therefore don't split.
			limbs[i] = append(limbs[i], t)
		}
	}
	// Alignment
	for i, limbs := range limbs {
		alignLookupLimbs(limbs, geometry.LimbWidths(uint(i)), modmap)
	}
	// Padding
	nlimbs := padLookupLimbs(limbs, geometry)
	// Done
	return lookup.NewVector(vector.Module, vector.Selector, nlimbs...)
}

// Alignment is related to the potential for so-called "irregular lookups".
// These only arise in relatively unlikely scenarios.  However, since the
// problem is dependent upon the particular field configuration used, it is
// important to support them in order to be truly field agnostic.  To understand
// the issue of alignment, consider this lookup (where X is u160 and Y is u128):
//
// (lookup (X) (Y))
//
// Let's assume we want to split this lookup for a maximum register width of
// u160.  Then, without alignment, we would get this:
//
// (lookup (X 0) (Y'0 Y'1))
//
// This translation is clearly unsound, because the most significant 32 bits of
// X are ignored.  The problem is that, since Y exceed the maximum register
// width it was split accordingly (and recall that splitting attempts to ensure
// a balanced division between limbs); however, since X did not exceed the
// maximum register width, it was not split.
//
// Irregular looks can be resolved only by introducing temporary registers to
// divide X into u128 limbs.  Something like this would be a valid translation
// of the above (where T'0,T'1 are u128 temporaries):
//
// (lookup (T'0 T'1) (Y'0 Y'1)) ; (vanish (X == (T'1 * 2^128) + T'0))
//
// NOTE: For now, this function only checks that limbs are aligned and panics
// otherwise.
func alignLookupLimbs(limbs []Term, geometry []uint, mapping schema.RegisterLimbsMap) {
	var (
		n       = len(geometry) - 1
		m       = len(limbs) - 1
		limbMap = mapping.LimbsMap()
	)
	// For now, this is just a check that we have proper alignment.
	for i, limb := range limbs {
		// Determine value range of limb
		valrange := limb.ValueRange(limbMap)
		// Determine bitwidth for that range
		bitwidth, _ := valrange.BitWidth()
		// Sanity check for irregular lookups
		if i != n && bitwidth > geometry[i] {
			panic(fmt.Sprintf("irregular lookup detected (u%d v u%d)", bitwidth, geometry[i]))
		} else if i != m && bitwidth != geometry[i] {
			panic(fmt.Sprintf("irregular lookup detected (u%d v u%d)", bitwidth, geometry[i]))
		}
	}
}

// Padding is about ensuring a matching number of columns for each source/target
// pairing in the lookup.  To understand the purpose of padding, consider this
// lookup (where X is u256 and Y is u128):
//
// (lookup (X) (Y))
//
// Let's assume we want to split this lookup for a maximum register width of
// u128.  Then, without padding, we would end up with this:
//
// (lookup (X'0 X'1) (Y))
//
// Here, we have a mismatched number of columns because Y did not need to be
// split.  To resolve this, we need to pad the translation of Y as follows:
//
// (lookup (X'0 X'1) (Y 0))
//
// Here, 0 has been appended to the translation of Y to match the number of
// columns required for X.
func padLookupLimbs(terms [][]Term, geometry lookup.Geometry) []Term {
	var nterms []Term

	for i, ith := range terms {
		// Determine expected geometry (i.e. number of columns) at this
		// position.
		n := len(geometry.LimbWidths(uint(i)))
		// Append available terms
		nterms = append(nterms, ith...)
		// Pad out with zeros to match geometry
		for m := n - len(nterms); m > 0; m-- {
			nterms = append(nterms, ir.Const64[Term](0))
		}
	}

	return nterms
}
