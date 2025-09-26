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
package compiler

import (
	"cmp"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
)

// BranchTable represents a sequence of zero or more branches.
type BranchTable[T any, E Expr[T, E]] struct {
	table  []Branch[T, E]
	active []bool
}

// NewBranchTable constructs a new branch table for a maximum number of branch
// targets.
func NewBranchTable[T any, E Expr[T, E]](n uint) BranchTable[T, E] {
	return BranchTable[T, E]{
		table:  make([]Branch[T, E], n),
		active: make([]bool, n),
	}
}

// Add a new branch to this branch table
func (p *BranchTable[T, E]) Add(target uint, branch Branch[T, E]) {
	//
	if p.active[target] {
		// subsequent branch to given target
		p.table[target] = p.table[target].Or(branch)
		return
	}
	// first branch to given target
	p.table[target] = branch
	p.active[target] = true
}

// Branch returns the branch associated with a given target
func (p *BranchTable[T, E]) Branch(target uint) Branch[T, E] {
	if !p.active[target] {
		panic("invalid branch target")
	}
	//
	return p.table[target]
}

// BranchTargets determines the set of active branch targets.
func (p *BranchTable[T, E]) BranchTargets() bit.Set {
	var branches bit.Set
	//
	for i, b := range p.active {
		if b {
			branches.Insert(uint(i))
		}
	}
	//
	return branches
}

// FindTarget checks whether a matching branch exists and, if so, returns the
// target of that branch.  This is useful for finding else branches, where we
// use this function to find the negation of the true branch.
func (p *BranchTable[T, E]) FindTarget(branch Branch[T, E]) (uint, bool) {
	for i, b := range p.active {
		if b && p.table[i].Equals(branch) {
			// hit
			return uint(i), true
		}
	}
	// miss
	return math.MaxUint, false
}

// ============================================================================
// Branch
// ============================================================================

// Branch abstracts the possible conditions under which a given branch
// is taken.
type Branch[T any, E Expr[T, E]] struct {
	disjuncts set.AnySortedSet[branchConjunct[T, E]]
}

// EmptyBranch constructs a new branch condition
func EmptyBranch[T any, E Expr[T, E]]() Branch[T, E] {
	return Branch[T, E]{}
}

// AtomicBranch constructs a branch from an atomic equality (or non-equality) condition.
func AtomicBranch[T any, E Expr[T, E]](sign bool, left, right io.RegisterId, constant big.Int) Branch[T, E] {
	var disjuncts set.AnySortedSet[branchConjunct[T, E]]
	//
	disjuncts.Insert(atomicConjunction[T, E](sign, left, right, constant))
	//
	return Branch[T, E]{disjuncts}
}

// Clone this branch making sure the resulting tree is disjoint
func (p *Branch[T, E]) Clone() Branch[T, E] {
	return Branch[T, E]{slices.Clone(p.disjuncts)}
}

// Equals returns true if the two branches are identical.
func (p *Branch[T, E]) Equals(other Branch[T, E]) bool {
	if len(p.disjuncts) != len(other.disjuncts) {
		return false
	}
	//
	for i := range len(p.disjuncts) {
		if p.disjuncts[i].Cmp(other.disjuncts[i]) != 0 {
			return false
		}
	}
	//
	return true
}

// And combines two paths together, such that both must be taken.  In other
// words, it computes the logical conjunction of their path conditions.
func (p *Branch[T, E]) And(other Branch[T, E]) Branch[T, E] {
	var br Branch[T, E]
	//
	for i, disjunct := range p.disjuncts {
		if i == 0 {
			br = disjunct.And(other)
		} else {
			br = br.Or(disjunct.And(other))
		}
	}
	//
	return br
}

// Or combines two paths together, such that either could be taken.  In other
// words, it computes the logical disjunction of their path conditions.
func (p *Branch[T, E]) Or(other Branch[T, E]) Branch[T, E] {
	var disjuncts set.AnySortedSet[branchConjunct[T, E]]
	//
	disjuncts.InsertSorted(&p.disjuncts)
	disjuncts.InsertSorted(&other.disjuncts)
	//
	return Branch[T, E]{disjuncts}
}

// Negate returns the logical negation of this branch (i.e. path condition).
func (p *Branch[T, E]) Negate() Branch[T, E] {
	var br Branch[T, E]
	//
	for i, d := range p.disjuncts {
		if i == 0 {
			br = d.Negate()
		} else {
			br = br.And(d.Negate())
		}
	}
	//
	return br
}

// Translate a given branch condition within the context of a given state
// reader.
func (p *Branch[T, E]) Translate(st StateReader[T, E]) E {
	var condition E
	//
	for i, c := range p.disjuncts {
		ith := c.Translate(st)
		//
		if i == 0 {
			condition = ith
		} else {
			condition = condition.Or(ith)
		}
	}
	//
	return condition
}

