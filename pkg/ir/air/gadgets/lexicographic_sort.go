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

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/assignment"
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
	selector air.Term
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
	return LexicographicSortingGadget{prefix, columns, signs, bitwidth, false, ir.Const64[air.Term](1)}
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
func (p *LexicographicSortingGadget) SetSelector(selector air.Term) {
	p.selector = selector
}

// Apply this lexicographic sorting gadget to a given schema.
func (p *LexicographicSortingGadget) Apply(module *air.ModuleBuilder) {
	deltaName := fmt.Sprintf("%s:delta", p.prefix)
	// Determine enclosing module for this gadget.
	ctx := trace.NewContext(module.Id(), 1)
	// Look up register
	deltaIndex, ok := module.HasRegister(deltaName)
	// Add new column (if it does not already exist)
	if !ok {
		// Allocate registers
		regs := assignment.LexicographicSortRegisters(uint(len(p.signs)), p.prefix, p.bitwidth)
		targets := make([]uint, len(regs))
		//
		for i, r := range regs {
			targets[i] = module.NewRegister(r)
		}
		// Extract delta index
		deltaIndex = targets[0]
		//
		module.AddAssignment(
			assignment.NewLexicographicSort(ctx, targets, p.signs, p.columns, p.bitwidth))
		// Construct selector bits.
		p.addLexicographicSelectorBits(ctx, deltaIndex, module)
		// Add necessary bitwidth constraints
		ApplyBitwidthGadget(deltaIndex, p.bitwidth, p.selector, module)
	}
	// Construct delta terms
	constraint := constructLexicographicDeltaConstraint(deltaIndex, p.columns, p.signs)
	// Apply selector
	constraint = ir.Product(p.selector, constraint)
	// Add delta constraint
	module.AddConstraint(
		air.NewVanishingConstraint(deltaName, ctx, util.None[int](), constraint))
}

// Add lexicographic selector bits, including the necessary constraints.  Each
// selector bit is given a binarity constraint to ensure it is always either 1
// or 0.  A selector bit can only be set if all bits to its left are unset, and
// there is a strict difference between the two values for its colummn.
//
// NOTE: this implementation differs from the original corset which used an
// additional "Eq" bit to help ensure at most one selector bit was enabled.
func (p *LexicographicSortingGadget) addLexicographicSelectorBits(context trace.Context, deltaIndex uint,
	schema *air.ModuleBuilder) {
	var (
		one   = ir.Const64[air.Term](1)
		ncols = uint(len(p.signs))
		// Calculate column index of first selector bit
		bitIndex = deltaIndex + 1
	)
	// Add binary constraints for selector bits
	for i := uint(0); i < ncols; i++ {
		// Add binarity constraints (i.e. to enfoce that this column is a bit).
		ApplyBinaryGadget(bitIndex+i, context, schema)
	}
	// Apply constraints to ensure at most one is set.
	terms := make([]air.Term, ncols)

	for i := uint(0); i < ncols; i++ {
		pterms := make([]air.Term, i+1)
		qterms := make([]air.Term, i)
		c_i := ir.NewRegisterAccess[air.Term](p.columns[i], 0)
		c_pi := ir.NewRegisterAccess[air.Term](p.columns[i], -1)
		terms[i] = ir.NewRegisterAccess[air.Term](bitIndex+i, 0)

		for j := uint(0); j < i; j++ {
			pterms[j] = ir.NewRegisterAccess[air.Term](bitIndex+j, 0)
			qterms[j] = ir.NewRegisterAccess[air.Term](bitIndex+j, 0)
		}
		// (∀j<=i.Bj=0) ==> C[k]=C[k-1]
		pterms[i] = ir.NewRegisterAccess[air.Term](bitIndex+i, 0)
		pDiff := ir.Subtract(c_i, c_pi)
		pName := fmt.Sprintf("%s:%d", p.prefix, i)
		schema.AddConstraint(
			air.NewVanishingConstraint(pName, context,
				util.None[int](), ir.Product(p.selector, ir.Subtract(one, ir.Sum(pterms...)), pDiff)))
		// (∀j<i.Bj=0) ∧ Bi=1 ==> C[k]≠C[k-1]
		qDiff := Normalise(ir.Subtract(c_i, c_pi), context, schema)
		qName := fmt.Sprintf("%s:%d", p.prefix, i)
		// bi = 0 || C[k]≠C[k-1]
		constraint := ir.Product(pterms[i], ir.Subtract(one, qDiff))
		if i != 0 {
			// (∃j<i.Bj≠0) || bi = 0 || C[k]≠C[k-1]
			constraint = ir.Product(ir.Subtract(one, ir.Sum(qterms...)), constraint)
		}

		schema.AddConstraint(
			air.NewVanishingConstraint(qName, context, util.None[int](), ir.Product(p.selector, constraint)))
	}
	//
	var (
		sum        = ir.Sum(terms...)
		constraint air.Term
	)
	// Apply strictness
	if p.strict {
		// (sum = 1)
		constraint = ir.Subtract(sum, one)
	} else {
		// (sum = 0) ∨ (sum = 1)
		constraint = ir.Product(sum, ir.Subtract(sum, one))
	}
	//
	name := fmt.Sprintf("%s:xor", p.prefix)
	schema.AddConstraint(
		air.NewVanishingConstraint(name, context, util.None[int](), ir.Product(p.selector, constraint)))
}

// Construct the lexicographic delta constraint.  This states that the delta
// column either holds 0 or the difference Ci[k] - Ci[k-1] (adjusted
// appropriately for the sign) between the ith column whose multiplexor bit is
// set. This is assumes that multiplexor bits are mutually exclusive (i.e. at
// most is one).
func constructLexicographicDeltaConstraint(deltaIndex uint, columns []uint, signs []bool) air.Term {
	ncols := uint(len(signs))
	// Calculate column index of first selector bit
	bitIndex := deltaIndex + 1
	// Construct delta terms
	terms := make([]air.Term, ncols)
	Dk := ir.NewRegisterAccess[air.Term](deltaIndex, 0)
	//
	for i := uint(0); i < ncols; i++ {
		var Xdiff air.Term
		// Ith bit column (at row k)
		Bk := ir.NewRegisterAccess[air.Term](bitIndex+i, 0)
		// Ith column (at row k)
		Xk := ir.NewRegisterAccess[air.Term](columns[i], 0)
		// Ith column (at row k-1)
		Xkm1 := ir.NewRegisterAccess[air.Term](columns[i], -1)
		if signs[i] {
			Xdiff = ir.Subtract(Xk, Xkm1)
		} else {
			Xdiff = ir.Subtract(Xkm1, Xk)
		}
		// if Bk then Xdiff
		terms[i] = ir.Product(Bk, Xdiff)
	}
	// Construct final constraint
	return ir.Subtract(Dk, ir.Sum(terms...))
}
