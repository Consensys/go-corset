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

	"github.com/consensys/go-corset/pkg/air"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// LexicographicSortingGadget adds sorting constraints for a sequence of one or
// more columns.  Sorting is done lexicographically starting from the leftmost
// column.  For example, consider lexicographically sorting two columns X and Y
// (in that order) in ascending (i.e. positive direction).  Then sorting ensures
// (X[k-1] < X[k]) or (X[k-1] == X[k] and Y[k-1] <= Y[k]).  The sign for each
// column determines whether its sorted into ascending (i.e. positive) or
// descending (i.e. negative) order.
//
// To implement this sort, a kind of "bit multiplexing" is used.  Specifically,
// a bit column is associated with each column being sorted, where exactly one
// of these bits can be 1.  That bit identifies the leftmost column Ci where
// Ci[k-1] < C[k].  For all columns Cj where j < i, we must have Cj[k-1] =
// Cj[k].  If all bits are zero then all columns match their previous row.
// Finally, a delta column is used in a similar fashion as for the single column
// case.  The delta value captures the difference Ci[k]-Ci[k-1] to ensure it is
// positive.  The delta column is constrained to a given bitwidth, with
// constraints added as necessary to ensure this.
type LexicographicSortingGadget struct {
	// Prefix is used to construct the delta column name.
	prefix string
	// Identifies column(s) being sorted
	columns []uint
	// Sort direction given for columns (true = ascending, false = descending).
	// Observe that it is not required for all columns to have a sort direction.
	// Columns without a sort direction can be ordered arbitrarily.
	signs []bool
	// Bitwidth of delta column.  This restricts the maximum distance between
	// any two sorted values.  A key requirement is to ensure the delta value is
	// "small" to prevent overflow.
	bitwidth uint
	// Strict implies that equal elements are not permitted.
	strict bool
	// Constraint active when selector is non-zero.
	selector air.Expr
	// Determines the largest bitwidth for which range constraints are
	// translated into AIR range constraints, versus  using a horizontal
	// bitwidth gadget.
	maxRangeConstraint uint
	// Disables the use of type proofs which exploit the limitless prover.
	// Specifically, modules with a recursive structure are created specifically
	// for the purpose of checking types.
	legacyTypeProofs bool
}

// NewLexicographicSortingGadget constructs a default sorting gadget which can
// then be configured.  The default gadget is non-strict and assumes all columns
// are ascending.
func NewLexicographicSortingGadget(prefix string, columns []uint, bitwidth uint) LexicographicSortingGadget {
	signs := make([]bool, len(columns))

	for i := range signs {
		signs[i] = true
	}
	//
	return LexicographicSortingGadget{
		prefix, columns, signs, bitwidth, false, air.NewConst64(1), 8, false}
}

// SetSigns configures the directions for all columns being sorted.
func (p *LexicographicSortingGadget) SetSigns(signs ...bool) {
	if len(p.columns) < len(signs) {
		panic("Inconsistent number of columns and signs for lexicographic sort.")
	}

	p.signs = signs
}

// SetStrict configures strictness
func (p *LexicographicSortingGadget) SetStrict(strict bool) {
	p.strict = strict
}

// SetSelector sets the selector for this constraint.
func (p *LexicographicSortingGadget) SetSelector(selector air.Expr) {
	p.selector = selector
}

// SetMaxRangeConstraint determines the cutoff for range cosntraints.
func (p *LexicographicSortingGadget) SetMaxRangeConstraint(width uint) {
	p.maxRangeConstraint = width
}

// SetLegacyTypeProofs enables or disables use of limitless type proofs.
func (p *LexicographicSortingGadget) SetLegacyTypeProofs(flag bool) {
	p.legacyTypeProofs = flag
}

// Apply this lexicographic sorting gadget to a given schema.
func (p *LexicographicSortingGadget) Apply(schema *air.Schema) {
	// Check preconditions
	// Determine enclosing module for this gadget.
	ctx := sc.ContextOfColumns(p.columns, schema)
	// Add trace computation
	deltaIndex := schema.AddAssignment(
		assignment.NewLexicographicSort(p.prefix, ctx, p.columns, p.signs, p.bitwidth))
	// Construct selecto bits.
	p.addLexicographicSelectorBits(ctx, deltaIndex, schema)
	// Construct delta terms
	constraint := constructLexicographicDeltaConstraint(deltaIndex, p.columns, p.signs)
	// Apply selector
	constraint = p.selector.Mul(constraint)
	// Add delta constraint
	deltaName := fmt.Sprintf("%s:delta", p.prefix)
	schema.AddVanishingConstraint(deltaName, 0, ctx, util.None[int](), constraint)
	// Add necessary bitwidth constraints.  Note, we don't need to consider
	// the selector here since the delta column is unique to this
	// constraint.  Furthermore, when the delta column is invalid (i.e. the
	// original source constraints are not sorted correctly), then the
	// assignment will assign zero (which is within bounds).
	gadget := NewBitwidthGadget(schema).
		WithLegacyTypeProofs(p.legacyTypeProofs).
		WithMaxRangeConstraint(p.maxRangeConstraint)
	// Apply bitwidth constraint
	gadget.Constrain(deltaIndex, p.bitwidth)
}

