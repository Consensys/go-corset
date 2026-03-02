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
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

var (
	biZERO big.Int = *big.NewInt(0)
	biONE  big.Int = *big.NewInt(1)
)

// Expr represents an arbitrary expression used within an instruction.
type Expr[I symbol.Symbol[I]] interface {
	// NonLocalUses returns the set of non-local declarations accessed by this
	// expression.  For example, external constants or memories used within.
	NonLocalUses() set.AnySortedSet[I]
	// RegistersRead returns the set of variables used (i.e. read) by this expression
	LocalUses() bit.Set
	// String returns a string representation of this expression.
	String(mapping variable.Map) string
	// ValueRange returns the interval of values that this term can evaluate to.
	// For terms accessing registers, this is determined by the declared width of
	// the register.
	ValueRange(env variable.Map) math.Interval
}

// BitWidth returns the minimum number of bits required to store any evaluation
// of this expression.  In addition, it provides an indicator as to whether or
// not any evaluation could result in a negative value.
func BitWidth[I symbol.Symbol[I]](e Expr[I], env variable.Map) (uint, bool) {
	var (
		// Determine set of all values that right-hand side can evaluate to
		values = e.ValueRange(env)
		// Determine bitwidth required to contain all values
		bitwidth, signed = values.BitWidth()
	)
	// For signed arithmetic, we need a specific sign bit.
	if signed {
		bitwidth++
	}
	//
	return bitwidth, signed
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
	case *Mul[I]:
		exprs = e.Exprs
		operator = "*"
	case *LocalAccess[I]:
		return mapping.Variable(e.Variable).Name
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
	default:
		return true
	}
}
