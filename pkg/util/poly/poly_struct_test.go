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
package poly

import (
	"testing"

	"github.com/consensys/go-corset/pkg/util/source"
)

func Test_PolyStruct_0(t *testing.T) {
	checkEquiv(t, "0", "0-0", "1-1", "2-1-1", "x-x", "(2*x)-(2*x)", "(2*x)-(x+x)", "(2*x)-x-x", "(x+y)-(x+y)")
}

func Test_PolyStruct_1(t *testing.T) {
	checkEquiv(t, "1", "0+1", "1+0", "2-1", "3-2", "0+(2-1)", "1+(x-x)", "(1+x+y)-(x+y)")
}
func Test_PolyStruct_x(t *testing.T) {
	checkEquiv(t, "x+(x-x)", "x", "1*x", "(2*x)-x", "y+(x-y)")
}
func Test_PolyStruct_xp1(t *testing.T) {
	checkEquiv(t, "x+1", "1+x", "(2-1)+x", "0+x+1", "0+1+x", "1+x+0")
}

func Test_PolyStruct_2x(t *testing.T) {
	checkEquiv(t, "x+x", "2*x", "x+x+(x-x)", "(3*x)-x")
}

func Test_PolyStruct_xpy(t *testing.T) {
	checkEquiv(t, "x+y", "y+x", "(2*y)+(x-y)")
}
func Test_PolyStruct_2xpy(t *testing.T) {
	checkEquiv(t, "2*(x+y)", "x+x+y+y", "x+y+x+y", "(2*x)+y+y", "(2*x)+(2*y)")
}
func Test_PolyStruct_2xxpxpy(t *testing.T) {
	checkEquiv(t, "(2*x*x)+x+y", "x+(2*x*x)+y", "x+y+(2*x*x)")
}

// =========================================================================================

func checkEquiv(t *testing.T, terms ...string) {
	var (
		ts   = make([]*ArrayPoly[string], len(terms))
		errs []source.SyntaxError
	)
	//
	for i, term := range terms {
		if ts[i], errs = Parse(term); len(errs) > 0 {
			panic(errs[0].Message())
		}
	}
	//
	for i := range len(ts) {
		l, r := ts[0], ts[i]
		// Check polynomials are equivalent
		if !l.Equal(r) {
			t.Errorf("polynomials not equivalent: %s vs %s", String(l, id), String(r, id))
		}
	}
}

func id(x string) string {
	return x
}
