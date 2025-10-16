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
	"github.com/consensys/go-corset/pkg/util/word"
)

// NaryFn describes a function which constructs an MIR term from a given set of zero or more terms.
type NaryFn func([]mirTerm) mirTerm

// BinaryLogicalFn describes a function whichs a logical MIR term from exactly two terms.
type BinaryLogicalFn func(l, r mirTerm) mir.LogicalTerm[word.BigEndian]

// UnconditionalTerm returns a term which has no condition associated with it.
func UnconditionalTerm(value mirTerm) IfTerm {
	var ncase = ifTermCase{ir.True[word.BigEndian, mir.LogicalTerm[word.BigEndian]](), value}
	//
	return IfTerm{[]ifTermCase{ncase}}
}

// IfThenElse constructs an IfTerm representing an if-then else expression.
func IfThenElse(cond mir.LogicalTerm[word.BigEndian], tt, ff IfTerm) IfTerm {
	var (
		n       = len(tt.cases)
		m       = len(ff.cases)
		ncases  = make([]ifTermCase, n+m)
		negCond = ir.Negation(cond)
	)
	// True branches
	for i, c := range tt.cases {
		condition := ir.Conjunction(cond, c.condition).Simplify(false)
		ncases[i] = ifTermCase{condition, c.target}
	}
	// False branches
	for i, c := range ff.cases {
		condition := ir.Conjunction(negCond, c.condition).Simplify(false)
		ncases[i+n] = ifTermCase{condition, c.target}
	}
	// Done
	return IfTerm{ncases}
}

// IfEqElse constructs an IfTerm representing an if-eq expression.
func IfEqElse(lhs IfTerm, rhs mirTerm, tt, ff mirTerm) IfTerm {
	var (
		n      = len(lhs.cases)
		ncases = make([]ifTermCase, 2*n)
	)
	// True branches
	for i, c := range lhs.cases {
		condition := ir.Conjunction(c.condition, ir.Equals[word.BigEndian, mir.LogicalTerm[word.BigEndian]](c.target, rhs))
		ncases[i] = ifTermCase{condition, tt}
	}
	// False branches
	for i, c := range lhs.cases {
		condition := ir.Conjunction(c.condition, ir.NotEquals[word.BigEndian, mir.LogicalTerm[word.BigEndian]](c.target, rhs))
		ncases[i+n] = ifTermCase{condition, ff}
	}
	// Done
	return IfTerm{ncases}
}

// MapIfTerms applies a given function to each target of the given argument
// terms, effectively producing their cross product.
func MapIfTerms(fn NaryFn, terms ...IfTerm) IfTerm {
	return IfTerm{
		constructTerms(0, terms, fn, make([]ifTermCase, len(terms))),
	}
}

// DisjunctIfTerms is similar to MapIfTerms but produces the logical disjunction
// of all constructed logical terms.
func DisjunctIfTerms(fn BinaryLogicalFn, lhs, rhs IfTerm) mir.LogicalTerm[word.BigEndian] {
	var terms []mir.LogicalTerm[word.BigEndian]
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
type IfTerm struct {
	cases []ifTermCase
}

// BitWidth returns the maximum bitwidth for any target term in this conditional
// under the given register mapping.  Specifically, the register mapping
// determines the width of registers within the term, from which the overall
// bitwidth is determined.  For example, given the term X+1 where X is u16, this
// function returns a bitwidth of 17bits.
func (p *IfTerm) BitWidth(env schema.RegisterMap) uint {
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
func (p *IfTerm) Equate(target schema.RegisterId) mir.LogicalTerm[word.BigEndian] {
	var (
		terms     = make([]mir.LogicalTerm[word.BigEndian], len(p.cases))
		targetVar = ir.NewRegisterAccess[word.BigEndian, mirTerm](target, 0)
	)
	//
	for i, c := range p.cases {
		ith := ir.Equals[word.BigEndian, mir.LogicalTerm[word.BigEndian]](targetVar, c.target)
		terms[i] = ir.Conjunction(c.condition, ith)
	}
	//
	return ir.Disjunction(terms...).Simplify(false)
}

// Map a given function over the targets of this set of conditional terms.
func (p *IfTerm) Map(fn func(mirTerm) mirTerm) IfTerm {
	var ncases = make([]ifTermCase, len(p.cases))
	//
	for i, c := range p.cases {
		ncases[i] = ifTermCase{
			c.condition,
			fn(c.target),
		}
	}
	//
	return IfTerm{ncases}
}

type ifTermCase struct {
	condition mir.LogicalTerm[word.BigEndian]
	target    mirTerm
}

func constructTerms(i int, terms []IfTerm, fn NaryFn, cases []ifTermCase) []ifTermCase {
	var ncases []ifTermCase
	//
	if i == len(terms) {
		var (
			args       = make([]mirTerm, len(cases))
			conditions = make([]mir.LogicalTerm[word.BigEndian], len(cases))
		)
		//
		for i, c := range cases {
			args[i] = c.target
			conditions[i] = c.condition
		}
		// Apply constructor
		return []ifTermCase{{ir.Conjunction(conditions...), fn(args)}}
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
