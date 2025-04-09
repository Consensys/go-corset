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
package hir

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

func evalAtTerm(e Term, k int, trace tr.Trace) ([]fr.Element, error) {
	switch e := e.(type) {
	case *Add:
		return evalAtAdd(e, k, trace)
	case *Cast:
		return evalAtCast(e, k, trace)
	case *Constant:
		return []fr.Element{e.Value}, nil
	case *ColumnAccess:
		val := trace.Column(e.Column).Get(k + e.Shift)
		return []fr.Element{val}, nil
	case *Equation:
		return evalAtEquation(e, k, trace)
	case *Exp:
		return evalAtExp(e, k, trace)
	case *IfZero:
		return evalAtIfZero(e, k, trace)
	case *LabelledConstant:
		return []fr.Element{e.Value}, nil
	case *List:
		return evalAtList(e, k, trace)
	case *Mul:
		return evalAtMul(e, k, trace)
	case *Norm:
		return evalAtNormalise(e, k, trace)
	case *Sub:
		return evalAtSub(e, k, trace)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func evalAtAdd(e *Add, k int, tr trace.Trace) ([]fr.Element, error) {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Add(&l, &r); return l }
	return evalAtTerms(k, tr, e.Args, fn)
}

func evalAtCast(e *Cast, k int, tr trace.Trace) ([]fr.Element, error) {
	var c big.Int
	//
	cast := e.Range()
	// Check whether argument evaluates to zero or not.
	vals, err := evalAtTerm(e.Arg, k, tr)
	//
	for i := 0; err == nil && i < len(vals); i++ {
		val := vals[i]
		// Extract big integer from field element
		val.BigInt(&c)
		// Dynamic cast check
		if !cast.Contains(&c) {
			// Construct error
			err = fmt.Errorf("cast failure (value %s not a u%d)", val.String(), e.BitWidth)
		}
	}
	// All good
	return vals, err
}

func evalAtEquation(e *Equation, k int, tr trace.Trace) ([]fr.Element, error) {
	args := []Term{e.Lhs, e.Rhs}
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Sub(&l, &r); return l }
	//
	return evalAtTerms(k, tr, args, fn)
}

func evalAtExp(e *Exp, k int, tr trace.Trace) ([]fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	vals, err := evalAtTerm(e.Arg, k, tr)
	//
	for i := range vals {
		// Compute exponent
		util.Pow(&vals[i], e.Pow)
	}
	// Done
	return vals, err
}

func evalAtIfZero(e *IfZero, k int, tr trace.Trace) ([]fr.Element, error) {
	vals := make([]fr.Element, 0)
	// Evaluate condition
	conditions, err := evalAtTerm(e.Condition, k, tr)
	// Check all results
	for i := 0; err == nil && i < len(conditions); i++ {
		var (
			vs   []fr.Element
			cond = conditions[i]
		)
		//q
		if cond.IsZero() && e.TrueBranch != nil {
			vs, err = evalAtTerm(e.TrueBranch, k, tr)
		} else if !cond.IsZero() && e.FalseBranch != nil {
			vs, err = evalAtTerm(e.FalseBranch, k, tr)
		}
		//
		vals = append(vals, vs...)
	}

	return vals, err
}

func evalAtMul(e *Mul, k int, tr trace.Trace) ([]fr.Element, error) {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Mul(&l, &r); return l }
	return evalAtTerms(k, tr, e.Args, fn)
}

func evalAtNormalise(e *Norm, k int, tr trace.Trace) ([]fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	vals, err := evalAtTerm(e.Arg, k, tr)
	// Normalise value (if necessary)
	for i := range vals {
		if !vals[i].IsZero() {
			vals[i].SetOne()
		}
	}
	// Done
	return vals, err
}

func evalAtList(e *List, k int, tr trace.Trace) ([]fr.Element, error) {
	var vals []fr.Element
	//
	for _, arg := range e.Args {
		if vs, err := evalAtTerm(arg, k, tr); err != nil {
			// error case
			return nil, err
		} else {
			vals = append(vals, vs...)
		}
	}
	//
	return vals, nil
}

func evalAtSub(e *Sub, k int, tr trace.Trace) ([]fr.Element, error) {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Sub(&l, &r); return l }
	return evalAtTerms(k, tr, e.Args, fn)
}

// EvalExprsAt evaluates all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func evalAtTerms(k int, tr trace.Trace, terms []Term,
	fn func(fr.Element, fr.Element) fr.Element) ([]fr.Element, error) {
	// Evaluate first argument.
	vals, err := evalAtTerm(terms[0], k, tr)
	// Continue evaluating the rest.
	for i := 1; err == nil && i < len(terms); i++ {
		var vs []fr.Element
		vs, err = evalAtTerm(terms[i], k, tr)
		vals = evalAtTermsApply(vals, vs, fn)
	}
	// Done.
	return vals, err
}

// Perform a vector operation using the given primitive operator "fn".
func evalAtTermsApply(lhs []fr.Element, rhs []fr.Element, fn func(fr.Element, fr.Element) fr.Element) []fr.Element {
	if len(rhs) == 1 {
		// Optimise for common case.
		for i, ith := range lhs {
			lhs[i] = fn(ith, rhs[0])
		}

		return lhs
	}
	// Harder case
	vals := make([]fr.Element, 0)
	// Perform n x m operations
	for _, ith := range lhs {
		for _, jth := range rhs {
			vals = append(vals, fn(ith, jth))
		}
	}

	return vals
}
