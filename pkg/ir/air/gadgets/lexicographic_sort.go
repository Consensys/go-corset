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
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
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
type LexicographicSortingGadget[F field.Element[F]] struct {
	// Prefix is used to construct the delta column name.
	prefix string
	// Identifies column(s) being sorted
	columns []sc.RegisterId
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
	selector air.Term[F]
	// Determines the largest bitwidth for which range constraints are
	// translated into AIR range constraints, versus  using a horizontal
	// bitwidth gadget.
	maxRangeConstraint uint
	// Enables the use of type proofs which exploit the
	// limitless prover. Specifically, modules with a recursive structure are
	// created specifically for the purpose of checking types.
	limitless bool
}

// NewLexicographicSortingGadget constructs a default sorting gadget which can
// then be configured.  The default gadget is non-strict and assumes all columns
// are ascending.
func NewLexicographicSortingGadget[F field.Element[F]](prefix string, columns []sc.RegisterId, bitwidth uint,
) *LexicographicSortingGadget[F] {
	//
	signs := make([]bool, len(columns))

	for i := range signs {
		signs[i] = true
	}
	//
	return &LexicographicSortingGadget[F]{prefix, columns, signs, bitwidth, false,
		ir.Const64[F, air.Term[F]](1), 8, false}
}

// WithSigns configures the directions for all columns being sorted.
func (p *LexicographicSortingGadget[F]) WithSigns(signs ...bool) *LexicographicSortingGadget[F] {
	if len(p.columns) < len(signs) {
		panic("Inconsistent number of columns and signs for lexicographic sort.")
	}

	p.signs = signs
	//
	return p
}

// WithStrictness configures strictness
func (p *LexicographicSortingGadget[F]) WithStrictness(strict bool) *LexicographicSortingGadget[F] {
	p.strict = strict
	return p
}

// WithSelector sets the selector for this constraint.
func (p *LexicographicSortingGadget[F]) WithSelector(selector air.Term[F]) *LexicographicSortingGadget[F] {
	p.selector = selector
	return p
}

// WithMaxRangeConstraint determines the cutoff for range cosntraints.
func (p *LexicographicSortingGadget[F]) WithMaxRangeConstraint(width uint) *LexicographicSortingGadget[F] {
	p.maxRangeConstraint = width
	return p
}

// WithLimitless enables or disables use of limitless type proofs.
func (p *LexicographicSortingGadget[F]) WithLimitless(flag bool) *LexicographicSortingGadget[F] {
	p.limitless = flag
	return p
}

// Apply this lexicographic sorting gadget to a given schema.
func (p *LexicographicSortingGadget[F]) Apply(mid sc.ModuleId, schema *air.SchemaBuilder[F]) {
	var (
		module = schema.Module(mid)
		//
		deltaName = fmt.Sprintf("%s:delta", p.prefix)
		// Look up register
		deltaIndex, ok = module.HasRegister(deltaName)
	)
	// Add new column (if it does not already exist)
	if !ok {
		// Allocate registers
		var (
			regs    = assignment.LexicographicSortRegisters(uint(len(p.signs)), p.prefix, p.bitwidth)
			sources = make([]sc.RegisterRef, len(p.columns))
			targets = make([]sc.RegisterRef, len(regs))
		)
		// Construct source refs
		for i, rid := range p.columns {
			sources[i] = sc.NewRegisterRef(mid, rid)
		}
		// Construct target refs
		for i, r := range regs {
			targets[i] = sc.NewRegisterRef(mid, module.NewRegister(r))
		}
		// Extract delta index
		deltaIndex = targets[0].Register()
		//
		module.AddAssignment(
			assignment.NewLexicographicSort[F](targets, p.signs, sources, p.bitwidth))
		// Construct selector bits.
		p.addLexicographicSelectorBits(deltaIndex, mid, schema)
		// Add necessary bitwidth constraints.  Note, we don't need to consider
		// the selector here since the delta column is unique to this
		// constraint.  Furthermore, when the delta column is invalid (i.e. the
		// original source constraints are not sorted correctly), then the
		// assignment will assign zero (which is within bounds).
		ref := sc.NewRegisterRef(mid, deltaIndex)
		// Constrict gadget
		gadget := NewBitwidthGadget(schema).
			WithLimitless(p.limitless).
			WithMaxRangeConstraint(p.maxRangeConstraint)
		// Apply bitwidth constraint
		gadget.Constrain(ref, p.bitwidth)
	}
	// Construct delta terms
	constraint := constructLexicographicDeltaConstraint[F](deltaIndex, p.columns, p.signs)
	// Apply selector
	constraint = ir.Product(p.selector, constraint)
	// Add delta constraint
	module.AddConstraint(
		air.NewVanishingConstraint(deltaName, module.Id(), util.None[int](), constraint))
}

