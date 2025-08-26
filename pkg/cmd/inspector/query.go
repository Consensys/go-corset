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
package inspector

import (
	"fmt"
	"math/big"
	"slices"
	"strings"

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
)

const qROW = 0
const qVAR = 1
const qNUM = 2
const qOR = 3
const qAND = 4
const qEQ = 5
const qNEQ = 6
const qLT = 7
const qLTEQ = 8
const qADD = 9
const qMUL = 10
const qSUB = 11

// Query represents a boolean expression which can be evaluated over a
// given set of columns.
type Query[F field.Element[F]] struct {
	// operation
	op int
	// arguments (if applicable)
	args []Query[F]
	// constant value (if applicable)
	number F
	// variable name (if applicable)
	name string
}

// Variable constructs a variable of the given name.
func (p *Query[F]) Variable(name string) *Query[F] {
	var query Query[F]
	//
	if name == "$" {
		query.op = qROW
		query.name = name
	} else {
		query.op = qVAR
		query.name = name
	}
	//
	return &query
}

// Number constructs a number with the given value
func (p *Query[F]) Number(number big.Int) *Query[F] {
	var (
		val   F
		query Query[F]
	)
	//
	val = val.SetBytes(number.Bytes())
	//
	query.op = qNUM
	query.number = val
	//
	return &query
}

// Or constructs a disjunction of queries.
func (p *Query[F]) Or(queries ...*Query[F]) *Query[F] {
	args := build_args(p, queries)
	//
	return &Query[F]{qOR, args, field.Zero[F](), ""}
}

// And constructs a conjunction of queries.
func (p *Query[F]) And(queries ...*Query[F]) *Query[F] {
	args := build_args(p, queries)
	//
	return &Query[F]{qAND, args, field.Zero[F](), ""}
}

// Equals constructs an equality between two queries.
func (p *Query[F]) Equals(rhs *Query[F]) *Query[F] {
	return &Query[F]{qEQ, []Query[F]{*p, *rhs}, field.Zero[F](), ""}
}

// NotEquals constructs a non-equality between two queries.
func (p *Query[F]) NotEquals(rhs *Query[F]) *Query[F] {
	return &Query[F]{qNEQ, []Query[F]{*p, *rhs}, field.Zero[F](), ""}
}

// LessThan constructs a (strict) inequality between two queries.
func (p *Query[F]) LessThan(rhs *Query[F]) *Query[F] {
	return &Query[F]{qLT, []Query[F]{*p, *rhs}, field.Zero[F](), ""}
}

// LessThanEquals constructs a (non-strict) inequality between two queries.
func (p *Query[F]) LessThanEquals(rhs *Query[F]) *Query[F] {
	return &Query[F]{qLTEQ, []Query[F]{*p, *rhs}, field.Zero[F](), ""}
}

// Add constructs the sum of one or more queries
func (p *Query[F]) Add(queries ...*Query[F]) *Query[F] {
	args := build_args(p, queries)
	//
	return &Query[F]{qADD, args, field.Zero[F](), ""}
}

// Mul constructs the product of one or more queries
func (p *Query[F]) Mul(queries ...*Query[F]) *Query[F] {
	args := build_args(p, queries)
	//
	return &Query[F]{qMUL, args, field.Zero[F](), ""}
}

// Sub constructs the subtraction of one or more queries
func (p *Query[F]) Sub(queries ...*Query[F]) *Query[F] {
	args := build_args(p, queries)
	//
	return &Query[F]{qSUB, args, field.Zero[F](), ""}
}

func build_args[F field.Element[F]](q *Query[F], queries []*Query[F]) []Query[F] {
	args := make([]Query[F], 1+len(queries))
	args[0] = *q
	//
	for i := range queries {
		args[i+1] = *queries[i]
	}
	//
	return args
}

// String produces a parseable string from this query.
func (p *Query[F]) String() string {
	return query_string(*p)
}

// Eval evaluates the given query in the given environment.
func (p *Query[F]) Eval(row uint, env map[string]tr.Column[F]) F {
	switch p.op {
	case qROW:
		var val F
		return val.SetUint64(uint64(row))
	case qVAR:
		if col, ok := env[p.name]; ok {
			return col.Get(int(row))
		}
		// error
		panic("unknown column \"%s\"")
	case qNUM:
		return p.number
	case qEQ:
		return eval_binary(row, env, p.args[0], p.args[1], eval_eq)
	case qNEQ:
		return eval_binary(row, env, p.args[0], p.args[1], eval_neq)
	case qLT:
		return eval_binary(row, env, p.args[0], p.args[1], eval_lt)
	case qLTEQ:
		return eval_binary(row, env, p.args[0], p.args[1], eval_lteq)
	case qOR:
		return eval_nary(row, env, p.args, eval_or)
	case qAND:
		return eval_nary(row, env, p.args, eval_and)
	case qADD:
		return eval_nary(row, env, p.args, eval_add)
	case qMUL:
		return eval_nary(row, env, p.args, eval_mul)
	case qSUB:
		return eval_nary(row, env, p.args, eval_sub)
	default:
		panic(fmt.Sprintf("unknown operator (%d)", p.op))
	}
}

