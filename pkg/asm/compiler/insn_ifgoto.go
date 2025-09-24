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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

func (p *StateTranslator[F, T, E, M]) translateSkip(cc uint, codes []micro.Code) E {
	var (
		code  = codes[cc].(*micro.Skip)
		clone = p.Clone()
		lhs   = clone.translateCode(cc+1, codes)
		rhs   = p.translateCode(cc+1+code.Skip, codes)
		left  = p.ReadRegister(code.Left)
		right E
	)
	//
	if !code.Right.IsUsed() {
		right = BigNumber[T, E](&code.Constant)
	} else {
		right = p.ReadRegister(code.Right)
	}
	//
	return IfElse(left.Equals(right), lhs, rhs)
}

func (p *StateTranslator[F, T, E, M]) translateSwitch(s Switch, codes []micro.Code) E {
	var (
		targets = s.BranchTargets()
		result  E
		first   = true
	)
	//
	for iter := targets.Iter(); iter.HasNext(); {
		var (
			// Determine next branch target to consider
			target = iter.Next()
			// Translate branch target
			tmp = p.translateBranches(target, s.BranchesFor(target), codes)
		)
		//
		if first {
			result = tmp
			first = false
		} else {
			result = result.And(tmp)
		}
	}
	//
	return result
}

func (p *StateTranslator[F, T, E, M]) translateBranches(target uint, branches []Branch, codes []micro.Code) E {
	panic("todo")
}

func traverseSkips(cc uint, codes []micro.Code) Switch {
	panic("got here")
}

// Condition represents (part of) the condition for a given branch.
type Condition struct {
	// Sign indicates whether this is an equality (==) or a non-equality (!=).
	Sign bool
	// Left and right comparisons
	Left, Right io.RegisterId
	//
	Constant big.Int
}

// Branch represents the amalgamation of one or more skip statements in such a
// way that we can optimise their translation.
type Branch struct {
	// Path of conditions for this
	Condition []Condition
	// Target micro code instruction (absolute address).
	Target uint
}

// Switch represents a sequence of zero or more branches.
type Switch struct {
	Branches []Branch
}

// BranchTargets returns the set of all branch targets for this switch.
func (p *Switch) BranchTargets() bit.Set {
	var targets bit.Set
	//
	for _, b := range p.Branches {
		targets.Insert(b.Target)
	}
	//
	return targets
}

// BranchesFor determines all branches for the given branch target.
func (p *Switch) BranchesFor(target uint) []Branch {
	var branches []Branch
	//
	for _, b := range p.Branches {
		if b.Target == target {
			branches = append(branches, b)
		}
	}
	//
	return branches
}
