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
	n := uint(len(e.Args))
	// Evaluate first argument
	val, err := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := uint(1); err == nil && i < n; i++ {
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
