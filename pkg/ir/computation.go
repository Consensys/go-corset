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
	"fmt"

	"github.com/consensys/go-corset/pkg/util/field"
)

// Computation represents an "unbound" term.  That is, it captures any possible
// term (i.e. rather than a fixed set as for MIR or AIR, etc).
type Computation[F any] interface {
	Term[F, Computation[F]]
}

// LogicalComputation represents an "unbound" term.  That is, it captures any
// possible term (i.e. rather than a fixed set as for MIR or AIR, etc).
type LogicalComputation[F any] interface {
	LogicalTerm[F, LogicalComputation[F]]
}

// ============================================================================
// Computation
// ============================================================================

// NewComputation takes an arbitrary term and converts in into an instance of
// computation by wrapping it.
func NewComputation[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](term Term[F, T]) Computation[F] {
	switch t := term.(type) {
	case *Add[F, T]:
		args := NewComputations[F, S](t.Args)
		return Sum(args...)
	case *Cast[F, T]:
		arg := NewComputation[F, S](t.Arg)
		return CastOf(arg, t.BitWidth)
	case *Constant[F, T]:
		return Const[F, Computation[F]](t.Value)
	case *Exp[F, T]:
		arg := NewComputation[F, S](t.Arg)
		return Exponent(arg, t.Pow)
	case *IfZero[F, S, T]:
		condition := NewLogicalComputation[F, S, T](t.Condition)
		trueBranch := NewComputation[F, S](t.TrueBranch)
		falseBranch := NewComputation[F, S](t.FalseBranch)
		// Done
		return IfElse(condition, trueBranch, falseBranch)
	case *LabelledConst[F, T]:
		return LabelledConstant[F, Computation[F]](t.Label, t.Value)
	case *Mul[F, T]:
		args := NewComputations[F, S](t.Args)
		return Product(args...)
	case *Norm[F, T]:
		arg := NewComputation[F, S, T](t.Arg)
		return Normalise(arg)
	case *RegisterAccess[F, T]:
		return NewRegisterAccess[F, Computation[F]](t.Register, t.Shift)
	case *Sub[F, T]:
		args := NewComputations[F, S](t.Args)
		return Subtract(args...)
	case *VectorAccess[F, T]:
		var nterms = make([]*RegisterAccess[F, Computation[F]], len(t.Vars))
		//
		for i, v := range t.Vars {
			nterms[i] = RawRegisterAccess[F, Computation[F]](v.Register, v.Shift)
		}
		//
		return NewVectorAccess(nterms)
	default:
		panic(fmt.Sprintf("unknown computation encountered: %s", term.Lisp(false, nil).String(false)))
	}
}

// NewComputations constructs an array of zero or more computations.
func NewComputations[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](terms []T) []Computation[F] {
	var computations = make([]Computation[F], len(terms))
	//
	for i, t := range terms {
		computations[i] = NewComputation[F, S](t)
	}
	//
	return computations
}

// ============================================================================
// Logical Comptuation
// ============================================================================

// NewLogicalComputation takes an arbitrary logical term and converts in into an
// instance of computation by wrapping it.
func NewLogicalComputation[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](term LogicalTerm[F, S],
) LogicalComputation[F] {
	switch t := term.(type) {
	case *Conjunct[F, S]:
		args := NewLogicalComputations[F, S, T](t.Args)
		return Conjunction(args...)
	case *Disjunct[F, S]:
		args := NewLogicalComputations[F, S, T](t.Args)
		return Disjunction(args...)
	case *Equal[F, S, T]:
		lhs := NewComputation[F, S](t.Lhs)
		rhs := NewComputation[F, S](t.Rhs)

		return Equals[F, LogicalComputation[F]](lhs, rhs)
	case *Ite[F, S]:
		var trueBranch, falseBranch LogicalComputation[F]

		condition := NewLogicalComputation[F, S, T](t.Condition)
		//
		if t.TrueBranch != nil {
			trueBranch = NewLogicalComputation[F, S, T](t.TrueBranch)
		}
		//
		if t.FalseBranch != nil {
			falseBranch = NewLogicalComputation[F, S, T](t.FalseBranch)
		}
		//
		return IfThenElse(condition, trueBranch, falseBranch)
	case *Inequality[F, S, T]:
		lhs := NewComputation[F, S](t.Lhs)
		rhs := NewComputation[F, S](t.Rhs)
		//
		if t.Strict {
			return LessThan[F, LogicalComputation[F]](lhs, rhs)
		}
		//
		return LessThanOrEquals[F, LogicalComputation[F]](lhs, rhs)
	case *Negate[F, S]:
		arg := NewLogicalComputation[F, S, T](t.Arg)
		return Negation(arg)
	case *NotEqual[F, S, T]:
		lhs := NewComputation[F, S](t.Lhs)
		rhs := NewComputation[F, S](t.Rhs)

		return NotEquals[F, LogicalComputation[F]](lhs, rhs)
	default:
		panic(fmt.Sprintf("unknown computation encountered: %s", term.Lisp(false, nil).String(false)))
	}
}

// NewLogicalComputations constructs an array of zero or more computations.
func NewLogicalComputations[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](terms []S) []LogicalComputation[F] {
	var computations = make([]LogicalComputation[F], len(terms))
	//
	for i, t := range terms {
		computations[i] = NewLogicalComputation[F, S, T](t)
	}
	//
	return computations
}

// ============================================================================
// Concretization
// ============================================================================

