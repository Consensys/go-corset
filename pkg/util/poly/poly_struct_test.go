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
	"math/big"
	"testing"
)

var ONE *ArrayPoly[string]

// NOTE: the goal of this testsuite will be to fully test that polynomials
// arrived at in different (but equalivent) ways are structurally identical. For
// example, adding 1 to x should be identical to adding x to 1.  Likewise,
// adding 2 to x then subtracting 1 should be identical to adding 1 to x as
// well.
//
// At the moment, the array polynomial does not exhibit these qualities and will
// need to be aggressively updated to implement them.

func Test_PolyStruct_01(t *testing.T) {
	var (
		lhs ArrayPoly[string]
		rhs ArrayPoly[string]
	)
	// No monomials should be equivalent to zero
	lhs.Set()
	//
	lhs = *lhs.Add(ONE)
	//
	rhs.Set(NewMonomial[string](*big.NewInt(1)))
	//
	assertEqual(t, &lhs, &rhs)
}

// =========================================================================================

func assertEqual(t *testing.T, lhs *ArrayPoly[string], rhs *ArrayPoly[string]) {
	if !lhs.Equal(rhs) {
		t.Errorf("polynomials not equals: %s vs %s", String(lhs, id), String(rhs, id))
	}
}

func id(x string) string {
	return x
}

func init() {
	ONE = ONE.Set(NewMonomial[string](*big.NewInt(1)))
}
