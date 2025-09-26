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
		table = NewBranchTable[T, E](uint(len(codes)))
		//
		worklist worklist[T, E]
	)
	//
	worklist.push(cc, Branch[T, E]{})
	//
	for !worklist.isEmpty() {
		item := worklist.pop()
		// Check whether we have a skip, or not
		if code, ok := codes[item.pc].(*micro.Skip); ok {
			// Determine branch targets
			nextTarget := item.pc + 1
			skipTarget := item.pc + code.Skip + 1
			//
			nextBranch := item.extend(AtomicBranch[T, E](true, code.Left, code.Right, code.Constant))
			skipBranch := item.extend(AtomicBranch[T, E](false, code.Left, code.Right, code.Constant))
			//
			worklist.push(nextTarget, nextBranch)
			worklist.push(skipTarget, skipBranch)
		} else {
			// end of the road
			table.Add(item.pc, item.branch)
		}
	}
	// Done
	return table
}

// Branch path represents a single path through a nest of skip statements.
type branchPath[T any, E Expr[T, E]] struct {
	pc     uint
	branch Branch[T, E]
}

func (p branchPath[T, E]) extend(branch Branch[T, E]) Branch[T, E] {
	// NOTE: the reason this method is needed is because we have no implicit
	// rerpesentation of logical truth or falsehood.  This means an empty path
	// does not behave in the expected manner.
	if len(p.branch.disjuncts) == 0 {
		return branch
	}
	//
	return p.branch.And(branch)
}

type worklist[T any, E Expr[T, E]] struct {
	paths []branchPath[T, E]
}

func (p *worklist[T, E]) isEmpty() bool {
	return len(p.paths) == 0
}

func (p *worklist[T, E]) pop() branchPath[T, E] {
	next := p.paths[0]
	p.paths = p.paths[1:]
	//
	return next
}

func (p *worklist[T, E]) push(target uint, branch Branch[T, E]) {
	p.paths = append(p.paths, branchPath[T, E]{target, branch})
}
