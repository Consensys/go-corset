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
package ast

import (
	"errors"
	"math/big"

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
)

// Executor for (resolved) ast programs.
type Executor[S machine.State[big.Int, Instruction]] struct{}

// Execute implementation for the machine.Executor interface
func (p Executor[S]) Execute(state S) (S, error) {
	var (
		err       error
		callstack = state.CallStack()
		// Extract executing frame
		frame = callstack.Pop()
		// Identify enclosing function
		fn = state.Function(frame.Function())
		// Determine current PC position
		pc = frame.PC()
		// Lookup instruction to execute
		insn = fn.CodeAt(pc)
	)
	//
	switch insn := insn.(type) {
	case *stmt.Assign[ResolvedSymbol]:
		panic("todo assignment")
	case *stmt.IfGoto[ResolvedSymbol]:
		v := executeCondition(insn.Cond, frame)
		//
		if v {
			frame.Goto(insn.Target)
		} else {
			frame.Goto(pc + 1)
		}
	case *stmt.Goto[ResolvedSymbol]:
		frame.Goto(insn.Target)
	case *stmt.Fail[ResolvedSymbol]:
		err = errors.New("machine panic")
	case *stmt.Return[ResolvedSymbol]:
		return state, nil
	default:
		panic("unknown instruction encountered")
	}
	//
	callstack.Push(frame)
	//
	return state, err
}

func executeCondition(cond expr.Condition, frame machine.Frame[big.Int]) bool {
	switch c := cond.(type) {
	case *expr.Cmp:
		return executeCmp(*c, frame)
	default:
		panic("unknown condition")
	}
}

func executeCmp(cond expr.Cmp, frame machine.Frame[big.Int]) bool {
	var (
		lhs = executeExpression(cond.Left, frame)
		rhs = executeExpression(cond.Right, frame)
		cmp = lhs.Cmp(&rhs)
	)
	//
	switch cond.Operator {
	case expr.EQ:
		return cmp == 0
	case expr.NEQ:
		return cmp != 0
	case expr.GT:
		return cmp > 0
	case expr.GTEQ:
		return cmp >= 0
	case expr.LTEQ:
		return cmp <= 0
	case expr.LT:
		return cmp < 0
	default:
		panic("unreachable")
	}
}

func executeExpression(e expr.Expr, frame machine.Frame[big.Int]) big.Int {
	switch e := e.(type) {
	case *expr.Add:
		return add(e.Exprs, frame)
	case *expr.ConstAccess:
		panic("todo")
	case *expr.Const:
		return e.Constant
	case *expr.Mul:
		return multiply(e.Exprs, frame)
	case *expr.Sub:
		return subtract(e.Exprs, frame)
	case *expr.VarAccess:
		return frame.Load(e.Variable)
	default:
		panic("unreachable")
	}
}

func add(es []expr.Expr, frame machine.Frame[big.Int]) big.Int {
	var result big.Int
	//
	for i, e := range es {
		ith := executeExpression(e, frame)

		if i == 0 {
			result.Set(&ith)
		} else {
			result.Add(&result, &ith)
		}
	}
	// Done
	return result
}

func multiply(es []expr.Expr, frame machine.Frame[big.Int]) big.Int {
	var result big.Int
	//
	for i, e := range es {
		ith := executeExpression(e, frame)

		if i == 0 {
			result.Set(&ith)
		} else {
			result.Mul(&result, &ith)
		}
	}
	// Done
	return result
}

func subtract(es []expr.Expr, frame machine.Frame[big.Int]) big.Int {
	var result big.Int
	//
	for i, e := range es {
		ith := executeExpression(e, frame)

		if i == 0 {
			result.Set(&ith)
		} else {
			result.Sub(&result, &ith)
		}
	}
	// Done
	return result
}
