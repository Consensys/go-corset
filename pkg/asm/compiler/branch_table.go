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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/asm/io/micro/dfa"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/logical"
)

// BranchRegisterId represents a register ID which can additionally indicate
// whether forwarding is active or not.  Forwarding indicates that the register
// was previously assigned in the given micro instruction and, hence, needs to
// be "forwarded" to the point where its used.
type BranchRegisterId struct {
	// Underlying register ID
	id register.Id
	// Indication of whether forwarding is active or not.
	forwarding bool
}

// Cmp implementation of the logical.Variable interface
func (p BranchRegisterId) Cmp(o BranchRegisterId) int {
	if p.forwarding == o.forwarding {
		return p.id.Cmp(o.id)
	} else if p.forwarding {
		return 1
	}
	//
	return -1
}

// String implementation of the logical.Variable interface
func (p BranchRegisterId) String() string {
	if p.forwarding {
		return p.id.String()
	}
	//
	return fmt.Sprintf("'%s", p.id.String())
}

// BranchCondition abstracts the possible conditions under which a given branch
// is taken.
type BranchCondition = logical.Proposition[BranchRegisterId, BranchEquality]

// FALSE represents an unreachable path
var FALSE BranchCondition = logical.Truth[BranchRegisterId, BranchEquality](false)

// TRUE represents an path which is always reached
var TRUE BranchCondition = logical.Truth[BranchRegisterId, BranchEquality](true)

// BranchConjunction represents the conjunction of two paths
type BranchConjunction = logical.Conjunction[BranchRegisterId, BranchEquality]

// BranchEquality represents an atomic branch equality
type BranchEquality = logical.Equality[BranchRegisterId]

// BranchTransferFunction represents a transfer function over branch state.
type BranchTransferFunction func(offset uint, code micro.Code, state BranchState) []dfa.Transfer[BranchState]

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
	return p.condition.String(func(rid BranchRegisterId) string {
		var name = mapping.Register(rid.id).Name()
		//
		if rid.forwarding {
			return name
		}
		//
		return fmt.Sprintf("'%s", name)
	})
}

func constructBranchTable[T any, E Expr[T, E]](insn micro.Instruction, reader RegisterReader[T, E],
) (dfa.Result[dfa.Writes], []BranchCondition) {
	//
	var (
		writes   = insn.Writes()
		branches = dfa.Construct(BranchState{TRUE}, insn.Codes, branchTableTransfer(writes))
		table    = make([]BranchCondition, len(insn.Codes))
	)
	//
	for i := 0; i < len(table); i++ {
		table[i] = branches.StateOf(uint(i)).condition
	}
	// Done
	return writes, table
}

func branchTableTransfer(writeMap dfa.Result[dfa.Writes]) BranchTransferFunction {
	return func(offset uint, code micro.Code, state BranchState) []dfa.Transfer[BranchState] {
		var (
			arcs   []dfa.Transfer[BranchState]
			writes = writeMap.StateOf(offset)
		)
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
				trueBranch = extendSkipIf(state, true, code, writes)
				// Determine false branch
				falseBranch = extendSkipIf(state, false, code, writes)
			)
			// join into branch target
			arcs = append(arcs, dfa.NewTransfer(trueBranch, offset+code.Skip+1))
			// join into following instruction
			return append(arcs, dfa.NewTransfer(falseBranch, offset+1))
		}
		// Transfer to following instruction
		return append(arcs, dfa.NewTransfer(state, offset+1))
	}
}

func extendSkipIf(tail BranchState, sign bool, code *micro.SkipIf, writes dfa.Writes) BranchState {
	var (
		head      BranchEquality
		rightUsed = code.Right.HasFirst()
		tailc     = tail.condition
		left      = BranchRegisterId{code.Left, writes.MaybeAssigned(code.Left)}
	)
	//
	switch {
	case !sign && rightUsed:
		right := code.Right.First()
		head = logical.Equals(left, BranchRegisterId{right, writes.MaybeAssigned(right)})
	case !sign && !rightUsed:
		head = logical.EqualsConst(left, code.Right.Second())
	case sign && rightUsed:
		right := code.Right.First()
		head = logical.NotEquals(left, BranchRegisterId{right, writes.MaybeAssigned(right)})
	case sign && !rightUsed:
		head = logical.NotEqualsConst(left, code.Right.Second())
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
func ReadRegister[T any, E Expr[T, E]](reg BranchRegisterId, reader RegisterReader[T, E]) E {
	return reader.ReadRegister(reg.id, reg.forwarding)
}
