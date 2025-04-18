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
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

var frZERO fr.Element = fr.NewElement(0)
var frONE fr.Element = fr.NewElement(1)

func evalAtConstraint(e Constraint, k int, trace tr.Trace) (fr.Element, error) {
	//
	for _, disjunct := range e.conjuncts {
		val, _, err := evalAtDisjunction(disjunct, k, trace)
		//
		if err != nil {
			return frONE, err
		} else if !val.IsZero() {
			// Failure
			return val, nil
		}
	}
	// Success
	return frZERO, nil
}

func evalAtDisjunction(e Disjunction, k int, trace tr.Trace) (fr.Element, uint, error) {
	//
	for i, eq := range e.atoms {
		val, err := evalAtEquation(eq, k, trace)
		//
		if err != nil {
			return frONE, uint(i), err
		} else if val.IsZero() {
			// Success
			return val, uint(i), nil
		}
	}
	// Failure
	return frONE, uint(0), nil
}

func evalAtEquation(e Equation, k int, trace tr.Trace) (fr.Element, error) {
	lhs, err1 := evalAtTerm(e.lhs, k, trace)
	rhs, err2 := evalAtTerm(e.rhs, k, trace)
	// error check
	if err1 != nil {
		return frONE, err1
	} else if err2 != nil {
		return frONE, err2
	}
	// perform comparison
	c := lhs.Cmp(&rhs)
	//
	switch e.kind {
	case EQUALS:
		if c == 0 {
			return frZERO, nil
		}
	case NOT_EQUALS:
		if c != 0 {
			return frZERO, nil
		}
	case LESS_THAN:
		if c < 0 {
			return frZERO, nil
		}
	case LESS_THAN_EQUALS:
		if c <= 0 {
			return frZERO, nil
		}
	case GREATER_THAN_EQUALS:
		if c >= 0 {
			return frZERO, nil
		}
	case GREATER_THAN:
		if c > 0 {
			return frZERO, nil
		}
	}
	// failure
	return frONE, nil
}

func evalAtTerm(e Term, k int, trace tr.Trace) (fr.Element, error) {
	switch e := e.(type) {
	case *Add:
		return evalAtAdd(e, k, trace)
	case *Cast:
		return evalAtCast(e, k, trace)
	case *Constant:
		return e.Value, nil
	case *ColumnAccess:
		return trace.Column(e.Column).Get(k + e.Shift), nil
	case *Exp:
		return evalAtExp(e, k, trace)
	case *Mul:
		return evalAtMul(e, k, trace)
	case *Norm:
		return evalAtNormalise(e, k, trace)
	case *Sub:
		return evalAtSub(e, k, trace)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func evalAtAdd(e *Add, k int, trace tr.Trace) (fr.Element, error) {
	// Evaluate first argument
	val, err := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(e.Args); i++ {
		var ith fr.Element
		// Evaluate ith argument
		ith, err = evalAtTerm(e.Args[i], k, trace)
		val.Add(&val, &ith)
	}
	// Done
	return val, err
}

func evalAtCast(e *Cast, k int, tr trace.Trace) (fr.Element, error) {
	var c big.Int
	//
	cast := e.Range()
	// Check whether argument evaluates to zero or not.
	val, err := evalAtTerm(e.Arg, k, tr)
	// Extract big integer from field element
	val.BigInt(&c)
	// Dynamic cast check
	if err == nil && !cast.Contains(&c) {
		// Construct error
		err = fmt.Errorf("cast failure (value %s not a u%d)", val.String(), e.BitWidth)
	}
	// All good
	return val, err
}

func evalAtExp(e *Exp, k int, tr trace.Trace) (fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	val, err := evalAtTerm(e.Arg, k, tr)
	// Compute exponent
	util.Pow(&val, e.Pow)
	// Done
	return val, err
}

func evalAtMul(e *Mul, k int, trace tr.Trace) (fr.Element, error) {
	// Evaluate first argument
	val, err := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(e.Args); i++ {
		var ith fr.Element
		// Can short-circuit evaluation?
		if val.IsZero() {
			return val, nil
		}
		// No
		ith, err = evalAtTerm(e.Args[i], k, trace)
		val.Mul(&val, &ith)
	}
	// Done
	return val, err
}

// EvalAt evaluates the normalisation of some expression by first evaluating
// that expression.  Then, zero is returned if the result is zero; otherwise one
// is returned.
func evalAtNormalise(e *Norm, k int, tr trace.Trace) (fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	val, err := evalAtTerm(e.Arg, k, tr)
	// Normalise value (if necessary)
	if !val.IsZero() {
		val.SetOne()
	}
	// Done
	return val, err
}

func evalAtSub(e *Sub, k int, trace tr.Trace) (fr.Element, error) {
	// Evaluate first argument
	val, err := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(e.Args); i++ {
		var ith fr.Element
		// Evaluate ith argument
		ith, err = evalAtTerm(e.Args[i], k, trace)
		val.Sub(&val, &ith)
	}
	// Done
	return val, err
}
