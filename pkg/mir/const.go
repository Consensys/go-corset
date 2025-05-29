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

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// Constant Propagation simply collapses constant expressions down to single
// values.  For example, "(+ 1 2)" would be collapsed down to "3".  This is then
// progagated throughout an expression, so that e.g. "(+ X (+ 1 2))" becomes "(+
// X 3)"", etc.  There is also an option to retain casts, or not.
func constantPropagationForTerm(e Term, casts bool, schema sc.Schema) Term {
	switch e := e.(type) {
	case *Add:
		return constantPropagationForAdd(e.Args, casts, schema)
	case *Cast:
		return constantPropagationForCast(e.Arg, casts, e.BitWidth, schema)
	case *Constant:
		return e
	case *ColumnAccess:
		return e
	case *Exp:
		return constantPropagationForExp(e.Arg, casts, e.Pow, schema)
	case *Mul:
		return constantPropagationForMul(e.Args, casts, schema)
	case *Norm:
		return constantPropagationForNorm(e.Arg, casts, schema)
	case *Sub:
		return constantPropagationForSub(e.Args, casts, schema)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func constantPropagationForAdd(terms []Term, casts bool, schema sc.Schema) Term {
	terms = constantPropagation(terms, addBinOp, frZERO, casts, schema)
	// Flatten any nested sums
	terms = util.Flatten(terms, flattern[*Add])
	// Remove any zeros
	terms = util.RemoveMatching(terms, isZero)
	// Check anything left
	switch len(terms) {
	case 0:
		return &Constant{frZERO}
	case 1:
		return terms[0]
	default:
		// Done
		return &Add{terms}
	}
}

func constantPropagationForCast(arg Term, casts bool, bitwidth uint, schema sc.Schema) Term {
	var bound fr.Element = fr.NewElement(2)
	// Determine bound for static type check
	util.Pow(&bound, uint64(bitwidth))
	// Propagate constants in the argument
	arg = constantPropagationForTerm(arg, casts, schema)
	//
	if c, ok := arg.(*Constant); ok && c.Value.Cmp(&bound) < 0 {
		// Done
		return c
	} else if ok {
		// Type failure
		panic(fmt.Sprintf("type cast failure (have %s with expected bitwidth %d)", c.Value.String(), bitwidth))
	} else if !casts {
		// elide cast
		return arg
	}
	//
	return &Cast{arg, bitwidth}
}

func constantPropagationForExp(arg Term, casts bool, pow uint64, schema sc.Schema) Term {
	arg = constantPropagationForTerm(arg, casts, schema)
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

func constantPropagationForMul(terms []Term, casts bool, schema sc.Schema) Term {
	terms = constantPropagation(terms, mulBinOp, frONE, casts, schema)
	// Flatten any nested products
	terms = util.Flatten(terms, flattern[*Mul])
	// Check for zero
	if util.ContainsMatching(terms, isZero) {
		// Yes, is zero
		return &Constant{fr.NewElement(0)}
	}
	// Remove any ones
	terms = util.RemoveMatching(terms, isOne)
	// Check whats left
	switch len(terms) {
	case 0:
		return &Constant{frONE}
	case 1:
		return terms[0]
	default:
		// Done
		return &Mul{terms}
	}
}

func constantPropagationForNorm(arg Term, casts bool, schema sc.Schema) Term {
	arg = constantPropagationForTerm(arg, casts, schema)
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

func constantPropagationForSub(terms []Term, casts bool, schema sc.Schema) Term {
	lhs := constantPropagationForTerm(terms[0], casts, schema)
	// Subtraction is harder to optimise for.  What we do is view "a - b - c" as
	// "a - (b+c)", and optimise the right-hand side as though it were addition.
	rhs := constantPropagationForAdd(terms[1:], casts, schema)
	// Check what's left
	lc, l_const := lhs.(*Constant)
	rc, r_const := rhs.(*Constant)
	ra, r_add := rhs.(*Add)
	r_zero := isZero(rhs)
	//
	switch {
	case r_zero:
		// Right-hand side zero, nothing to subtract.
		return lhs
	case l_const && r_const:
		// Both sides constant, result is constant.
		c := lc.Value
		c = *c.Sub(&c, &rc.Value)
		//
		return &Constant{c}
	case l_const && r_add:
		nterms := util.Prepend(lhs, ra.Args)
		// if rhs has constant, subtract it.
		if rc, ok := findConstant(ra.Args); ok {
			c := lc.Value
			c = *c.Sub(&c, &rc)
			nterms = mergeConstants(c, nterms)
		}
		//
		return &Sub{nterms}
	case r_add:
		// Default case, recombine.
		return &Sub{util.Prepend(lhs, ra.Args)}
	default:
		return &Sub{[]Term{lhs, rhs}}
	}
}

func findConstant(terms []Term) (fr.Element, bool) {
	for _, t := range terms {
		if c, ok := t.(*Constant); ok {
			return c.Value, true
		}
	}
	//
	return frZERO, false
}

type binop func(fr.Element, fr.Element) fr.Element

// General purpose constant propagation mechanism.  This reduces all terms to
// constants (where possible) and combines terms according to a given
// combinator.
func constantPropagation(terms []Term, fn binop, acc fr.Element, casts bool, schema sc.Schema) []Term {
	// Count how many terms reduced to constants.
	count := 0
	nterms := make([]Term, len(terms))
	// Propagate through all children
	for i, e := range terms {
		nterms[i] = constantPropagationForTerm(e, casts, schema)
		// Check for constant
		c, ok := nterms[i].(*Constant)
		// Try to continue sum
		if ok {
			// Apply combinator
			acc = fn(acc, c.Value)
			// Increase count of constants
			count++
		}
	}
	// Merge all constants
	return mergeConstants(acc, nterms)
}

// Replace all constants within a given sequence of expressions with a single
// constant (whose value has been precomputed from those constants).  The new
// value replaces the first constant in the list.
func mergeConstants(constant fr.Element, terms []Term) []Term {
	j := 0
	first := true
	//
	for i := range terms {
		// Check for constant
		if _, ok := terms[i].(*Constant); ok && first {
			terms[j] = &Constant{constant}
			first = false
			j++
		} else if !ok {
			// Retain non-constant expression
			terms[j] = terms[i]
			j++
		}
	}
	// Return slice
	return terms[0:j]
}

func addBinOp(lhs fr.Element, rhs fr.Element) fr.Element {
	return *lhs.Add(&lhs, &rhs)
}

func mulBinOp(lhs fr.Element, rhs fr.Element) fr.Element {
	return *lhs.Mul(&lhs, &rhs)
}

func flattern[T Term](term Term) []Term {
	if _, ok := term.(T); ok {
		switch t := term.(type) {
		case *Add:
			return t.Args
		case *Sub:
			return t.Args
		case *Mul:
			return t.Args
		default:
			panic("unreachable")
		}
	}
	//
	return nil
}
