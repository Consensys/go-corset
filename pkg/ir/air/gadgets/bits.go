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
package gadgets

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ApplyBinaryGadget adds a binarity constraint for a given column in the schema
// which enforces that all values in the given column are either 0 or 1. For a
// column X, this corresponds to the vanishing constraint X * (X-1) == 0.
func ApplyBinaryGadget(column uint, ctx trace.Context, module *air.ModuleBuilder) {
	// Identify target register
	register := module.Register(column)
	// Determine column name
	name := register.Name
	// Construct X
	X := ir.NewRegisterAccess[air.Term](column, 0)
	// Construct X == 0
	X_eq0 := ir.Subtract(X, ir.Const64[air.Term](0))
	// Construct X == 0
	X_eq1 := ir.Subtract(X, ir.Const64[air.Term](1))
	// Construct (X==0) ∨ (X==1)
	X_X_m1 := ir.Product(X_eq0, X_eq1)
	// Done!
	module.AddConstraint(
		air.NewVanishingConstraint(fmt.Sprintf("%s:u1", name), ctx, util.None[int](), X_X_m1))
}

// ApplyBitwidthGadget ensures all values in a given column fit within a given
// number of bits.  This is implemented using a *byte decomposition* which adds
// n columns and a vanishing constraint (where n*8 >= nbits).
func ApplyBitwidthGadget(col uint, nbits uint, selector air.Term, schema *air.Schema) {
	// var (
	// 	// Determine ranges required for the give bitwidth
	// 	ranges = splitColumnRanges(nbits)
	// 	// Identify number of columns required.
	// 	n = uint(len(ranges))
	// )
	// // Sanity check
	// if nbits == 0 {
	// 	panic("zero bitwidth constraint encountered")
	// }
	// // Identify target column
	// column := schema.Columns().Nth(col)
	// // Calculate how many bytes required.
	// es := make([]air.Expr, n)
	// name := column.Name
	// coefficient := fr.NewElement(1)
	// // Add decomposition assignment
	// index := schema.AddAssignment(
	// 	assignment.NewByteDecomposition(name, column.Context, col, nbits))
	// // Construct Columns
	// for i := uint(0); i < n; i++ {
	// 	// Create Column + Constraint
	// 	es[i] = air.NewColumnAccess(index+i, 0).Mul(air.NewConst(coefficient))

	// 	schema.AddRangeConstraint(index+i, 0, ranges[i])
	// 	// Update coefficient
	// 	coefficient.Mul(&coefficient, &ranges[i])
	// }
	// // Construct (X:0 * 1) + ... + (X:n * 2^n)
	// sum := air.Sum(es...)
	// // Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	// X := air.NewColumnAccess(col, 0)
	// eq := selector.Mul(X.Equate(sum))
	// // Construct column name
	// schema.AddVanishingConstraint(fmt.Sprintf("%s:u%d", name, nbits), 0, column.Context, util.None[int](), eq)
	panic("todo")
}

func splitColumnRanges(nbits uint) []fr.Element {
	var (
		n      = nbits / 8
		m      = int64(nbits % 8)
		ranges []fr.Element
		fr256  = fr.NewElement(256)
	)
	//
	if m == 0 {
		ranges = make([]fr.Element, n)
	} else {
		var last fr.Element
		// Most significant column has smaller range.
		ranges = make([]fr.Element, n+1)
		// Determine final range
		last.Exp(fr.NewElement(2), big.NewInt(m))
		//
		ranges[n] = last
	}
	//
	for i := uint(0); i < n; i++ {
		ranges[i] = fr256
	}
	//
	return ranges
}
