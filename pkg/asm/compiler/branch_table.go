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
	"fmt"
	"math"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/logical"
)

// BranchCondition abstracts the possible conditions under which a given branch
// is taken.
type BranchCondition = logical.Proposition[io.RegisterId, BranchEquality]

// BranchConjunction represents the conjunction of two paths
type BranchConjunction = logical.Conjunction[io.RegisterId, BranchEquality]

// BranchEquality represents an atomic branch equality
type BranchEquality = logical.Equality[io.RegisterId]

// BranchTable represents a sequence of zero or more branches.
type BranchTable[T any, E Expr[T, E]] struct {
	table  []BranchCondition
	active []bool
}

// NewBranchTable constructs a new branch table for a maximum number of branch
// targets.
func NewBranchTable[T any, E Expr[T, E]](n uint) BranchTable[T, E] {
	return BranchTable[T, E]{
		table:  make([]BranchCondition, n),
		active: make([]bool, n),
	}
}

// Add a new branch to this branch table
func (p *BranchTable[T, E]) Add(target uint, branch BranchCondition) {
	if branch.IsFalse() {
		return
	} else if p.active[target] {
		// subsequent branch to given target
		p.table[target] = p.table[target].Or(branch)
		return
	}
	// first branch to given target
	p.table[target] = branch
	p.active[target] = true
}

// Branch returns the branch associated with a given target
func (p *BranchTable[T, E]) Branch(target uint) BranchCondition {
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
func (p *BranchTable[T, E]) FindTarget(branch BranchCondition) (uint, bool) {
	for i, b := range p.active {
		if b && p.table[i].Equals(branch) {
			// hit
			return uint(i), true
		}
	}
	// miss
	return math.MaxUint, false
}

func (p *BranchTable[T, E]) String(mapping func(io.RegisterId) string) string {
	var (
		builder strings.Builder
		first   bool = true
	)
	//
	builder.WriteString("[")
	//
	for i, branch := range p.table {
		if p.active[i] {
			if !first {
				builder.WriteString("; ")
			}
			//
			first = false
			//
			builder.WriteString("(")
			builder.WriteString(branch.String(mapping))
			builder.WriteString(fmt.Sprintf(")=>%d", i))
		}
	}
	//
	builder.WriteString("]")
	//
	return builder.String()
}

// ============================================================================
// Translation
// ============================================================================

// TranslateBranchCondition translates a given branch condition within the
// context of a given state reader.
func TranslateBranchCondition[T any, E Expr[T, E]](p BranchCondition, st StateReader[T, E]) E {
	var condition E
	//
	for i, c := range p.Conjuncts() {
		ith := translateBranchConjunct(c, st)
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

// Translate a given branch condition within the context of a given state
// reader.
func translateBranchConjunct[T any, E Expr[T, E]](p BranchConjunction, st StateReader[T, E]) E {
	var condition E
	//
	for i, atom := range p.Atoms() {
		ith := translateBranchEquality(atom, st)
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

// Translate a given condition within the context of a given state translator.
func translateBranchEquality[T any, E Expr[T, E]](p BranchEquality, st StateReader[T, E]) E {
	var (
		left  = st.ReadRegister(p.Left)
		right E
	)
	//
	if p.Right.HasSecond() {
		bi := p.Right.Second()
		right = BigNumber[T, E](&bi)
	} else {
		right = st.ReadRegister(p.Right.First())
	}
	//
	if p.Sign {
		return left.Equals(right)
	}
	//
	return left.NotEquals(right)
}
