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
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

func (p *StateTranslator[F, T, E, M]) translateSkip(cc uint, codes []micro.Code) E {
	// Traverse consecutive skips to determine the branch table.
	var branches = p.traverseSkips(cc, codes)
	// translate branch table
	return p.translateBranchTable(branches, codes)
}

func (p *StateTranslator[F, T, E, M]) translateBranchTable(tbl BranchTable[T, E], codes []micro.Code) E {
	var (
		targets   = tbl.BranchTargets()
		result  E = True[T, E]()
	)
	//
	for iter := targets.Iter(); iter.HasNext(); {
		var (
			trueClone  = p.Clone()
			trueTarget = iter.Next()
			trueBody   = trueClone.translateCode(trueTarget, codes)
			branch     = tbl.Branch(trueTarget)
			expr       E
		)
		// Attempt to translate as if/else
		if falseTarget, ok := tbl.FindTarget(branch.Negate()); ok {
			var (
				falseClone = p.Clone()
				falseBody  = falseClone.translateCode(falseTarget, codes)
			)
			//
			expr = IfElse(branch.Translate(p), trueBody, falseBody)
			// Remove false target from future consideration
			targets.Remove(falseTarget)
		} else {
			// If/else not possible, so translate as a standalone if.
			expr = If(branch.Translate(p), trueBody)
		}
		//
		result = result.And(expr)
	}
	//
	return result
}

func (p *StateTranslator[F, T, E, M]) traverseSkips(cc uint, codes []micro.Code) BranchTable[T, E] {
	var (
		table    = NewBranchTable[T, E](uint(len(codes)))
		branches = make([]Branch[T, E], len(codes))
		//
		worklist worklist[T, E]
	)
	//
	worklist.push(cc)
	//
	for !worklist.isEmpty() {
		//
		pc := worklist.pop()
		branch := branches[pc]
		// Check whether we have a skip, or not
		if code, ok := codes[pc].(*micro.Skip); ok {
			// Determine branch targets
			nextTarget := pc + 1
			skipTarget := pc + code.Skip + 1
			//
			nextBranch := extend(branch, AtomicBranch[T, E](true, code.Left, code.Right, code.Constant))
			skipBranch := extend(branch, AtomicBranch[T, E](false, code.Left, code.Right, code.Constant))
			//
			branches[nextTarget] = branches[nextTarget].Or(nextBranch)
			branches[skipTarget] = branches[skipTarget].Or(skipBranch)
			//
			worklist.push(nextTarget)
			worklist.push(skipTarget)
		} else {
			// end of the road
			table.Add(pc, branch)
		}
	}
	// Done
	return table
}

func extend[T any, E Expr[T, E]](tail Branch[T, E], head Branch[T, E]) Branch[T, E] {
	// NOTE: the reason this method is needed is because we have no implicit
	// rerpesentation of logical truth or falsehood.  This means an empty path
	// does not behave in the expected manner.
	if len(tail.disjuncts) == 0 {
		return head
	}
	//
	return tail.And(head)
}

type worklist[T any, E Expr[T, E]] struct {
	targets bit.Set
}

func (p *worklist[T, E]) isEmpty() bool {
	return p.targets.Count() == 0
}

func (p *worklist[T, E]) pop() uint {
	iter := p.targets.Iter()
	// calling hasNext is required for Next to work correctly.
	if !iter.HasNext() {
		panic("unreachable")
	}
	//
	next := iter.Next()
	p.targets.Remove(next)
	//
	return next
}

func (p *worklist[T, E]) push(target uint) {
	p.targets.Insert(target)
}
