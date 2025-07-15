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
	"fmt"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/word"

	"github.com/consensys/go-corset/pkg/util/field"
)

// ApplyMapping applies a given mapping to a set of registers producing a
// corresponding set of limbs.  In essence, each register is convert to its
// limbs in turn, and these are all appended together in order of ococurence.
func ApplyMapping(mapping sc.ModuleRegisterMap, rids []sc.RegisterId) []sc.LimbId {
	var limbs []sc.LimbId
	//
	for _, rid := range rids {
		limbs = append(limbs, mapping.LimbIds(rid)...)
	}
	//
	return limbs
}

// LimbsOf returns those limbs corresponding to a given set of identifiers.
func LimbsOf(mapping sc.ModuleRegisterMap, lids []sc.LimbId) []sc.Limb {
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

// LowerRawColumns lowers a given set of raw columns into a given field implementation.
func LowerRawColumns(columns []trace.RawColumn[word.BigEndian]) []trace.RawColumn[fr.Element] {
	var loweredColumns []trace.RawFrColumn
	//
	for _, ith := range columns {
		lowered := LowerRawColumn(ith)
		loweredColumns = append(loweredColumns, lowered)
	}
	//
	return loweredColumns
}

// LowerRawColumn lowers a given raw column into a given field implementation.
func LowerRawColumn(column trace.RawColumn[word.BigEndian]) trace.RawColumn[fr.Element] {
	var (
		data  = column.Data
		ndata = field.NewFrArray(data.Len(), data.BitWidth())
	)
	//
	for i := range data.Len() {
		var val fr.Element
		// Initial field element from big endian bytes.
		val.SetBytes(data.Get(i).Bytes())
		//
		ndata.Set(i, val)
	}
	//
	return trace.RawColumn[fr.Element]{
		Module: column.Module,
		Name:   column.Name,
		Data:   ndata,
	}
}

// SplitRawColumns splits a given set of trace columns using the given register mapping.
func SplitRawColumns(columns []trace.RawColumn[word.BigEndian], mapping sc.RegisterMap) []trace.RawFrColumn {
	var splitColumns []trace.RawFrColumn
	//
	for _, ith := range columns {
		split := SplitRawColumn(ith, mapping)
		splitColumns = append(splitColumns, split...)
	}
	//
	return splitColumns
}

// SplitRawColumn splits a given raw column using the given register mapping.
func SplitRawColumn(column trace.RawColumn[word.BigEndian], mapping sc.RegisterMap) []trace.RawFrColumn {
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
		return []trace.RawColumn[fr.Element]{LowerRawColumn(column)}
	}
	// Yes, must split this column into two or more limbs.
	columns := make([]trace.RawFrColumn, len(limbIds))
	// Determine limbs of this register
	limbs := LimbsOf(modmap, limbIds)
	// Construct empty arrays for the given limbs
	for i, limb := range limbs {
		ith := field.NewFrArray(height, limb.Width)
		columns[i] = trace.RawFrColumn{Module: column.Module, Name: limb.Name, Data: ith}
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
func splitFieldElement(val word.BigEndian, widths []uint) []fr.Element {
	var (
		n = len(widths)
		//
		bitwidth = sum(widths...)
		// Determine bytewidth
		bytewidth = byteWidth(bitwidth)
		// Extract bytes whilst ensuring they are in little endian form, and
		// that they match the expected bitwidth.
		bytes = reverseAndPad(val.Bytes(), bytewidth)
		//
		bits     = bit.NewReader(bytes[:])
		buf      [32]byte
		elements = make([]fr.Element, n)
	)
	// read actual bits
	for i, w := range widths {
		var ith fr.Element
		// Read bits
		m := bits.ReadInto(w, buf[:])
		// Convert back to big endian
		array.ReverseInPlace(buf[:m])
		// Done
		ith.SetBytes(buf[:m])
		elements[i] = ith
	}
	//
	return elements
}

func reverseAndPad(bytes []byte, n uint) []byte {
	// Make sure bytes is both padded and cloned.
	switch {
	case n > uint(len(bytes)):
		bytes = array.FrontPad(bytes, n, 0)
	case n == uint(len(bytes)):
		bytes = slices.Clone(bytes)
	case n < uint(len(bytes)):
		panic(fmt.Sprintf("have %d bytes, expected at most %d", len(bytes), n))
	}
	// In place reversal
	array.ReverseInPlace(bytes)
	//
	return bytes
}

func sum(vals ...uint) uint {
	val := uint(0)
	//
	for _, v := range vals {
		val += v
	}
	//
	return val
}

func byteWidth(bitwidth uint) uint {
	bytewidth := bitwidth / 8
	//
	if bitwidth%8 != 0 {
		bytewidth++
	}
	//
	return bytewidth
}
