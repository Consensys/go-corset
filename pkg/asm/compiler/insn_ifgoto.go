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
		// For now, do the minimal thing
		code = codes[cc].(*micro.Skip)
		// Determine branch targets
		nextTarget = cc + 1
		skipTarget = cc + code.Skip + 1
	)
	//
	table.Add(nextTarget, AtomicBranch[T, E](true, code.Left, code.Right, code.Constant))
	table.Add(skipTarget, AtomicBranch[T, E](false, code.Left, code.Right, code.Constant))
	//
	return table
}
