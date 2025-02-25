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
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// ApplyConstantPropagation simply collapses constant expressions down to single
// values.  For example, "(+ 1 2)" would be collapsed down to "3".
func constantPropagationForTerm(e Term, schema sc.Schema) Term {
	switch e := e.(type) {
	case *Add:
		return constantPropagationForAdd(e.Args, schema)
	case *Constant:
		return e
	case *ColumnAccess:
		return e
	case *Exp:
		return constantPropagationForExp(e.Arg, e.Pow, schema)
	case *Mul:
		return constantPropagationForMul(e.Args, schema)
	case *Norm:
		return constantPropagationForNorm(e.Arg, schema)
	case *Sub:
		return constantPropagationForSub(e.Args, schema)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func constantPropagationForAdd(es []Term, schema sc.Schema) Term {
	sum := fr.NewElement(0)
	count := 0
	rs := make([]Term, len(es))
	//
	for i, e := range es {
		rs[i] = constantPropagationForTerm(e, schema)
		// Check for constant
		c, ok := rs[i].(*Constant)
		// Try to continue sum
		if ok {
			sum.Add(&sum, &c.Value)
			// Increase count of constants
			count++
		}
	}
	// Check if constant
	if count == len(es) {
		// Propagate constant
		return &Constant{sum}
	} else if count > 1 {
		rs = mergeConstants(sum, rs)
	}
	// Done
	return &Add{rs}
}

func constantPropagationForExp(arg Term, pow uint64, schema sc.Schema) Term {
	arg = constantPropagationForTerm(arg, schema)
	//
	if c, ok := arg.(*Constant); ok {
		var val fr.Element
		// Clone value
		val.Set(&c.Value)
		// Compute exponent (in place)
		util.Pow(&val, pow)
		// Done
		return &Constant{val}
	}
	//
	return &Exp{arg, pow}
}

func constantPropagationForMul(es []Term, schema sc.Schema) Term {
	one := fr.NewElement(1)
	prod := one
	rs := make([]Term, len(es))
	ones := 0
	consts := 0
	//
	for i, e := range es {
		rs[i] = constantPropagationForTerm(e, schema)
		// Check for constant
		c, ok := rs[i].(*Constant)
		//
		if ok && c.Value.IsZero() {
			// No matter what, outcome is zero.
			return &Constant{c.Value}
		} else if ok && c.Value.IsOne() {
			ones++
			consts++
			rs[i] = nil
		} else if ok {
			// Continue building constant
			prod.Mul(&prod, &c.Value)
			//
			consts++
		}
	}
	// Check if constant
	if consts == len(es) {
		return &Constant{prod}
	} else if ones > 0 {
		rs = util.RemoveMatching[Term](rs, func(item Term) bool { return item == nil })
	}
	// Sanity check what's left.
	if len(rs) == 1 {
		return rs[0]
	} else if consts-ones > 1 {
		// Combine constants
		rs = mergeConstants(prod, rs)
	}
	// Done
	return &Mul{rs}
}

func constantPropagationForNorm(arg Term, schema sc.Schema) Term {
	arg = constantPropagationForTerm(arg, schema)
	//
	if c, ok := arg.(*Constant); ok {
		var val fr.Element
		// Clone value
		val.Set(&c.Value)
		// Normalise (in place)
		if !val.IsZero() {
			val.SetOne()
		}
		// Done
		return &Constant{val}
	}
	//
	return &Norm{arg}
}

func constantPropagationForSub(es []Term, schema sc.Schema) Term {
	var sum fr.Element
	// count non-constant terms
	count := 0
	rs := make([]Term, len(es))
	//
	for i, e := range es {
		rs[i] = constantPropagationForTerm(e, schema)
		// Check for constant
		c, ok := rs[i].(*Constant)
		// Try to continue sum
		if ok && i == 0 {
			sum = c.Value
		} else if ok {
			sum.Sub(&sum, &c.Value)
		} else {
			count++
		}
	}
	// Check for any non-constant terms
	if count == 0 {
		// Propagate constant
		return &Constant{sum}
	} else if count != len(es) {
		// Apply simplifications
		rs = removeNonLeadingZeros(rs)
	}
	// Sanity check what's left
	if len(rs) == 1 {
		return rs[0]
	}
	// Done
	return &Sub{rs}
}

// Remove all zeros which don't arise at the beginning.
func removeNonLeadingZeros(terms []Term) []Term {
	return util.RemoveMatchingIndexed(terms, func(i int, v Term) bool {
		if c, ok := v.(*Constant); ok && i != 0 && c.Value.IsZero() {
			return true
		}
		//
		return false
	})
}

// Replace all constants within a given sequence of expressions with a single
// constant (whose value has been precomputed from those constants).  The new
// value replaces the first constant in the list.
func mergeConstants(constant fr.Element, es []Term) []Term {
	j := 0
	first := true
	//
	for i := range es {
		// Check for constant
		if _, ok := es[i].(*Constant); ok && first {
			es[j] = &Constant{constant}
			first = false
			j++
		} else if !ok {
			// Retain non-constant expression
			es[j] = es[i]
			j++
		}
	}
	// Return slice
	return es[0:j]
}
