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
package gadgets

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

// BitwidthGadget is a general-purpose mechanism for enforcing type constraints
// (i.e. that a given register has a given bitwidth).  Depending on the width
// and config used, this generates appropriate constraints and/or modules as
// necessary for enforcing bitwidth constraints.
type BitwidthGadget struct {
	// Determines the largest bitwidth for which range constraints are
	// translated into AIR range constraints, versus  using a horizontal
	// bitwidth gadget.
	maxRangeConstraint uint
	// Enables the use of type proofs which exploit the
	// limitless prover. Specifically, modules with a recursive structure are
	// created specifically for the purpose of checking types.
	limitless bool
	// Schema into which constraints are placed.
	schema *air.Schema
}

// NewBitwidthGadget constructs a new bitwidth gadget.
func NewBitwidthGadget(schema *air.Schema) *BitwidthGadget {
	return &BitwidthGadget{
		maxRangeConstraint: 8,
		limitless:          false,
		schema:             schema,
	}
}

// WithMaxRangeConstraint determines the cutoff for range cosntraints.
func (p *BitwidthGadget) WithMaxRangeConstraint(width uint) *BitwidthGadget {
	p.maxRangeConstraint = width
	return p
}

// WithLimitless enables or disables use of limitless type proofs.
func (p *BitwidthGadget) WithLimitless(flag bool) *BitwidthGadget {
	p.limitless = flag
	return p
}

// Constrain ensures all values in a given register fit within a given bitwidth.
func (p *BitwidthGadget) Constrain(col uint, bitwidth uint) {
	// Base cases
	switch {
	case bitwidth <= 1:
		p.applyBinaryGadget(col)
		return
	case bitwidth <= p.maxRangeConstraint:
		// Add range constraint
		p.schema.AddRangeConstraint(col, 0, bitwidth)
		// Done
		return
	case p.limitless:
		p.applyRecursiveBitwidthGadget(col, bitwidth)
	default:
		// NOTE: this should be deprecated once the limitless prover is well
		// established.
		p.applyHorizontalBitwidthGadget(col, bitwidth)
	}
}

// ApplyBinaryGadget adds a binarity constraint for a given column in the schema
// which enforces that all values in the given column are either 0 or 1. For a
// column X, this corresponds to the vanishing constraint X * (X-1) == 0.
func (p *BitwidthGadget) applyBinaryGadget(col uint) {
	var (
		// Identify target column
		column = p.schema.Columns().Nth(col)
		// Construct column handle
		handle = fmt.Sprintf("%s:u1", column.Name)
	)
	// Construct X
	X := air.NewColumnAccess(col, 0)
	// Construct X-1
	X_m1 := X.Sub(air.NewConst64(1))
	// Construct X * (X-1)
	X_X_m1 := X.Mul(X_m1)
	// Done!
	p.schema.AddVanishingConstraint(handle, 0, column.Context, util.None[int](), X_X_m1)
}

