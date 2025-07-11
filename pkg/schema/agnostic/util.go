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
package agnostic

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ApplyMapping applies a given mapping to a set of registers producing a
// corresponding set of limbs.  In essence, each register is convert to its
// limbs in turn, and these are all appended together in order of ococurence.
func ApplyMapping(mapping sc.RegisterMapping, rids []sc.RegisterId) []sc.LimbId {
	var limbs []sc.LimbId
	//
	for _, rid := range rids {
		limbs = append(limbs, mapping.LimbIds(rid)...)
	}
	//
	return limbs
}

// LimbsOf returns those limbs corresponding to a given set of identifiers.
func LimbsOf(mapping sc.RegisterMapping, lids []sc.LimbId) []sc.Limb {
	var (
		limbs []sc.Limb = make([]sc.Limb, len(lids))
	)
	//
	for i, lid := range lids {
		limbs[i] = mapping.Limb(lid)
	}
	//
	return limbs
}

// SplitRawColumns splits a given set of trace columns using the given register mapping.
func SplitRawColumns(columns []trace.RawColumn, mapping sc.RegisterMappings) []trace.RawColumn {
	var splitColumns []trace.RawColumn
	//
	for _, ith := range columns {
		split := SplitRawColumn(ith, mapping)
		splitColumns = append(splitColumns, split...)
	}
	//
	return splitColumns
}

// SplitRawColumn splits a given raw column using the given register mapping.
func SplitRawColumn(column trace.RawColumn, mapping sc.RegisterMappings) []trace.RawColumn {
	var (
		height = column.Data.Len()
		// Access mapping for enclosing module
		modmap = mapping.ModuleOf(column.Module)
		// Determine register id for this column
		reg = modmap.RegisterOf(column.Name)
		// Determine limbs of this register
		limbIds = modmap.LimbIds(reg)
	)
	// Check whether any work actually required
	if len(limbIds) == 1 {
		// No, this register was not split into any limbs.  Therefore, no need
		// to split the column into any limbs.
		return []trace.RawColumn{column}
	}
	// Yes, must split this column into two or more limbs.
	columns := make([]trace.RawColumn, len(limbIds))
	// Determine limbs of this register
	limbs := LimbsOf(modmap, limbIds)
	// Construct empty arrays for the given limbs
	for i, limb := range limbs {
		ith := field.NewFrArray(height, limb.Width)
		columns[i] = trace.RawColumn{Module: column.Module, Name: limb.Name, Data: ith}
	}
	// Determine limb widths of this register (for constant splitting)
	limbWidths := WidthsOfLimbs(modmap, modmap.LimbIds(reg))
	// Deconstruct all data
	for i := range height {
		// Extract ith data
		ith := column.Data.Get(i)
		// Assign split components
		for j, v := range splitFieldElement(ith, limbWidths) {
			columns[j].Data.Set(i, v)
		}
	}
	// Done
	return columns
}

// split a given field element into a given set of limbs, where the least
// significant comes first.  NOTE: this is really a temporary function which
// should be eliminated when RawColumn is moved away from fr.Element.
func splitFieldElement(val fr.Element, widths []uint) []fr.Element {
	var (
		bytes    = val.Bytes()
		bits     = bit.NewReader(bytes[:])
		buf      [32]byte
		elements = make([]fr.Element, len(widths))
	)
	// Reverse bytes of field element into little endian format.  This is
	// necessary because the bit reader reads according to a little endian
	// layout.
	array.ReverseInPlace(bytes[:])
	//
	for i, w := range widths {
		var ith fr.Element
		// Read bits
		m := bits.ReadInto(w, buf[:])
		// Sort back into big endian layout!
		array.ReverseInPlace(buf[:m])
		// Done
		ith.SetBytes(buf[:m])
		elements[i] = ith
	}
	//
	return elements
}
