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
package agnostic

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/poly"
)

var xRegId = schema.NewRegisterId(0)
var yRegId = schema.NewRegisterId(1)

func x(coeff int64) Polynomial {
	return Var(xRegId, coeff)
}

func y(coeff int64) Polynomial {
	return Var(yRegId, coeff)
}

// Var constructs a polynomial representing a given variable multiplied by a
// given coefficient.
func Var(id schema.RegisterId, coeff int64) Polynomial {
	var (
		p Polynomial
		c = *big.NewInt(coeff)
	)
	//
	p = p.Set(poly.NewMonomial(c, id))
	//
	return p
}

func Test_Poly_01(t *testing.T) {
	// x where x:u8 requires 8 bits
	check(t, 8, x(1), 8)
}
func Test_Poly_02(t *testing.T) {
	// 2x where x:u8 requires 9 bits
	check(t, 9, x(2), 8)
}
func Test_Poly_03(t *testing.T) {
	// x+y where x:u8, y:u8 requires 9 bits
	check(t, 9, x(1).Add(y(1)), 8, 8)
}
func Test_Poly_03a(t *testing.T) {
	// x+2y where x:u8, y:u8 requires 10 bits
	check(t, 10, x(1).Add(y(2)), 8, 8)
}
func Test_Poly_04(t *testing.T) {
	// x+y where x:u16, y:u8 requires 17 bits
	check(t, 17, x(1).Add(y(1)), 16, 8)
}
func Test_Poly_05(t *testing.T) {
	// x-y where x:u8, y:u8 requires 9 bits
	check(t, 9, x(1).Sub(y(1)), 8, 8)
}
func Test_Poly_06(t *testing.T) {
	// x-y where x:u16, y:u8 requires 16 bits
	check(t, 16, x(1).Sub(y(1)), 16, 8)
}

func check(t *testing.T, bitwidth uint, p Polynomial, widths ...uint) {
	var regs = make([]schema.Register, len(widths))
	//
	for i := range regs {
		regs[i] = schema.NewInputRegister("?", widths[i])
	}
	// Determine computed bitwidth
	actual := WidthOfPolynomial(p, regs)
	// Check for any differences
	if actual != bitwidth {
		err := fmt.Sprintf("invalid bitwidth (expected %d got %d)", bitwidth, actual)
		t.Error(err)
	}
}
