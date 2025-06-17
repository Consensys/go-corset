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

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// Module provides an abstraction for modules in the underlying constraint
// system.
type Module[T any, E Expr[T, E], M any] interface {
	// SetName sets the name of this module.
	Initialise(fn MicroFunction, mid uint) M

	// NewColumn constructs a new column of the given name and bitwidth within
	// this module.
	NewColumn(kind schema.RegisterType, name string, bitwidth uint) T

	// NewConstraint constructs a new vanishing constraint with the given name
	// within this module.  An optional "domain" can be given which determines
	// whether or not this is a "local" or "global" constraint.  Specifically, a
	// local constraint applies only on one row whereas a global constraints
	// applies on all rows.  The domain (if supplied) determines the row where a
	// local constraint applies, with negative values being offset from the last
	// row.  Thus, a domain value of 0 (reps -1) represents the first (resp.
	// last) row of the module.
	NewConstraint(name string, domain util.Option[int], expr E)

	// NewLookup constructs a new lookup constraint
	NewLookup(name string, from []E, target uint, to []E)

	// String returns an appropriately formatted representation of the module.
	String() string
}

// Expr provides an abstraction over expressions in the constraint language.
// Using an abstraction, rather than concrete constraint expressions directly,
// makes it relatively easier to support multiple target languages.
type Expr[T, E any] interface {
	// Add constructs a sum between this expression and zero or more
	Add(exprs ...E) E

	// And constructs a conjunction between this expression and zero or more
	// expressions.
	And(...E) E

	// Equals constructs an equality between two expressions.
	Equals(rhs E) E

	// Then constructs an implication between two expressions.
	Then(trueBranch E) E

	// ThenElse constructs an if-then-else expression with this expression
	// acting as the condition.
	ThenElse(trueBranch E, falseBranch E) E

	// Multiply constructs a product between this expression and zero or more
	// expressions.
	Multiply(...E) E

	// NotEquals constructs a non-equality between two expressions.
	NotEquals(rhs E) E

	// Number constructs a constant expression.
	BigInt(number big.Int) E

	// Or constructs a disjunction between this expression and zero or more
	// expressions.
	Or(...E) E

	// Variable constructs a variable with a given shift.
	Variable(name T, shift int) E
}

// BigNumber constructs a constant expression from a big integer.
func BigNumber[T any, E Expr[T, E]](c *big.Int) E {
	var (
		empty E
		val   big.Int
	)
	// Clone big integer
	val.Set(c)
	//
	return empty.BigInt(val)
}

// If constructs an if-then expression.
func If[T any, E Expr[T, E]](condition E, trueBranch E) E {
	return condition.Then(trueBranch)
}

// IfElse constructs an if-then-else expression.
func IfElse[T any, E Expr[T, E]](condition E, trueBranch E, falseBranch E) E {
	return condition.ThenElse(trueBranch, falseBranch)
}

// Number constructs a constant expression from an unsigned integer.
func Number[T any, E Expr[T, E]](c uint) E {
	return BigNumber[T, E](big.NewInt(int64(c)))
}

// Sum constructs a sum over one or more expressions.
func Sum[T any, E Expr[T, E]](exprs []E) E {
	if len(exprs) == 0 {
		return Number[T, E](0)
	}
	//
	return exprs[0].Add(exprs[1:]...)
}

// Product constructs a product over one or more expressions.
func Product[T any, E Expr[T, E]](exprs []E) E {
	if len(exprs) == 0 {
		return Number[T, E](0)
	}
	//
	return exprs[0].Multiply(exprs[1:]...)
}

// Variable is just a convenient wrapper for creating abstract expressions
// representing variable accesses.
func Variable[T any, E Expr[T, E]](id T, shift int) E {
	var empty E
	//
	return empty.Variable(id, shift)
}