// ApplyHorizontalBitwidthGadget ensures all values in a given column fit within
// a given bitwidth.  This is implemented using a *horizontal byte
// decomposition* which adds n columns and a vanishing constraint (where n*8 >=
// bitwidth).
func (p *BitwidthGadget) applyHorizontalBitwidthGadget(col uint, bitwidth uint) {
	var (
		// Determine ranges required for the give bitwidth
		ranges, widths = splitColumnRanges(bitwidth)
		// Identify number of columns required.
		n = uint(len(ranges))
	)
	// Sanity check
	if bitwidth == 0 {
		panic("zero bitwidth constraint encountered")
	}
	// Identify target column
	column := p.schema.Columns().Nth(col)
	// Calculate how many bytes required.
	es := make([]air.Expr, n)
	name := column.Name
	coefficient := fr.NewElement(1)
	// Add decomposition assignment
	index := p.schema.AddAssignment(
		assignment.NewByteDecomposition(name, column.Context, col, bitwidth))
	// Construct Columns
	for i := uint(0); i < n; i++ {
		// Create Column + Constraint
		es[i] = air.NewColumnAccess(index+i, 0).Mul(air.NewConst(coefficient))

		p.schema.AddRangeConstraint(index+i, 0, widths[i])
		// Update coefficient
		coefficient.Mul(&coefficient, &ranges[i])
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	sum := air.Sum(es...)
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := air.NewColumnAccess(col, 0)
	eq := X.Equate(sum)
	// Construct column name
	p.schema.AddVanishingConstraint(
		fmt.Sprintf("%s:u%d", name, bitwidth), 0, column.Context, util.None[int](), eq)
}

// ApplyRecursiveBitwidthGadget ensures all values in a given column fit within
// a given bitwidth. This is implemented using a combination of reference tables
// and lookups.  Specifically, if the width is below 16bits, then a static
// reference table is created along with a corresponding lookup,  Otherwise, a
// recursive procedure is applied whereby a table is created for the given width
// which divides each value into two smaller values.  This procedure is then
// recursively applied to those columns, etc.
func (p *BitwidthGadget) applyRecursiveBitwidthGadget(col uint, bitwidth uint) {
	panic("todo")
}

// ============================================================================
// Helpers (for recursive)
// ============================================================================

// Determine the split of limbs for the given bitwidth.  For example, 33bits
// could be broken into 16bit and 17bit limbs, or into 8bit and 25bit limbs or
// into 32bit and 1bit limbs.  The current algorithm ensures the least
// significant limb is always a power of 2, and it tries to balance the limbs as
// much as possible (i.e. to reduce the tree depth).
func determineLimbSplit(bitwidth uint) (uint, uint) {
	var (
		pivot      = bitwidth / 2
		loMaxWidth = uint(1)
		loMinWidth = uint(1)
	)
	// Find nearest power of 2 (upper bound)
	for ; loMaxWidth < pivot; loMaxWidth = loMaxWidth * 2 {
		loMinWidth = loMaxWidth
	}
	// Decide which option gives better balance
	lowerDelta := pivot - loMinWidth
	upperDelta := loMaxWidth - pivot
	//
	if lowerDelta < upperDelta {
		return loMinWidth, bitwidth - loMinWidth
	}
	//
	return loMaxWidth, bitwidth - loMaxWidth
}

// Combine all values from the given source registers into a single array of
// data, whilst eliminating duplicates.
func combineSources(bitwidth uint, sources []field.FrArray) field.FrArray {
	var arr = field.NewFrIndexArray(0, bitwidth)
	// Always include zero to work around limitations of FrIndexArray.  This is
	// not actually inefficient, since all columns are subject to an initial
	// padding row anyway.
	arr.Append(fr.NewElement(0))
	//
	for _, src := range sources {
		for i := range src.Len() {
			ith := src.Get(i)
			// Add ith item if not already seen.
			if _, ok := arr.IndexOf(ith); !ok {
				arr.Append(src.Get(i))
			}
		}
	}
	// Done
	return arr
}

func computeDecomposition(loWidth, hiWidth uint, vArr field.FrArray) []field.FrArray {
	// FIXME: using an index array here ensures the underlying data is
	// represented using a full field element, rather than e.g. some smaller
	// number of bytes.  This is needed to handle reject tests which can produce
	// values outside the range of the computed register, but which we still
	// want to check are actually rejected (i.e. since they are simulating what
	// an attacker might do).
	var (
		vLoArr = field.NewFrIndexArray(vArr.Len(), loWidth)
		vHiArr = field.NewFrIndexArray(vArr.Len(), hiWidth)
	)
	//
	for i := range vArr.Len() {
		ith := vArr.Get(i)
		lo, hi := decompose(loWidth, ith)
		vLoArr.Set(i, lo)
		vHiArr.Set(i, hi)
	}
	//
	return []field.FrArray{vArr, vLoArr, vHiArr}
}

// Decompose a given field element into its least and most significant limbs,
// based on the required bitwidth for the least significant limb.
func decompose(loWidth uint, ith fr.Element) (fr.Element, fr.Element) {
	// Extract bytes from element
	var (
		bytes      = ith.Bytes()
		loFr, hiFr fr.Element
	)
	// Sanity check assumption
	if loWidth%8 != 0 {
		panic("unreachable")
	}
	//
	n := 32 - (loWidth / 8)
	hiFr.SetBytes(bytes[:n])
	loFr.SetBytes(bytes[n:])
	//
	return loFr, hiFr
}

// ============================================================================
// Helpers (for horizontal)
// ============================================================================

func splitColumnRanges(nbits uint) ([]fr.Element, []uint) {
	var (
		n      = nbits / 8
		m      = int64(nbits % 8)
		ranges []fr.Element
		widths []uint
		fr256  = fr.NewElement(256)
	)
	//
	if m == 0 {
		ranges = make([]fr.Element, n)
		widths = make([]uint, n)
	} else {
		var last fr.Element
		// Most significant column has smaller range.
		ranges = make([]fr.Element, n+1)
		widths = make([]uint, n+1)
		// Determine final range
		last.Exp(fr.NewElement(2), big.NewInt(m))
		//
		ranges[n] = last
		widths[n] = uint(m)
	}
	//
	for i := uint(0); i < n; i++ {
		ranges[i] = fr256
		widths[i] = 8
	}
	//
	return ranges, widths
}