// Add lexicographic selector bits, including the necessary constraints.  Each
// selector bit is given a binarity constraint to ensure it is always either 1
// or 0.  A selector bit can only be set if all bits to its left are unset, and
// there is a strict difference between the two values for its colummn.
//
// NOTE: this implementation differs from the original corset which used an
// additional "Eq" bit to help ensure at most one selector bit was enabled.
func (p *LexicographicSortingGadget[F]) addLexicographicSelectorBits(deltaIndex sc.RegisterId, mid sc.ModuleId,
	schema *air.SchemaBuilder[F]) {
	//
	var (
		module = schema.Module(mid)
		one    = ir.Const64[F, air.Term[F]](1)
		ncols  = uint(len(p.signs))
		// Calculate column index of first selector bit
		bitIndex = deltaIndex.Unwrap() + 1
	)
	// Add binary constraints for selector bits
	for i := uint(0); i < ncols; i++ {
		ref := sc.NewRegisterRef(mid, sc.NewRegisterId(bitIndex+i))
		// Add binarity constraints (i.e. to enforce that this column is a bit).
		NewBitwidthGadget(schema).Constrain(ref, 1)
	}
	// Apply constraints to ensure at most one is set.
	terms := make([]air.Term[F], ncols)

	for i := uint(0); i < ncols; i++ {
		ith_id := sc.NewRegisterId(bitIndex + i)
		pterms := make([]air.Term[F], i+1)
		qterms := make([]air.Term[F], i)
		c_i := ir.NewRegisterAccess[F, air.Term[F]](p.columns[i], 0)
		c_pi := ir.NewRegisterAccess[F, air.Term[F]](p.columns[i], -1)
		terms[i] = ir.NewRegisterAccess[F, air.Term[F]](ith_id, 0)

		for j := uint(0); j < i; j++ {
			jth_id := sc.NewRegisterId(bitIndex + j)
			pterms[j] = ir.NewRegisterAccess[F, air.Term[F]](jth_id, 0)
			qterms[j] = ir.NewRegisterAccess[F, air.Term[F]](jth_id, 0)
		}
		// (∀j<=i.Bj=0) ==> C[k]=C[k-1]
		pterms[i] = ir.NewRegisterAccess[F, air.Term[F]](ith_id, 0)
		pDiff := ir.Subtract(c_i, c_pi)
		pName := fmt.Sprintf("%s:%d", p.prefix, i)
		module.AddConstraint(
			air.NewVanishingConstraint(pName, module.Id(),
				util.None[int](), ir.Product(p.selector, ir.Subtract(one, ir.Sum(pterms...)), pDiff)))
		// (∀j<i.Bj=0) ∧ Bi=1 ==> C[k]≠C[k-1]
		qDiff := Normalise(ir.Subtract(c_i, c_pi), module)
		qName := fmt.Sprintf("%s:%d", p.prefix, i)
		// bi = 0 || C[k]≠C[k-1]
		constraint := ir.Product(pterms[i], ir.Subtract(one, qDiff))
		if i != 0 {
			// (∃j<i.Bj≠0) || bi = 0 || C[k]≠C[k-1]
			constraint = ir.Product(ir.Subtract(one, ir.Sum(qterms...)), constraint)
		}

		module.AddConstraint(
			air.NewVanishingConstraint(qName, module.Id(), util.None[int](), ir.Product(p.selector, constraint)))
	}
	//
	var (
		sum        = ir.Sum(terms...)
		constraint air.Term[F]
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
	module.AddConstraint(
		air.NewVanishingConstraint(name, module.Id(), util.None[int](), ir.Product(p.selector, constraint)))
}

// Construct the lexicographic delta constraint.  This states that the delta
// column either holds 0 or the difference Ci[k] - Ci[k-1] (adjusted
// appropriately for the sign) between the ith column whose multiplexer bit is
// set. This is assumes that multiplexer bits are mutually exclusive (i.e. at
// most is one).
func constructLexicographicDeltaConstraint[F field.Element[F]](deltaIndex sc.RegisterId, columns []sc.RegisterId,
	signs []bool) air.Term[F] {
	//
	var (
		ncols = uint(len(signs))
		// Calculate column index of first selector bit
		bitIndex = deltaIndex.Unwrap() + 1
		// Construct delta terms
		terms = make([]air.Term[F], ncols)
		Dk    = ir.NewRegisterAccess[F, air.Term[F]](deltaIndex, 0)
	)
	//
	for i := uint(0); i < ncols; i++ {
		var (
			Xdiff   air.Term[F]
			ith_bit = sc.NewRegisterId(bitIndex + i)
		)
		// Ith bit column (at row k)
		Bk := ir.NewRegisterAccess[F, air.Term[F]](ith_bit, 0)
		// Ith column (at row k)
		Xk := ir.NewRegisterAccess[F, air.Term[F]](columns[i], 0)
		// Ith column (at row k-1)
		Xkm1 := ir.NewRegisterAccess[F, air.Term[F]](columns[i], -1)
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