// Add lexicographic selector bits, including the necessary constraints.  Each
// selector bit is given a binarity constraint to ensure it is always either 1
// or 0.  A selector bit can only be set if all bits to its left are unset, and
// there is a strict difference between the two values for its colummn.
//
// NOTE: this implementation differs from the original corset which used an
// additional "Eq" bit to help ensure at most one selector bit was enabled.
func (p *LexicographicSortingGadget) addLexicographicSelectorBits(context trace.Context, deltaIndex uint,
	schema *air.Schema) {
	ncols := uint(len(p.signs))
	// Calculate column index of first selector bit
	bitIndex := deltaIndex + 1
	// Add binary constraints for selector bits
	for i := uint(0); i < ncols; i++ {
		// Add binarity constraints (i.e. to enfoce that this column is a bit).
		NewBitwidthGadget(schema).Constrain(bitIndex+i, 1)
	}
	// Apply constraints to ensure at most one is set.
	terms := make([]air.Expr, ncols)
	for i := uint(0); i < ncols; i++ {
		terms[i] = air.NewColumnAccess(bitIndex+i, 0)
		pterms := make([]air.Expr, i+1)
		qterms := make([]air.Expr, i)

		for j := uint(0); j < i; j++ {
			pterms[j] = air.NewColumnAccess(bitIndex+j, 0)
			qterms[j] = air.NewColumnAccess(bitIndex+j, 0)
		}
		// (∀j<=i.Bj=0) ==> C[k]=C[k-1]
		pterms[i] = air.NewColumnAccess(bitIndex+i, 0)
		pDiff := air.NewColumnAccess(p.columns[i], 0).Sub(air.NewColumnAccess(p.columns[i], -1))
		pName := fmt.Sprintf("%s:%d", p.prefix, i)
		schema.AddVanishingConstraint(pName, 0, context,
			util.None[int](), p.selector.Mul(air.NewConst64(1).Sub(air.Sum(pterms...)).Mul(pDiff)))
		// (∀j<i.Bj=0) ∧ Bi=1 ==> C[k]≠C[k-1]
		qDiff := Normalise(air.NewColumnAccess(p.columns[i], 0).Sub(air.NewColumnAccess(p.columns[i], -1)), schema)
		qName := fmt.Sprintf("%s:%d", p.prefix, i)
		// bi = 0 || C[k]≠C[k-1]
		constraint := air.NewColumnAccess(bitIndex+i, 0).Mul(air.NewConst64(1).Sub(qDiff))

		if i != 0 {
			// (∃j<i.Bj≠0) || bi = 0 || C[k]≠C[k-1]
			constraint = air.NewConst64(1).Sub(air.Sum(qterms...)).Mul(constraint)
		}

		schema.AddVanishingConstraint(qName, 1, context, util.None[int](), p.selector.Mul(constraint))
	}
	//
	var (
		sum        = air.Sum(terms...)
		constraint air.Expr
	)
	// Apply strictness
	if p.strict {
		// (sum = 1)
		constraint = sum.Equate(air.NewConst64(1))
	} else {
		// (sum = 0) ∨ (sum = 1)
		constraint = sum.Mul(sum.Equate(air.NewConst64(1)))
	}
	//
	name := fmt.Sprintf("%s:xor", p.prefix)
	schema.AddVanishingConstraint(name, 0, context, util.None[int](), p.selector.Mul(constraint))
}

// Construct the lexicographic delta constraint.  This states that the delta
// column either holds 0 or the difference Ci[k] - Ci[k-1] (adjusted
// appropriately for the sign) between the ith column whose multiplexor bit is
// set. This is assumes that multiplexor bits are mutually exclusive (i.e. at
// most is one).
func constructLexicographicDeltaConstraint(deltaIndex uint, columns []uint, signs []bool) air.Expr {
	ncols := uint(len(signs))
	// Calculate column index of first selector bit
	bitIndex := deltaIndex + 1
	// Construct delta terms
	terms := make([]air.Expr, ncols)
	Dk := air.NewColumnAccess(deltaIndex, 0)

	for i := uint(0); i < ncols; i++ {
		var Xdiff air.Expr
		// Ith bit column (at row k)
		Bk := air.NewColumnAccess(bitIndex+i, 0)
		// Ith column (at row k)
		Xk := air.NewColumnAccess(columns[i], 0)
		// Ith column (at row k-1)
		Xkm1 := air.NewColumnAccess(columns[i], -1)
		if signs[i] {
			Xdiff = Xk.Sub(Xkm1)
		} else {
			Xdiff = Xkm1.Sub(Xk)
		}
		// if Bk then Xdiff
		terms[i] = Bk.Mul(Xdiff)
	}
	// Construct final constraint
	return Dk.Equate(air.Sum(terms...))
}
