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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/asm/io/micro/dfa"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/logical"
)

// BranchGroupId represents a register ID which can additionally indicate
// whether forwarding is active or not.  Forwarding indicates that the register
// was previously assigned in the given micro instruction and, hence, needs to
// be "forwarded" to the point where its used.
type BranchGroupId struct {
	// First underlying register in group
	id register.Id
	// Number of registers in group
	n uint
	// Indication of whether forwarding is active or not.
	forwarding bool
}

func newGroupId(vec register.Vector, forwarding bool) BranchGroupId {
	var first = vec.Registers()[0].Unwrap()
	// Sanity check all registers in the vector are allocated in the expected
	// order (i.e. consecutively, starting from the least significant limb).
	for i := range len(vec.Registers()) {
		expected := register.NewId(first + uint(i))
		//
		if vec.Registers()[i] != expected {
			panic("invalid register group")
		}
	}
	//
	return BranchGroupId{
		vec.Registers()[0],
		uint(len(vec.Registers())),
		forwarding,
	}
}

// Cmp implementation of the logical.Variable interface
func (p BranchGroupId) Cmp(o BranchGroupId) int {
	if p.forwarding == o.forwarding {
		if c := p.id.Cmp(o.id); c != 0 {
			return c
		}
		//
		return cmp.Compare(p.n, o.n)
	} else if p.forwarding {
		return 1
	}
	//
	return -1
}

// String implementation of the logical.Variable interface
func (p BranchGroupId) String() string {
	var (
		first = p.id.Unwrap()
		last  = first + p.n - 1
		id    = fmt.Sprintf("{%d...%d}", first, last)
	)
	//
	if p.forwarding {
		return id
	}
	//
	return fmt.Sprintf("'%s", id)
}

// BranchCondition abstracts the possible conditions under which a given branch
// is taken.
type BranchCondition = logical.Proposition[BranchGroupId, BranchEquality]

// FALSE represents an unreachable path
var FALSE BranchCondition = logical.Truth[BranchGroupId, BranchEquality](false)

// TRUE represents an path which is always reached
var TRUE BranchCondition = logical.Truth[BranchGroupId, BranchEquality](true)

// BranchConjunction represents the conjunction of two paths
type BranchConjunction = logical.Conjunction[BranchGroupId, BranchEquality]

// BranchEquality represents an atomic branch equality
type BranchEquality = logical.Equality[BranchGroupId]

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
	return p.condition.String(func(rid BranchGroupId) string {
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
		left      = newGroupId(code.Left, writes.MayAnybeAssigned(code.Left.Registers()))
	)
	//
	switch {
	case !sign && rightUsed:
		right := code.Right.First()
		head = logical.Equals(left, newGroupId(right, writes.MayAnybeAssigned(right.Registers())))
	case !sign && !rightUsed:
		head = logical.EqualsConst(left, code.Right.Second())
	case sign && rightUsed:
		right := code.Right.First()
		head = logical.NotEquals(left, newGroupId(right, writes.MayAnybeAssigned(right.Registers())))
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
		zero  = BigNumber[T, E](big.NewInt(0))
		left  = ReadRegisters[T, E](p.Left, reader)
		right []E
	)
	//
	if p.Right.HasSecond() {
		// bi := p.Right.Second()
		// right = BigNumber[T, E](&bi)
		panic("got here")
	} else {
		right = ReadRegisters[T, E](p.Right.First(), reader)
	}
	// pad out with zeros
	n := uint(max(len(left), len(right)))
	left = array.BackPad(left, n, zero)
	right = array.BackPad(right, n, zero)
	//
	if p.Sign {
		var condition E
		//
		for i := range n {
			if i == 0 {
				condition = left[i].Equals(right[i])
			} else {
				condition = condition.And(left[i].Equals(right[i]))
			}
		}
		//
		return condition
	}
	//
	var condition E
	//
	for i := range n {
		if i == 0 {
			condition = left[i].NotEquals(right[i])
		} else {
			condition = condition.Or(left[i].NotEquals(right[i]))
		}
	}
	//
	return condition
}

// ReadRegisters constructs a suitable accessor for referring to a given set of
// registers. This applies forwarding as appropriate.
func ReadRegisters[T any, E Expr[T, E]](reg BranchGroupId, reader RegisterReader[T, E]) []E {
	var reads = make([]E, reg.n)
	//
	for i := range reg.n {
		ith := register.NewId(reg.id.Unwrap() + i)
		reads[i] = reader.ReadRegister(ith, reg.forwarding)
	}
	//
	return reads
}