// ConcretizeComputation concretizes a computation by converting it from
// operating in one field to operating in another.  This requires rebuilding the
// tree for the new field.
func ConcretizeComputation[F1 field.Element[F1], F2 field.Element[F2]](c Computation[F1]) Computation[F2] {
	switch t := c.(type) {
	case *Add[F1, Computation[F1]]:
		args := ConcretizeComputations[F1, F2](t.Args)
		return Sum(args...)
	case *Cast[F1, Computation[F1]]:
		arg := ConcretizeComputation[F1, F2](t.Arg)
		return CastOf(arg, t.BitWidth)
	case *Constant[F1, Computation[F1]]:
		var val F2
		return Const[F2, Computation[F2]](val.SetBytes(t.Value.Bytes()))
	case *Exp[F1, Computation[F1]]:
		arg := ConcretizeComputation[F1, F2](t.Arg)
		return Exponent(arg, t.Pow)
	case *IfZero[F1, LogicalComputation[F1], Computation[F1]]:
		condition := ConcretizeLogicalComputation[F1, F2](t.Condition)
		trueBranch := ConcretizeComputation[F1, F2](t.TrueBranch)
		falseBranch := ConcretizeComputation[F1, F2](t.FalseBranch)
		// Done
		return IfElse(condition, trueBranch, falseBranch)
	case *LabelledConst[F1, Computation[F1]]:
		var val F2
		return LabelledConstant[F2, Computation[F2]](t.Label, val.SetBytes(t.Value.Bytes()))
	case *Mul[F1, Computation[F1]]:
		args := ConcretizeComputations[F1, F2](t.Args)
		return Product(args...)
	case *Norm[F1, Computation[F1]]:
		arg := ConcretizeComputation[F1, F2](t.Arg)
		return Normalise(arg)
	case *RegisterAccess[F1, Computation[F1]]:
		return NewRegisterAccess[F2, Computation[F2]](t.Register, t.Shift)
	case *Sub[F1, Computation[F1]]:
		args := ConcretizeComputations[F1, F2](t.Args)
		return Subtract(args...)
	case *VectorAccess[F1, Computation[F1]]:
		var nterms = make([]*RegisterAccess[F2, Computation[F2]], len(t.Vars))
		//
		for i, v := range t.Vars {
			nterms[i] = RawRegisterAccess[F2, Computation[F2]](v.Register, v.Shift)
		}
		//
		return NewVectorAccess(nterms)
	default:
		panic(fmt.Sprintf("unknown computation encountered: %s", c.Lisp(false, nil).String(false)))
	}
}

// ConcretizeComputations concretizes an array of zero or more logical computations.
func ConcretizeComputations[F1 field.Element[F1], F2 field.Element[F2]](cs []Computation[F1]) []Computation[F2] {
	var computations = make([]Computation[F2], len(cs))
	//
	for i, t := range cs {
		computations[i] = ConcretizeComputation[F1, F2](t)
	}
	//
	return computations
}

// ConcretizeLogicalComputation concretizes a logical computation by converting
// it from operating in one field to operating in another.  This requires
// rebuilding the tree for the new field.
func ConcretizeLogicalComputation[F1 field.Element[F1], F2 field.Element[F2]](c LogicalComputation[F1],
) LogicalComputation[F2] {
	//
	switch t := c.(type) {
	case *Conjunct[F1, LogicalComputation[F1]]:
		args := ConcretizeLogicalComputations[F1, F2](t.Args)
		return Conjunction(args...)
	case *Disjunct[F1, LogicalComputation[F1]]:
		args := ConcretizeLogicalComputations[F1, F2](t.Args)
		return Disjunction(args...)
	case *Equal[F1, LogicalComputation[F1], Computation[F1]]:
		lhs := ConcretizeComputation[F1, F2](t.Lhs)
		rhs := ConcretizeComputation[F1, F2](t.Rhs)

		return Equals[F2, LogicalComputation[F2]](lhs, rhs)
	case *Ite[F1, LogicalComputation[F1]]:
		var trueBranch, falseBranch LogicalComputation[F2]

		condition := ConcretizeLogicalComputation[F1, F2](t.Condition)
		//
		if t.TrueBranch != nil {
			trueBranch = ConcretizeLogicalComputation[F1, F2](t.TrueBranch)
		}
		//
		if t.FalseBranch != nil {
			falseBranch = ConcretizeLogicalComputation[F1, F2](t.FalseBranch)
		}
		//
		return IfThenElse(condition, trueBranch, falseBranch)
	case *Inequality[F1, LogicalComputation[F1], Computation[F1]]:
		lhs := ConcretizeComputation[F1, F2](t.Lhs)
		rhs := ConcretizeComputation[F1, F2](t.Rhs)
		//
		if t.Strict {
			return LessThan[F2, LogicalComputation[F2]](lhs, rhs)
		}
		//
		return LessThanOrEquals[F2, LogicalComputation[F2]](lhs, rhs)
	case *Negate[F1, LogicalComputation[F1]]:
		arg := ConcretizeLogicalComputation[F1, F2](t.Arg)
		return Negation(arg)
	case *NotEqual[F1, LogicalComputation[F1], Computation[F1]]:
		lhs := ConcretizeComputation[F1, F2](t.Lhs)
		rhs := ConcretizeComputation[F1, F2](t.Rhs)

		return NotEquals[F2, LogicalComputation[F2]](lhs, rhs)
	default:
		panic(fmt.Sprintf("unknown computation encountered: %s", c.Lisp(false, nil).String(false)))
	}
}

// ConcretizeLogicalComputations concretizes an array of zero or more logical computations.
func ConcretizeLogicalComputations[F1 field.Element[F1], F2 field.Element[F2]](cs []LogicalComputation[F1],
) []LogicalComputation[F2] {
	//
	var computations = make([]LogicalComputation[F2], len(cs))
	//
	for i, t := range cs {
		computations[i] = ConcretizeLogicalComputation[F1, F2](t)
	}
	//
	return computations
}
