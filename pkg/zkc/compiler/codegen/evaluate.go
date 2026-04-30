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
package codegen

import (
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

func evalConstants(
	es []Expr, definition bool, declarations []Declaration, env data.ResolvedEnvironment,
) ([]word.Uint, string) {
	words := make([]word.Uint, len(es))

	var errorMessage string

	for i, e := range es {
		var errorMsg string

		words[i], errorMsg = EvalConstant(e, definition, declarations, env)

		if errorMsg != "" {
			errorMessage = errorMsg
		}
	}
	//
	return words, errorMessage
}

// EvalConstant evaluates a compile-time constant expression using the
// provided declaration list and type environment.  It is used both during
// function code generation and when initialising static memory contents.
func EvalConstant(
	e Expr, definition bool, declarations []Declaration, env data.ResolvedEnvironment,
) (res word.Uint, errorMessage string) {
	var overflow bool

	bitwidth := data.BitWidthOf(e.Type(), env)
	//
	switch e := e.(type) {
	case *expr.Add[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res, overflow = word.Sum(bitwidth, args...)

		if overflow && definition {
			errorMessage = "arithmetic overflow"
		}

		return
	case *expr.Sub[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res, overflow = word.Subtract(bitwidth, args...)

		if overflow && definition {
			errorMessage = "arithmetic underflow"
		}

		return

	case *expr.BitwiseAnd[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return word.BitwiseAnd(bitwidth, args...), ""
	case *expr.Const[symbol.Resolved]:
		var c word.Uint
		//
		return c.SetBigInt(&e.Constant), ""
	case *expr.Mul[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res, overflow = word.Product(bitwidth, args...)

		if overflow && definition {
			errorMessage = "arithmetic overflow"
		}

		return
	case *expr.Div[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res = word.Quotient(bitwidth, args...)

		return
	case *expr.Rem[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res = word.Remainder(bitwidth, args...)

		return
	case *expr.BitwiseNot[symbol.Resolved]:
		arg, _ := EvalConstant(e.Expr, definition, declarations, env)
		return arg.Not(bitwidth), ""
	case *expr.BitwiseOr[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return word.BitwiseOr(bitwidth, args...), ""
	case *expr.Shl[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return word.BitwiseShl(bitwidth, args...), ""
	case *expr.Shr[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return word.BitwiseShr(bitwidth, args...), ""
	case *expr.Xor[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return word.BitwiseXor(bitwidth, args...), ""
	case *expr.Cast[symbol.Resolved]:
		inner, _ := EvalConstant(e.Expr, definition, declarations, env)
		width := e.CastType.AsUint(env).BitWidth()
		sliced := inner.Slice(width)

		if inner.Cmp(sliced) != 0 && definition {
			errorMessage = "cast overflow"
		}

		return sliced, errorMessage
	case *expr.ExternAccess[symbol.Resolved]:
		c, ok := declarations[e.Name.Index].(*decl.ResolvedConstant)
		if !ok {
			return res, "not a constant expression"
		}

		res, _ = EvalConstant(c.ConstExpr, false, declarations, env)

		return res, ""
	default:
		return res, "not a constant expression"
	}
}
