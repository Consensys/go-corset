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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
)

const qVAR = 0
const qNUM = 1
const qOR = 2
const qAND = 3
const qEQ = 4
const qNEQ = 5
const qLT = 6
const qLTEQ = 7
const qADD = 8
const qMUL = 9
const qSUB = 10

// Query represents a boolean expression which can be evaluated over a
// given set of columns.
type Query struct {
	// operation
	op int
	// arguments (if applicable)
	args []Query
	// constant value (if applicable)
	number fr.Element
	// variable name (if applicable)
	name string
}

// Variable constructs a variable of the given name.
func (p *Query) Variable(name string) *Query {
	var query Query
	query.op = qVAR
	query.name = name
	//
	return &query
}

// Number constructs a number with the given value
func (p *Query) Number(number big.Int) *Query {
	var (
		val   fr.Element
		query Query
	)
	//
	val.SetBigInt(&number)
	//
	query.op = qNUM
	query.number = val
	//
	return &query
}

// Or constructs a disjunction of queries.
func (p *Query) Or(queries ...*Query) *Query {
	args := build_args(p, queries)
	//
	return &Query{qOR, args, fr.Element{}, ""}
}

// And constructs a conjunction of queries.
func (p *Query) And(queries ...*Query) *Query {
	args := build_args(p, queries)
	//
	return &Query{qAND, args, fr.Element{}, ""}
}

// Equals constructs an equality between two queries.
func (p *Query) Equals(rhs *Query) *Query {
	return &Query{qEQ, []Query{*p, *rhs}, fr.Element{}, ""}
}

// NotEquals constructs a non-equality between two queries.
func (p *Query) NotEquals(rhs *Query) *Query {
	return &Query{qNEQ, []Query{*p, *rhs}, fr.Element{}, ""}
}

// LessThan constructs a (strict) inequality between two queries.
func (p *Query) LessThan(rhs *Query) *Query {
	return &Query{qLT, []Query{*p, *rhs}, fr.Element{}, ""}
}

// LessThanEquals constructs a (non-strict) inequality between two queries.
func (p *Query) LessThanEquals(rhs *Query) *Query {
	return &Query{qLTEQ, []Query{*p, *rhs}, fr.Element{}, ""}
}

// Add constructs the sum of one or more queries
func (p *Query) Add(queries ...*Query) *Query {
	args := build_args(p, queries)
	//
	return &Query{qADD, args, fr.Element{}, ""}
}

// Mul constructs the product of one or more queries
func (p *Query) Mul(queries ...*Query) *Query {
	args := build_args(p, queries)
	//
	return &Query{qMUL, args, fr.Element{}, ""}
}

// Sub constructs the subtraction of one or more queries
func (p *Query) Sub(queries ...*Query) *Query {
	args := build_args(p, queries)
	//
	return &Query{qSUB, args, fr.Element{}, ""}
}

func build_args(q *Query, queries []*Query) []Query {
	args := make([]Query, 1+len(queries))
	args[0] = *q
	//
	for i := range queries {
		args[i+1] = *queries[i]
	}
	//
	return args
}

// String produces a parseable string from this query.
func (p *Query) String() string {
	return query_string(*p)
}

// Eval evaluates the given query in the given environment.
func (p *Query) Eval(row uint, env map[string]tr.Column[bls12_377.Element]) fr.Element {
	switch p.op {
	case qVAR:
		if col, ok := env[p.name]; ok {
			return col.Get(int(row)).Element
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

type binary_op = func(fr.Element, fr.Element) fr.Element
type nary_op = func([]fr.Element) fr.Element

func eval_binary(row uint, env map[string]tr.Column[bls12_377.Element], lhs Query, rhs Query, fn binary_op) fr.Element {
	// Evaluate left-hand side
	lv := lhs.Eval(row, env)
	// Evaluate right-hand side
	rv := rhs.Eval(row, env)
	// Performan binary operation
	return fn(lv, rv)
}

func eval_nary(row uint, env map[string]tr.Column[bls12_377.Element], args []Query, fn nary_op) fr.Element {
	vals := make([]fr.Element, len(args))
	// Evaluate arguments
	for i, arg := range args {
		vals[i] = arg.Eval(row, env)
	}
	//
	return fn(vals)
}

func eval_eq(lhs fr.Element, rhs fr.Element) fr.Element {
	// Perform comparison
	if lhs.Cmp(&rhs) == 0 {
		return fr.NewElement(0)
	}
	//
	return fr.One()
}

func eval_neq(lhs fr.Element, rhs fr.Element) fr.Element {
	// Perform comparison
	if lhs.Cmp(&rhs) != 0 {
		return fr.NewElement(0)
	}
	//
	return fr.One()
}

func eval_lt(lhs fr.Element, rhs fr.Element) fr.Element {
	// Perform comparison
	if lhs.Cmp(&rhs) < 0 {
		return fr.NewElement(0)
	}
	//
	return fr.One()
}

func eval_lteq(lhs fr.Element, rhs fr.Element) fr.Element {
	// Perform comparison
	if lhs.Cmp(&rhs) <= 0 {
		return fr.NewElement(0)
	}
	//
	return fr.One()
}

func eval_or(vals []fr.Element) fr.Element {
	for _, v := range vals {
		if v.IsZero() {
			// Success
			return v
		}
	}
	// Fail
	return fr.One()
}

func eval_and(vals []fr.Element) fr.Element {
	//
	for _, v := range vals {
		if !v.IsZero() {
			// Fail
			return v
		}
	}
	// Success
	return fr.NewElement(0)
}

func eval_add(vals []fr.Element) fr.Element {
	val := fr.NewElement(0)
	//
	for _, v := range vals {
		val.Add(&val, &v)
	}
	//
	return val
}

func eval_mul(vals []fr.Element) fr.Element {
	val := fr.NewElement(1)
	//
	for _, v := range vals {
		val.Mul(&val, &v)
	}
	//
	return val
}

func eval_sub(vals []fr.Element) fr.Element {
	val := vals[0]
	//
	for _, v := range vals[1:] {
		val.Sub(&val, &v)
	}
	//
	return val
}

func query_string(p Query, braces ...int) string {
	var str string
	//
	switch p.op {
	case qVAR:
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

func query_strings(op string, spacing bool, queries []Query, braces ...int) string {
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
