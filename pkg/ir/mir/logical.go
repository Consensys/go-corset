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
package mir

import (
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/ir/schema/expr"
	"github.com/consensys/go-corset/pkg/ir/schema/term"
)

type LogicalTerm interface {
	schema.LogicalTerm[LogicalTerm]
}

type Logical = expr.Logical[LogicalTerm]

func Conjunction(terms ...Logical) Logical {
	panic("todo")
}

func Disjunction(terms ...Logical) Logical {
	panic("todo")
}

// Equals constructs an equation representing the equality of two expressions.
func Equals(lhs Expr, rhs Expr) Logical {
	term := &term.Equation[Term]{
		Kind: term.EQUALS,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical{Term: term}
}

// Negate constructs the logical negation of the given Logical[T].
func Negate(expr Logical) Logical {
	panic("todo")
}

// NotEquals constructs an equation representing the non-equality of two
// expressions.
func NotEquals(lhs Expr, rhs Expr) Logical {
	term := &term.Equation[Term]{
		Kind: term.EQUALS,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical{Term: term}
}
