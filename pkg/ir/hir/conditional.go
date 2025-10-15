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
package hir

import (
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
)

// NaryFn describes a function which constructs an MIR term from a given set of zero or more terms.
type NaryFn[F field.Element[F]] func([]mir.Term[F]) mir.Term[F]

// BinaryLogicalFn describes a function whichs a logical MIR term from exactly two terms.
type BinaryLogicalFn[F field.Element[F]] func(l, r mir.Term[F]) mir.LogicalTerm[F]

// UnconditionalTerm returns a term which has no condition associated with it.
func UnconditionalTerm[F field.Element[F]](value mir.Term[F]) IfTerm[F] {
	var ncase = ifTermCase[F]{ir.True[F, mir.LogicalTerm[F]](), value}
	//
	return IfTerm[F]{[]ifTermCase[F]{ncase}}
}

// IfThenElse constructs an IfTerm representing an if-then else expression.
func IfThenElse[F field.Element[F]](cond mir.LogicalTerm[F], tt, ff IfTerm[F]) IfTerm[F] {
	var (
		n       = len(tt.cases)
		m       = len(ff.cases)
		ncases  = make([]ifTermCase[F], n+m)
		negCond = ir.Negation(cond)
	)
	// True branches
	for i, c := range tt.cases {
		condition := ir.Conjunction(cond, c.condition).Simplify(false)
		ncases[i] = ifTermCase[F]{condition, c.target}
	}
	// False branches
	for i, c := range ff.cases {
		condition := ir.Conjunction(negCond, c.condition).Simplify(false)
		ncases[i+n] = ifTermCase[F]{condition, c.target}
	}
	// Done
	return IfTerm[F]{ncases}
}

// IfEqElse constructs an IfTerm representing an if-eq expression.
func IfEqElse[F field.Element[F]](lhs IfTerm[F], rhs mir.Term[F], tt, ff mir.Term[F]) IfTerm[F] {
	var (
		n      = len(lhs.cases)
		ncases = make([]ifTermCase[F], 2*n)
	)
	// True branches
	for i, c := range lhs.cases {
		condition := ir.Conjunction(c.condition, ir.Equals[F, mir.LogicalTerm[F]](c.target, rhs))
		ncases[i] = ifTermCase[F]{condition, tt}
	}
	// False branches
	for i, c := range lhs.cases {
		condition := ir.Conjunction(c.condition, ir.NotEquals[F, mir.LogicalTerm[F]](c.target, rhs))
		ncases[i+n] = ifTermCase[F]{condition, ff}
	}
	// Done
	return IfTerm[F]{ncases}
}

// MapIfTerms applies a given function to each target of the given argument
// terms, effectively producing their cross product.
func MapIfTerms[F field.Element[F]](fn NaryFn[F], terms ...IfTerm[F]) IfTerm[F] {
	return IfTerm[F]{
		constructTerms(0, terms, fn, make([]ifTermCase[F], len(terms))),
	}
}

// DisjunctIfTerms is similar to MapIfTerms but produces the logical disjunction
// of all constructed logical terms.
func DisjunctIfTerms[F field.Element[F]](fn BinaryLogicalFn[F], lhs, rhs IfTerm[F]) mir.LogicalTerm[F] {
	var terms []mir.LogicalTerm[F]
	//
	for _, lCase := range lhs.cases {
		for _, rCase := range rhs.cases {
			target := fn(lCase.target, rCase.target)
			terms = append(terms, ir.Conjunction(lCase.condition, rCase.condition, target))
		}
	}
	//
	return ir.Disjunction(terms...).Simplify(false)
}

// IfTerm represents a set of one or more conditional terms.  Observe that the
// conditions are expected to be total.  Hence, if there is only one term, then
// its condition must be true.
type IfTerm[F field.Element[F]] struct {
	cases []ifTermCase[F]
}

// BitWidth returns the maximum bitwidth for any target term in this conditional
// under the given register mapping.  Specifically, the register mapping
// determines the width of registers within the term, from which the overall
// bitwidth is determined.  For example, given the term X+1 where X is u16, this
// function returns a bitwidth of 17bits.
func (p *IfTerm[F]) BitWidth(env schema.RegisterMap) uint {
	var bitwidth uint
	//
	for _, c := range p.cases {
		// Determine the integer range for the given expression
		vals := c.target.ValueRange(env)
		// Extract bitwidth
		width, sign := vals.BitWidth()
		// Sanity check
		if sign {
			panic("cannot determine bitwidth of (signed) term")
		}
		//
		bitwidth = max(bitwidth, width)
	}
	//
	return bitwidth
}

// Equate returns a logical condition that constraints the target register to
// hold the values represented by this term on each row.
func (p *IfTerm[F]) Equate(target schema.RegisterId) mir.LogicalTerm[F] {
	var (
		terms     = make([]mir.LogicalTerm[F], len(p.cases))
		targetVar = ir.NewRegisterAccess[F, mir.Term[F]](target, 0)
	)
	//
	for i, c := range p.cases {
		ith := ir.Equals[F, mir.LogicalTerm[F]](targetVar, c.target)
		terms[i] = ir.Conjunction(c.condition, ith)
	}
	//
	return ir.Disjunction(terms...).Simplify(false)
}

// Map a given function over the targets of this set of conditional terms.
func (p *IfTerm[F]) Map(fn func(mir.Term[F]) mir.Term[F]) IfTerm[F] {
	var ncases = make([]ifTermCase[F], len(p.cases))
	//
	for i, c := range p.cases {
		ncases[i] = ifTermCase[F]{
			c.condition,
			fn(c.target),
		}
	}
	//
	return IfTerm[F]{ncases}
}

type ifTermCase[F any] struct {
	condition mir.LogicalTerm[F]
	target    mir.Term[F]
}

func constructTerms[F field.Element[F]](i int, terms []IfTerm[F], fn NaryFn[F], cases []ifTermCase[F]) []ifTermCase[F] {
	var ncases []ifTermCase[F]
	//
	if i == len(terms) {
		var (
			args       = make([]mir.Term[F], len(cases))
			conditions = make([]mir.LogicalTerm[F], len(cases))
		)
		//
		for i, c := range cases {
			args[i] = c.target
			conditions[i] = c.condition
		}
		// Apply constructor
		return []ifTermCase[F]{{ir.Conjunction(conditions...), fn(args)}}
	}
	//
	for _, c := range terms[i].cases {
		cases[i] = c
		// Recursively construct terms for this position
		ith := constructTerms(i+1, terms, fn, cases)
		// Append them all together
		ncases = append(ncases, ith...)
	}
	//
	return ncases
}
