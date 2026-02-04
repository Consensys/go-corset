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
	"github.com/consensys/go-corset/pkg/asm/io/micro/dfa"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/logical"
)

// BranchCondition abstracts the possible conditions under which a given branch
// is taken.
type BranchCondition = logical.Proposition[io.RegisterId, BranchEquality]

// FALSE represents an unreachable path
var FALSE BranchCondition = logical.Truth[io.RegisterId, BranchEquality](false)

// TRUE represents an path which is always reached
var TRUE BranchCondition = logical.Truth[io.RegisterId, BranchEquality](true)

// BranchConjunction represents the conjunction of two paths
type BranchConjunction = logical.Conjunction[io.RegisterId, BranchEquality]

// BranchEquality represents an atomic branch equality
type BranchEquality = logical.Equality[io.RegisterId]

// BranchState adapts a branch condition to be an instance of dfa.State.
type BranchState struct {
	condition BranchCondition
}

// Join implementation for dfa.State interface
func (p BranchState) Join(st BranchState) BranchState {
	return BranchState{p.condition.Or(st.condition)}
}

// String implementation for dfa.State interface
func (p BranchState) String(mapping register.Map) string {
	return p.condition.String(func(rid register.Id) string {
		return mapping.Register(rid).Name()
	})
}

func constructBranchTable[T any, E Expr[T, E]](insn micro.Instruction, reader RegisterReader[T, E],
) (dfa.Result[dfa.Writes], []E) {
	//
	var (
		writes   = insn.Writes()
		branches = dfa.Construct(BranchState{TRUE}, insn.Codes, branchTableTransfer)
		table    = make([]E, len(insn.Codes))
	)
	//
	for i := 0; i < len(table); i++ {
		ith := branches.StateOf(uint(i)).condition
		table[i] = translateBranchCondition[T, E](ith, reader)
	}
	// Done
	return writes, table
}

func branchTableTransfer(offset uint, code micro.Code, state BranchState) []dfa.Transfer[BranchState] {
	var arcs []dfa.Transfer[BranchState]
	//
	switch code := code.(type) {
	case *micro.Fail, *micro.Ret, *micro.Jmp:
		return nil
	case *micro.Skip:
		// join into branch target
		return append(arcs, dfa.NewTransfer(state, offset+code.Skip+1))
	case *micro.SkipIf:
		var (
			// Determine true branch
			trueBranch = extendSkipIf(state, false, code)
			// Determine false branch
			falseBranch = extendSkipIf(state, true, code)
		)
		// join into branch target
		arcs = append(arcs, dfa.NewTransfer(trueBranch, offset+code.Skip+1))
		// join into following instruction
		return append(arcs, dfa.NewTransfer(falseBranch, offset+1))
	}
	// Transfer to following instruction
	return append(arcs, dfa.NewTransfer(state, offset+1))
}

func extendSkipIf(tail BranchState, sign bool, code *micro.SkipIf) BranchState {
	var (
		head      BranchEquality
		rightUsed = code.Right.HasFirst()
		tailc     = tail.condition
	)
	//
	switch {
	case sign && rightUsed:
		head = logical.Equals(code.Left, code.Right.First())
	case sign && !rightUsed:
		head = logical.EqualsConst(code.Left, code.Right.Second())
	case !sign && rightUsed:
		head = logical.NotEquals(code.Left, code.Right.First())
	case !sign && !rightUsed:
		head = logical.NotEqualsConst(code.Left, code.Right.Second())
	}
	// NOTE: the reason this method is needed is because we have no implicit
	// rerpesentation of logical truth or falsehood.  This means an empty path
	// does not behave in the expected manner.
	if len(tailc.Conjuncts()) == 0 {
		return BranchState{logical.NewProposition(head)}
	}
	//
	return BranchState{tailc.And(logical.NewProposition(head))}
}

// TranslateBranchCondition translates a given branch condition within the
// context of a given state reader.
func translateBranchCondition[T any, E Expr[T, E]](p BranchCondition, reader RegisterReader[T, E]) E {
	var condition E
	//
	if p.IsTrue() {
		var zero = BigNumber[T, E](big.NewInt(0))
		return zero.Equals(zero)
	} else if p.IsFalse() {
		panic("unreachable")
	}
	//
	for i, c := range p.Conjuncts() {
		ith := translateBranchConjunct[T, E](c, reader)
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
func translateBranchConjunct[T any, E Expr[T, E]](p BranchConjunction, reader RegisterReader[T, E]) E {
	var condition E
	//
	for i, atom := range p.Atoms() {
		ith := translateBranchEquality[T, E](atom, reader)
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
func translateBranchEquality[T any, E Expr[T, E]](p BranchEquality, reader RegisterReader[T, E]) E {
	var (
		left  = ReadRegister[T, E](p.Left, reader)
		right E
	)
	//
	if p.Right.HasSecond() {
		bi := p.Right.Second()
		right = BigNumber[T, E](&bi)
	} else {
		right = ReadRegister[T, E](p.Right.First(), reader)
	}
	//
	if p.Sign {
		return left.Equals(right)
	}
	//
	return left.NotEquals(right)
}

// ReadRegister constructs a suitable accessor for referring to a given register.
// This applies forwarding as appropriate.
func ReadRegister[T any, E Expr[T, E]](regId io.RegisterId, reader RegisterReader[T, E]) E {
	return reader.ReadRegister(regId, false)
}
