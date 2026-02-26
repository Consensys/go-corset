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
package expr

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

const (
	// EQ indicates an equality condition
	EQ CmpOp = 0
	// NEQ indicates a non-equality condition
	NEQ CmpOp = 1
	// LT indicates a less-than condition
	LT CmpOp = 2
	// GT indicates a greater-than condition
	GT CmpOp = 3
	// LTEQ indicates a less-than-or-equals condition
	LTEQ CmpOp = 4
	// GTEQ indicates a greater-than-or-equals condition
	GTEQ CmpOp = 5
)

// CmpOp represents the set of possible operators for a comparison condition.
type CmpOp uint8

// Cmp represents a comparison, such as "==", ">=", etc.
type Cmp struct {
	// Operator indicates the condition
	Operator CmpOp
	// Left-hand side
	Left Expr
	// Right-hand side
	Right Expr
}

// Negate implementation for Condition interface.
func (p *Cmp) Negate() Condition {
	var op CmpOp
	//
	switch p.Operator {
	case EQ:
		op = NEQ
	case NEQ:
		op = EQ
	case LT:
		op = GTEQ
	case LTEQ:
		op = GT
	case GT:
		op = LTEQ
	case GTEQ:
		op = LT
	default:
		panic("unreachable")
	}
	//
	return &Cmp{op, p.Left, p.Right}
}

// Uses implementation for the Expr interface.
func (p *Cmp) Uses() bit.Set {
	var reads bit.Set
	//
	reads.Union(p.Left.Uses())
	reads.Union(p.Right.Uses())
	//
	return reads
}

func (p *Cmp) String(env variable.Map) string {
	var (
		l  = p.Left.String(env)
		r  = p.Right.String(env)
		op string
	)
	//
	switch p.Operator {
	case EQ:
		op = "=="
	case NEQ:
		op = "!="
	case LT:
		op = "<"
	case LTEQ:
		op = "<="
	case GT:
		op = ">"
	case GTEQ:
		op = ">="
	default:
		panic("unreachable")
	}
	//
	return fmt.Sprintf("%s%s%s", l, op, r)
}