type binary_op[F field.Element[F]] func(F, F) F
type nary_op[F field.Element[F]] func([]F) F

func eval_binary[F field.Element[F]](row uint, env map[string]tr.Column[F], lhs Query[F], rhs Query[F],
	fn binary_op[F]) F {
	// Evaluate left-hand side
	lv := lhs.Eval(row, env)
	// Evaluate right-hand side
	rv := rhs.Eval(row, env)
	// Performan binary operation
	return fn(lv, rv)
}

func eval_nary[F field.Element[F]](row uint, env map[string]tr.Column[F], args []Query[F], fn nary_op[F]) F {
	vals := make([]F, len(args))
	// Evaluate arguments
	for i, arg := range args {
		vals[i] = arg.Eval(row, env)
	}
	//
	return fn(vals)
}

func eval_eq[F field.Element[F]](lhs F, rhs F) F {
	// Perform comparison
	if lhs.Cmp(rhs) == 0 {
		return field.Zero[F]()
	}
	//
	return field.One[F]()
}

func eval_neq[F field.Element[F]](lhs F, rhs F) F {
	// Perform comparison
	if lhs.Cmp(rhs) != 0 {
		return field.Zero[F]()
	}
	//
	return field.One[F]()
}

func eval_lt[F field.Element[F]](lhs F, rhs F) F {
	// Perform comparison
	if lhs.Cmp(rhs) < 0 {
		return field.Zero[F]()
	}
	//
	return field.One[F]()
}

func eval_lteq[F field.Element[F]](lhs F, rhs F) F {
	// Perform comparison
	if lhs.Cmp(rhs) <= 0 {
		return field.Zero[F]()
	}
	//
	return field.One[F]()
}

func eval_or[F field.Element[F]](vals []F) F {
	for _, v := range vals {
		if v.IsZero() {
			// Success
			return v
		}
	}
	// Fail
	return field.One[F]()
}

func eval_and[F field.Element[F]](vals []F) F {
	//
	for _, v := range vals {
		if !v.IsZero() {
			// Fail
			return v
		}
	}
	// Success
	return field.Zero[F]()
}

func eval_add[F field.Element[F]](vals []F) F {
	val := field.Zero[F]()
	//
	for _, v := range vals {
		val = val.Add(v)
	}
	//
	return val
}

func eval_mul[F field.Element[F]](vals []F) F {
	val := field.One[F]()
	//
	for _, v := range vals {
		val = val.Mul(v)
	}
	//
	return val
}

func eval_sub[F field.Element[F]](vals []F) F {
	val := vals[0]
	//
	for _, v := range vals[1:] {
		val = val.Sub(v)
	}
	//
	return val
}

func query_string[F field.Element[F]](p Query[F], braces ...int) string {
	var str string
	//
	switch p.op {
	case qVAR, qROW:
		return p.name
	case qNUM:
		return p.number.String()
	case qEQ:
		str = query_strings("==", false, p.args)
	case qNEQ:
		str = query_strings("!=", false, p.args)
	case qLT:
		str = query_strings("<", false, p.args)
	case qLTEQ:
		str = query_strings("<=", false, p.args)
	case qOR:
		str = query_strings("∨", true, p.args, qOR, qAND)
	case qAND:
		str = query_strings("∧", true, p.args, qOR, qAND)
	case qADD:
		str = query_strings("+", false, p.args, qADD, qMUL, qSUB)
	case qMUL:
		str = query_strings("*", false, p.args, qADD, qMUL, qSUB)
	case qSUB:
		str = query_strings("-", false, p.args, qADD, qMUL, qSUB)
	default:
		panic(fmt.Sprintf("unknown operator (%d)", p.op))
	}
	// Check whether braces required
	if slices.Contains(braces, p.op) {
		return fmt.Sprintf("(%s)", str)
	}
	// nope
	return str
}

func query_strings[F field.Element[F]](op string, spacing bool, queries []Query[F], braces ...int) string {
	var builder strings.Builder
	//
	for i, q := range queries {
		if i != 0 && spacing {
			builder.WriteString(" ")
			builder.WriteString(op)
			builder.WriteString(" ")
		} else if i != 0 {
			builder.WriteString(op)
		}

		builder.WriteString(query_string(q, braces...))
	}
	//
	return builder.String()
}
