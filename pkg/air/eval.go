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
package air

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
)

func evalAtTerm(e Term, k int, trace tr.Trace) fr.Element {
	switch e := e.(type) {
	case *Add:
		return evalAtAdd(e, k, trace)
	case *Constant:
		return e.Value
	case *ColumnAccess:
		return trace.Column(e.Column).Get(k + e.Shift)
	case *Mul:
		return evalAtMul(e, k, trace)
	case *Sub:
		return evalAtSub(e, k, trace)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}

func evalAtAdd(e *Add, k int, trace tr.Trace) fr.Element {
	// Evaluate first argument
	val := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith := evalAtTerm(e.Args[i], k, trace)
		val.Add(&val, &ith)
	}
	// Done
	return val
}

func evalAtMul(e *Mul, k int, trace tr.Trace) fr.Element {
	n := uint(len(e.Args))
	// Evaluate first argument
	val := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := uint(1); i < n; i++ {
		var ith fr.Element
		// Can short-circuit evaluation?
		if val.IsZero() {
			return val
		}
		// No
		ith = evalAtTerm(e.Args[i], k, trace)
		//
		val.Mul(&val, &ith)
	}
	// Done
	return val
}

func evalAtSub(e *Sub, k int, trace tr.Trace) fr.Element {
	// Evaluate first argument
	val := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith := evalAtTerm(e.Args[i], k, trace)
		val.Sub(&val, &ith)
	}
	// Done
	return val
}