func (p *Branch[T, E]) String(mapping func(io.RegisterId) string) string {
	var (
		builder strings.Builder
		braces  = len(p.disjuncts) > 1
	)
	//
	for i, c := range p.disjuncts {
		if i != 0 {
			builder.WriteString(" ∨ ")
		}
		//
		builder.WriteString(c.String(braces, mapping))
	}
	//
	return builder.String()
}

// ============================================================================
// conjunct
// ============================================================================

type branchConjunct[T any, E Expr[T, E]] struct {
	conjuncts set.AnySortedSet[branchEquality[T, E]]
}

func atomicConjunction[T any, E Expr[T, E]](sign bool, left, right io.RegisterId, constant big.Int,
) branchConjunct[T, E] {
	var conjuncts set.AnySortedSet[branchEquality[T, E]]
	//
	conjuncts.Insert(branchEquality[T, E]{sign, left, right, constant})
	//
	return branchConjunct[T, E]{conjuncts}
}

// And constructs the logical conjunction of this branch and the given branch.
func (p branchConjunct[T, E]) And(o Branch[T, E]) Branch[T, E] {
	var disjuncts set.AnySortedSet[branchConjunct[T, E]]
	//
	if len(o.disjuncts) == 0 {
		panic("got here")
	}
	//
	for _, disjunct := range o.disjuncts {
		var nc branchConjunct[T, E]
		nc.conjuncts.InsertSorted(&p.conjuncts)
		nc.conjuncts.InsertSorted(&disjunct.conjuncts)
		disjuncts.Insert(nc)
	}
	// Done
	return Branch[T, E]{disjuncts}
}

// Cmp implementation for Comparable interface
func (p branchConjunct[T, E]) Cmp(o branchConjunct[T, E]) int {
	return array.Compare(p.conjuncts, o.conjuncts)
}

// Negate returns the logical negation of this conjunct.
func (p branchConjunct[T, E]) Negate() Branch[T, E] {
	var br Branch[T, E]
	//
	for i, eq := range p.conjuncts {
		if i == 0 {
			br = eq.Negate()
		} else {
			br = br.Or(eq.Negate())
		}
	}
	//
	return br
}

// Translate a given branch condition within the context of a given state
// reader.
func (p branchConjunct[T, E]) Translate(st StateReader[T, E]) E {
	var condition E
	//
	for i, c := range p.conjuncts {
		ith := c.Translate(st)
		//
		if i == 0 {
			condition = ith
		} else {
			condition = condition.And(ith)
		}
	}
	//
	return condition
}

func (p *branchConjunct[T, E]) String(braces bool, mapping func(io.RegisterId) string) string {
	var builder strings.Builder
	//
	braces = braces && len(p.conjuncts) > 1
	//
	if braces {
		builder.WriteString("(")
	}
	//
	for i, c := range p.conjuncts {
		if i != 0 {
			builder.WriteString(" ∧ ")
		}
		//
		builder.WriteString(c.String(mapping))
	}
	//
	if braces {
		builder.WriteString(")")
	}
	//
	return builder.String()
}

// ============================================================================
// equality
// ============================================================================

// branchEquality represents an atomic condition which checks for the equality (or
// non-equality) between one variable and another variable or a constant.
type branchEquality[T any, E Expr[T, E]] struct {
	// Sign indicates whether this is an equality (==) or a non-equality (!=).
	Sign bool
	// Left and right comparisons
	Left, Right io.RegisterId
	//
	Constant big.Int
}

// Cmp implementation for Comparable interface
func (p branchEquality[T, E]) Cmp(o branchEquality[T, E]) int {
	switch {
	case p.Sign && !o.Sign:
		return -1
	case !p.Sign && o.Sign:
		return 1
	case p.Left != o.Left:
		return cmp.Compare(p.Left.Unwrap(), p.Left.Unwrap())
	case p.Right != o.Right:
		return cmp.Compare(p.Right.Unwrap(), p.Right.Unwrap())
	default:
		return p.Constant.Cmp(&o.Constant)
	}
}

// Negate this equality (i.e. turn it from "==" to "!=" or vice-versa)
func (p branchEquality[T, E]) Negate() Branch[T, E] {
	return AtomicBranch[T, E](!p.Sign, p.Left, p.Right, p.Constant)
}

// Translate a given condition within the context of a given state translator.
func (p *branchEquality[T, E]) Translate(st StateReader[T, E]) E {
	var (
		left  = st.ReadRegister(p.Left)
		right E
	)
	//
	if !p.Right.IsUsed() {
		right = BigNumber[T, E](&p.Constant)
	} else {
		right = st.ReadRegister(p.Right)
	}
	//
	if p.Sign {
		return left.Equals(right)
	}
	//
	return left.NotEquals(right)
}

func (p *branchEquality[T, E]) String(mapping func(io.RegisterId) string) string {
	var (
		l = mapping(p.Left)
		r string
	)
	//
	if p.Right.IsUsed() {
		r = mapping(p.Right)
	} else {
		r = p.Constant.String()
	}
	//
	if p.Sign {
		return fmt.Sprintf("%s=%s", l, r)
	}
	//
	return fmt.Sprintf("%s≠%s", l, r)
}
