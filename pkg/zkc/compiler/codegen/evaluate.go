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
	"github.com/consensys/go-corset/pkg/zkc/vm"
)

func evalConstants(
	es []Expr, definition bool, declarations []Declaration, env data.ResolvedEnvironment,
) ([]vm.Uint, string) {
	words := make([]vm.Uint, len(es))

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

// EvalConstant evaluates a compile-time constant expression using the provided
// declaration list and type environment.  It is used during function code
// generation and when initialising static memory contents, and also during
// typing (for array type size expressions).  As a result of the latter, it must
// be robust against error.  That is, it may be called on a malformed expression
// and, hence, it must handle this gracefully.
func EvalConstant(
	e Expr, definition bool, declarations []Declaration, env data.ResolvedEnvironment,
) (res vm.Uint, errorMessage string) {
	var (
		overflow, ok bool
		bitwidth     uint
	)
	// NOTE: we must sanity check the bitwidth identified is valid in order to
	// ensure this function is robust against errors.  This is necessary because
	// it is used during typing and, thus, could be called on a malformed
	// expression as a result.
	if bitwidth, ok = data.BitWidthOf(e.Type(), env); !ok {
		return res, "invalid constant"
	}
	//
	switch e := e.(type) {
	case *expr.Add[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res, overflow = Sum(bitwidth, args...)

		if overflow && definition {
			errorMessage = "arithmetic overflow"
		}

		return
	case *expr.Sub[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res, overflow = Subtract(bitwidth, args...)

		if overflow && definition {
			errorMessage = "arithmetic underflow"
		}

		return

	case *expr.BitwiseAnd[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return BitwiseAnd(bitwidth, args...), ""
	case *expr.Const[symbol.Resolved]:
		var c vm.Uint
		//
		return c.SetBigInt(&e.Constant), ""
	case *expr.Mul[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res, overflow = Product(bitwidth, args...)

		if overflow && definition {
			errorMessage = "arithmetic overflow"
		}

		return
	case *expr.Div[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res = Quotient(bitwidth, args...)

		return
	case *expr.Rem[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		res = Remainder(bitwidth, args...)

		return
	case *expr.BitwiseNot[symbol.Resolved]:
		arg, _ := EvalConstant(e.Expr, definition, declarations, env)
		return arg.Not(bitwidth), ""
	case *expr.BitwiseOr[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return BitwiseOr(bitwidth, args...), ""
	case *expr.Shl[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return BitwiseShl(bitwidth, args...), ""
	case *expr.Shr[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return BitwiseShr(bitwidth, args...), ""
	case *expr.Xor[symbol.Resolved]:
		args, _ := evalConstants(e.Exprs, definition, declarations, env)
		return BitwiseXor(bitwidth, args...), ""
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

// Sum a given set of words together.
func Sum[W vm.Word[W]](bitwidth uint, values ...W) (W, bool) {
	var (
		res      W
		overflow bool
	)
	//
	for i, v := range values {
		var carry bool
		//
		if i == 0 {
			res = v
		} else {
			res, carry = res.Add(bitwidth, v)
			//
			overflow = overflow || carry
		}
	}
	//
	return res, overflow
}

// Subtract a given set of words together, producing the difference and an
// underflow indicator.
func Subtract[W vm.Word[W]](bitwidth uint, values ...W) (W, bool) {
	var (
		res       W
		underflow bool
	)
	//
	for i, v := range values {
		var borrow bool
		//
		if i == 0 {
			res = v
		} else {
			res, borrow = res.Sub(bitwidth, v)
			//
			underflow = underflow || borrow
		}
	}
	//
	return res, underflow
}

// BitwiseAnd computes the bitwise AND of a set of words.
func BitwiseAnd[W vm.Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.And(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseOr computes the bitwise OR of a set of words.
func BitwiseOr[W vm.Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Or(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseXor computes the bitwise XOR of a set of words.
func BitwiseXor[W vm.Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Xor(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseShl computes a left-shift chain over a set of words.
func BitwiseShl[W vm.Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Shl(bitwidth, v)
		}
	}
	//
	return res
}

// BitwiseShr computes a right-shift chain over a set of words.
func BitwiseShr[W vm.Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Shr(bitwidth, v)
		}
	}
	//
	return res
}

// Quotient divides a sequence of words left-to-right.
func Quotient[W vm.Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Div(bitwidth, v)
		}
	}
	//
	return res
}

// Remainder computes the remainder of dividing a sequence of words left-to-right.
func Remainder[W vm.Word[W]](bitwidth uint, values ...W) W {
	var res W
	//
	for i, v := range values {
		if i == 0 {
			res = v
		} else {
			res = res.Rem(bitwidth, v)
		}
	}
	//
	return res
}

// Product mulitplies a given set of words together.
func Product[W vm.Word[W]](bitwidth uint, values ...W) (W, bool) {
	var (
		res      W
		overflow bool
	)
	//
	for i, v := range values {
		var carry bool

		if i == 0 {
			res = v
		} else {
			res, carry = res.Mul(bitwidth, v)
			//
			overflow = overflow || carry
		}
	}
	//
	return res, overflow
}
