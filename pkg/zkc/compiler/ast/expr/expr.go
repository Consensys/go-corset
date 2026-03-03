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
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Expr represents an arbitrary expression used within an instruction.
type Expr[I symbol.Symbol[I]] interface {
	// BitWidth returns the minimum number of bits required to hold any
	// evaluation of this expression.
	BitWidth() uint
	// NonLocalUses returns the set of non-local declarations accessed by this
	// expression.  For example, external constants or memories used within.
	NonLocalUses() set.AnySortedSet[I]
	// RegistersRead returns the set of variables used (i.e. read) by this expression
	LocalUses() bit.Set
	// String returns a string representation of this expression.
	String(mapping variable.Map) string
}

// Uses determines the (unique) set of registers read by any expression
// in the given set of expressions.
func Uses[I symbol.Symbol[I]](exprs ...Expr[I]) []variable.Id {
	var (
		reads []variable.Id
		bits  bit.Set
	)
	// extract all usages
	for _, e := range exprs {
		bits.Union(e.LocalUses())
	}
	// Collect them all up
	for iter := bits.Iter(); iter.HasNext(); {
		next := iter.Next()
		//
		reads = append(reads, next)
	}
	//
	return reads
}

// String provides a generic facility for converting an expression into a
// suitable string.
func String[I symbol.Symbol[I]](e Expr[I], mapping variable.Map) string {
	var (
		exprs    []Expr[I]
		operator string
		builder  strings.Builder
	)
	//
	switch e := e.(type) {
	case *Add[I]:
		operator = "+"
		exprs = e.Exprs
	case *Const[I]:
		return stringOfConstant(e.Constant, e.Base)
	case *LocalAccess[I]:
		return mapping.Variable(e.Variable).Name
	case *Mul[I]:
		exprs = e.Exprs
		operator = "*"
	case *NonLocalAccess[I]:
		return e.Name.String()
	case *Sub[I]:
		exprs = e.Exprs
		operator = "-"
	default:
		panic("unreachable")
	}
	//
	for i, e := range exprs {
		if i != 0 {
			builder.WriteString(" ")
			builder.WriteString(operator)
			builder.WriteString(" ")
		}
		//
		if needsBraces[I](e) {
			builder.WriteString("(")
			builder.WriteString(String[I](e, mapping))
			builder.WriteString(")")
		} else {
			builder.WriteString(String[I](e, mapping))
		}
	}
	//
	return builder.String()
}

func stringOfConstant(val big.Int, base uint) string {
	switch base {
	case 2:
		return fmt.Sprintf("0b%s", val.Text(2))
	case 16:
		return fmt.Sprintf("0x%s", val.Text(16))
	default:
		return val.String()
	}
}

func needsBraces[I symbol.Symbol[I]](e Expr[I]) bool {
	switch e.(type) {
	case *Const[I]:
		return false
	case *LocalAccess[I]:
		return false
	case *NonLocalAccess[I]:
		return false
	default:
		return true
	}
}
