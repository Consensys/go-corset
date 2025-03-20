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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
)

const qVAR = 0
const qNUM = 1
const qOR = 2
const qAND = 3
const qEQ = 4
const qNEQ = 5
const qLT = 6
const qLTEQ = 7

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

// Matches determines whether or not this query holds on the given row, assuming
// the given column values.
func (p *Query) Matches(row uint, env map[string]tr.Column) (bool, error) {
	val, err := p.Eval(row, env)
	//
	if err != nil {
		return false, err
	}
	//
	return val.IsZero(), nil
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
	args := make([]Query, 1+len(queries))
	args[0] = *p
	//
	for i := range queries {
		args[i+1] = *queries[i]
	}
	//
	return &Query{qOR, args, fr.Element{}, ""}
}

// And constructs a conjunction of queries.
func (p *Query) And(queries ...*Query) *Query {
	args := make([]Query, 1+len(queries))
	args[0] = *p
	//
	for i := range queries {
		args[i+1] = *queries[i]
	}
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

// Eval evaluates the given query in the given environment.
func (p *Query) Eval(row uint, env map[string]tr.Column) (fr.Element, error) {
	switch p.op {
	case qVAR:
		if col, ok := env[p.name]; ok {
			return col.Get(int(row)), nil
		}
		// error
		return fr.One(), fmt.Errorf("unknown column \"%s\"", p.name)
	case qNUM:
		return p.number, nil
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
	default:
		return fr.One(), fmt.Errorf("unknown operator (%d)", p.op)
	}
}

type binary_op = func(fr.Element, fr.Element) fr.Element
type nary_op = func([]fr.Element) fr.Element

func eval_binary(row uint, env map[string]tr.Column, lhs Query, rhs Query, fn binary_op) (fr.Element, error) {
	var (
		lv, rv fr.Element
		err    error
	)
	// Evaluate left-hand side
	if lv, err = lhs.Eval(row, env); err != nil {
		return lv, err
	}
	// Evaluate right-hand side
	if rv, err = rhs.Eval(row, env); err != nil {
		return rv, err
	}
	// Performan binary operation
	return fn(lv, rv), nil
}

func eval_nary(row uint, env map[string]tr.Column, args []Query, fn nary_op) (fr.Element, error) {
	var (
		vals = make([]fr.Element, len(args))
		err  error
	)
	// Evaluate arguments
	for i, arg := range args {
		if vals[i], err = arg.Eval(row, env); err != nil {
			return vals[i], err
		}
	}
	//
	return fn(vals), nil
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
