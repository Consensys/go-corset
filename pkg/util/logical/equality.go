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
package logical

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
)

// Equality represents a fundamental atom which is either an Equality (=) or a
// non-Equality (≠) between a variable and either a variable or a constant.
type Equality[I array.Comparable[I]] struct {
	// Sign indicates whether this is an Equality (==) or a non-Equality (!=).
	Sign bool
	// Left variable
	Left I
	// Right variable or constant
	Right util.Union[I, big.Int]
}

// Equals returns an equality over two variables (i.e. l = r)
func Equals[I array.Comparable[I]](l I, r I) Equality[I] {
	// Ensure lowest variable always on the left-hand side.
	if l.Cmp(r) > 0 {
		l, r = r, l
	}
	//
	return Equality[I]{true, l, util.Union1[I, big.Int](r)}
}

// EqualsConst returns an equality over a variable and a constant
func EqualsConst[I array.Comparable[I]](l I, r big.Int) Equality[I] {
	return Equality[I]{true, l, util.Union2[I, big.Int](r)}
}

// NotEquals returns a non-Equality over two variables (i.e. l ≠ r)
func NotEquals[I array.Comparable[I]](l I, r I) Equality[I] {
	// Ensure lowest variable always on the left-hand side.
	if l.Cmp(r) > 0 {
		l, r = r, l
	}
	//
	return Equality[I]{false, l, util.Union1[I, big.Int](r)}
}

// NotEqualsConst returns an equality over a variable and a constant
func NotEqualsConst[I array.Comparable[I]](l I, r big.Int) Equality[I] {
	return Equality[I]{false, l, util.Union2[I, big.Int](r)}
}

// Cmp implementation for Comparable interface
func (p Equality[I]) Cmp(o Equality[I]) int {
	if p.Sign != o.Sign {
		if p.Sign {
			return -1
		}
		//
		return 1
	}
	//
	if c := p.Left.Cmp(o.Left); c != 0 {
		return c
	}
	//
	return cmpRhs(p.Right, o.Right)
}

// Contradicts determines whether two equalities contradict each other.  There
// are only a few ways this can happen.
func (p Equality[I]) Contradicts(o Equality[I]) bool {
	var (
		pEqConst = p.Sign && p.Right.HasSecond()
		oEqConst = o.Sign && o.Right.HasSecond()
	)
	//
	if p.Cmp(o) == 0 {
		// p && p ==> T
		return false
	} else if p.Cmp(o.Negate()) == 0 {
		// p && !p ==> _|_
		return true
	} else if pEqConst && oEqConst {
		// x=c1 && x=c2 -> _|_
		pRight := p.Right.Second()
		oRight := o.Right.Second()

		return p.Left.Cmp(o.Left) == 0 && pRight.Cmp(&oRight) != 0
	}
	//
	return false
}

// Is implementation of Atom interface
func (p Equality[I]) Is(truth bool) bool {
	// x == x ==> truth
	// x != x ==> false
	return p.Sign == truth && p.Right.HasFirst() && p.Left.Cmp(p.Right.First()) == 0
}

// Negate this Equality (i.e. turn it from "==" to "!=" or vice-versa)
func (p Equality[I]) Negate() Equality[I] {
	return Equality[I]{!p.Sign, p.Left, p.Right}
}

// Subsumes checks whether this Equality subsumes the other
func (p Equality[I]) Subsumes(o Equality[I]) bool {
	if p.Cmp(o) == 0 {
		return true
	} else if !p.Sign || o.Sign {
		// (i) x≠? does not subsume anything
		// (ii) nothing subsumes x=?
		return false
	} else if p.Left.Cmp(o.Left) == 0 && p.Right.HasSecond() && o.Right.HasSecond() {
		// e.g. x=1 subsumes x≠2
		return true
	}
	//
	return false
}

func (p Equality[I]) String(mapping func(I) string) string {
	var (
		l = mapping(p.Left)
		r string
	)
	//
	if p.Right.HasFirst() {
		r = mapping(p.Right.First())
	} else {
		bi := p.Right.Second()
		r = bi.String()
	}
	//
	if p.Sign {
		return fmt.Sprintf("%s=%s", l, r)
	}
	//
	return fmt.Sprintf("%s≠%s", l, r)
}

func cmpRhs[I array.Comparable[I]](l util.Union[I, big.Int], r util.Union[I, big.Int]) int {
	switch {
	case l.HasFirst() && r.HasSecond():
		return -1
	case l.HasSecond() && r.HasFirst():
		return 1
	case l.HasFirst():
		return l.First().Cmp(r.First())
	default:
		lbi := l.Second()
		rbi := r.Second()

		return lbi.Cmp(&rbi)
	}
}
