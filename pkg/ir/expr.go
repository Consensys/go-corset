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
package ir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// CastOf constructs a new expression which has been annotated by the user to be
// within a given range.
func CastOf[T Term[T]](arg Expr[T], bitwidth uint) Expr[T] {
	var term Term[T] = &Cast[T]{Arg: arg.Term, BitWidth: bitwidth}
	return Expr[T]{term.(T)}
}

// NewRegisterAccess constructs an AIR expression representing the value of a given
// register on the current row.
func NewRegisterAccess[T Term[T]](register uint, shift int) Expr[T] {
	var term Term[T] = &RegisterAccess[T]{Register: register, Shift: shift}
	return Expr[T]{term.(T)}
}

// Const construct an AIR expression representing a given constant.
func Const[T Term[T]](val fr.Element) Expr[T] {
	var term Term[T] = &Constant[T]{Value: val}
	return Expr[T]{term.(T)}
}

// Const64 construct an AIR expression representing a given constant from a
// uint64.
func Const64[T Term[T]](val uint64) Expr[T] {
	var (
		element         = fr.NewElement(val)
		term    Term[T] = &Constant[T]{Value: element}
	)
	//
	return Expr[T]{term.(T)}
}

// Exponent constructs a new expression representing the given argument
// raised to a given a given power.
func Exponent[T Term[T]](arg Expr[T], pow uint64) Expr[T] {
	panic("todo")
}

// If constructs a new conditional branch, where the true branch is taken when
// the condition evaluates to zero.
func If[T Term[T]](condition Expr[T], trueBranch Expr[T]) Expr[T] {
	var term Term[T] = &IfZero[T]{condition.Term, trueBranch.Term, nil}
	return Expr[T]{term.(T)}
}

// IfElse constructs a new conditional branch, where either the true branch or
// the false branch can (optionally) be VOID (but both cannot).  Note, the true
// branch is taken when the condition evaluates to zero.
func IfElse[T Term[T]](condition Expr[T], trueBranch Expr[T], falseBranch Expr[T]) Expr[T] {
	var term Term[T] = &IfZero[T]{condition.Term, trueBranch.Term, falseBranch.Term}
	return Expr[T]{term.(T)}
}

// LabelledConstant construct an expression representing a constant with a given
// label.
func LabelledConstant[T Term[T]](label string, value fr.Element) Expr[T] {
	var term Term[T] = &LabelledConst[T]{Label: label, Value: value}
	return Expr[T]{term.(T)}
}

// Normalise normalises the result of evaluating a given expression to be
// either 0 (if its value was 0) or 1 (otherwise).
func Normalise[T Term[T]](arg Expr[T]) Expr[T] {
	var term Term[T] = &Norm[T]{arg.Term}
	return Expr[T]{term.(T)}
}

// Product returns the product of zero or more multiplications.
func Product[T Term[T]](exprs ...Expr[T]) Expr[T] {
	panic("todo")
}

// Subtract returns the subtraction of the subsequent expressions from the
// first.
func Subtract[T Term[T]](exprs ...Expr[T]) Expr[T] {
	panic("todo")
}

// Sum zero or more expressions together.
func Sum[T Term[T]](exprs ...Expr[T]) Expr[T] {
	panic("todo")
}

// ============================================================================

// Expr encapsulates the notion of an "arithmetic expression".  That is
// something which is evaluated to a given value using some combination of
// arithmetic operations.
type Expr[T Term[T]] struct {
	// Term to be evaluated, etc.
	Term T
}

// AsConstant determines whether or not this is a constant expression. If so,
// the constant is returned; otherwise, nil is returned.  NOTE: this does not
// perform any form of simplification to determine this.
func (e Expr[T]) AsConstant() *fr.Element {
	panic("todo")
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Expr[T]) Bounds() util.Bounds { return e.Term.Bounds() }

// Lisp converts this schema element into a simple S-Expression, for example so
// it can be printed.
func (e Expr[T]) Lisp(module schema.Module) sexp.SExp {
	return e.Term.Lisp(module)
}

// RequiredRegisters returns the set of registers on which this expression depends.
// That is, registers whose values may be accessed when evaluating this
// expression on a given trace.
func (e Expr[T]) RequiredRegisters() *set.SortedSet[uint] {
	return e.Term.RequiredRegisters()
}

// RequiredCells returns the set of trace cells on which this expression
// depends. That is, evaluating this expression at the given row in the given
// trace will read these cells.
func (e Expr[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return e.Term.RequiredCells(row, tr)
}

// EvalAt evaluates a register access at a given row in a trace, which returns the
// value at that row of the register in question or nil is that row is
// out-of-bounds.
func (e Expr[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	return e.Term.EvalAt(k, tr)
}

// Shift all register accesses within the expression by a given amount.
func (e Expr[T]) Shift(shift int) Expr[T] {
	return Expr[T]{e.Term.ApplyShift(shift)}
}

// TestAt evaluates this expression in a given tabular context and checks it
// against zero. Observe that if this expression is *undefined* within this
// context then it returns "nil".  An expression can be undefined for
// several reasons: firstly, if it accesses a row which does not exist (e.g.
// at index -1); secondly, if it accesses a register which does not exist.
func (e Expr[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	val, err := e.Term.EvalAt(k, tr)
	//
	return val.IsZero(), 0, err
}
