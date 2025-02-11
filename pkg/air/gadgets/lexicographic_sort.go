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
	"strings"

	"github.com/consensys/go-corset/pkg/air"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ApplyLexicographicSortingGadget Add sorting constraints for a sequence of one
// or more columns.  Sorting is done lexicographically starting from the
// leftmost column.  For example, consider lexicographically sorting two columns
// X and Y (in that order) in ascending (i.e. positive direction).  Then sorting
// ensures (X[k-1] < X[k]) or (X[k-1] == X[k] and Y[k-1] <= Y[k]).  The sign for
// each column determines whether its sorted into ascending (i.e. positive) or
// descending (i.e. negative) order.
//
// To implement this sort, a kind of "bit multiplexing" is used.  Specifically,
// a bit column is associated with each column being sorted, where exactly one
// of these bits can be 1.  That bit identifies the leftmost column Ci where
// Ci[k-1] < C[k].  For all columns Cj where j < i, we must have Cj[k-1] =
// Cj[k].  If all bits are zero then all columns match their previous row.
// Finally, a delta column is used in a similar fashion as for the single column
// case (see above).  The delta value captures the difference Ci[k]-Ci[k-1] to
// ensure it is positive.  The delta column is constrained to a given bitwidth,
// with constraints added as necessary to ensure this.
func ApplyLexicographicSortingGadget(columns []uint, signs []bool, bitwidth uint, schema *air.Schema) {
	ncols := len(columns)
	// Check preconditions
	if ncols != len(signs) {
		panic("Inconsistent number of columns and signs for lexicographic sort.")
	}
	// Determine enclosing module for this gadget.
	ctx := sc.ContextOfColumns(columns, schema)
	// Construct a unique prefix for this sort.
	prefix := constructLexicographicSortingPrefix(columns, signs, schema)
	// Add trace computation
	deltaIndex := schema.AddAssignment(
		assignment.NewLexicographicSort(prefix, ctx, columns, signs, bitwidth))
	// Construct selecto bits.
	addLexicographicSelectorBits(prefix, ctx, deltaIndex, columns, schema)
	// Construct delta terms
	constraint := constructLexicographicDeltaConstraint(deltaIndex, columns, signs)
	// Add delta constraint
	deltaName := fmt.Sprintf("%s:delta", prefix)
	schema.AddVanishingConstraint(deltaName, ctx, util.None[int](), constraint)
	// Add necessary bitwidth constraints
	ApplyBitwidthGadget(deltaIndex, bitwidth, schema)
}

// Construct a unique identifier for the given sort.  This should not conflict
// with the identifier for any other sort.
func constructLexicographicSortingPrefix(columns []uint, signs []bool, schema *air.Schema) string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder
	// Concatenate column names with their signs.
	for i := 0; i < len(columns); i++ {
		ith := schema.Columns().Nth(columns[i])
		id.WriteString(ith.Name)

		if signs[i] {
			id.WriteString("+")
		} else {
			id.WriteString("-")
		}
	}
	// Done
	return id.String()
}

// Add lexicographic selector bits, including the necessary constraints.  Each
// selector bit is given a binarity constraint to ensure it is always either 1
// or 0.  A selector bit can only be set if all bits to its left are unset, and
// there is a strict difference between the two values for its colummn.
//
// NOTE: this implementation differs from the original corset which used an
// additional "Eq" bit to help ensure at most one selector bit was enabled.
func addLexicographicSelectorBits(prefix string, context trace.Context,
	deltaIndex uint, columns []uint, schema *air.Schema) {
	ncols := uint(len(columns))
	// Calculate column index of first selector bit
	bitIndex := deltaIndex + 1
	// Add binary constraints for selector bits
	for i := uint(0); i < ncols; i++ {
		// Add binarity constraints (i.e. to enfoce that this column is a bit).
		ApplyBinaryGadget(bitIndex+i, schema)
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
		pDiff := air.NewColumnAccess(columns[i], 0).Sub(air.NewColumnAccess(columns[i], -1))
		pName := fmt.Sprintf("%s:%d:a", prefix, i)
		schema.AddVanishingConstraint(pName, context,
			util.None[int](), air.NewConst64(1).Sub(air.Sum(pterms...)).Mul(pDiff))
		// (∀j<i.Bj=0) ∧ Bi=1 ==> C[k]≠C[k-1]
		qDiff := Normalise(air.NewColumnAccess(columns[i], 0).Sub(air.NewColumnAccess(columns[i], -1)), schema)
		qName := fmt.Sprintf("%s:%d:b", prefix, i)
		// bi = 0 || C[k]≠C[k-1]
		constraint := air.NewColumnAccess(bitIndex+i, 0).Mul(air.NewConst64(1).Sub(qDiff))

		if i != 0 {
			// (∃j<i.Bj≠0) || bi = 0 || C[k]≠C[k-1]
			constraint = air.NewConst64(1).Sub(air.Sum(qterms...)).Mul(constraint)
		}

		schema.AddVanishingConstraint(qName, context, util.None[int](), constraint)
	}

	sum := air.Sum(terms...)
	// (sum = 0) ∨ (sum = 1)
	constraint := sum.Mul(sum.Equate(air.NewConst64(1)))
	name := fmt.Sprintf("%s:xor", prefix)
	schema.AddVanishingConstraint(name, context, util.None[int](), constraint)
}

// Construct the lexicographic delta constraint.  This states that the delta
// column either holds 0 or the difference Ci[k] - Ci[k-1] (adjusted
// appropriately for the sign) between the ith column whose multiplexor bit is
// set. This is assumes that multiplexor bits are mutually exclusive (i.e. at
// most is one).
func constructLexicographicDeltaConstraint(deltaIndex uint, columns []uint, signs []bool) air.Expr {
	ncols := uint(len(columns))
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

// AddBitArray adds an array of n bit columns using a given prefix, including
// the necessary binarity constraints.
func AddBitArray(prefix string, count int, schema *air.Schema) []uint {
	bits := make([]uint, count)

	for i := 0; i < count; i++ {
		// // Construct bit column name
		// ith := fmt.Sprintf("%s:%d", prefix, i)
		// // Add (computed) column
		// bits[i] = schema.AddColumn(ith, sc.NewUintType(1))
		// Add binarity constraints (i.e. to enfoce that this column is a bit).
		ApplyBinaryGadget(bits[i], schema)
	}
	//
	return bits
}
