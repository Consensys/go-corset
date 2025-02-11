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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/util"
)

// ApplyBinaryGadget adds a binarity constraint for a given column in the schema
// which enforces that all values in the given column are either 0 or 1. For a
// column X, this corresponds to the vanishing constraint X * (X-1) == 0.
func ApplyBinaryGadget(col uint, schema *air.Schema) {
	// Identify target column
	column := schema.Columns().Nth(col)
	// Determine column name
	name := column.Name
	// Construct X
	X := air.NewColumnAccess(col, 0)
	// Construct X-1
	X_m1 := X.Sub(air.NewConst64(1))
	// Construct X * (X-1)
	X_X_m1 := X.Mul(X_m1)
	// Done!
	schema.AddVanishingConstraint(fmt.Sprintf("%s:u1", name), column.Context, util.None[int](), X_X_m1)
}

// ApplyBitwidthGadget ensures all values in a given column fit within a given
// number of bits.  This is implemented using a *byte decomposition* which adds
// n columns and a vanishing constraint (where n*8 >= nbits).
func ApplyBitwidthGadget(col uint, nbits uint, schema *air.Schema) {
	if nbits%8 != 0 {
		panic("asymmetric bitwidth constraints not yet supported")
	} else if nbits == 0 {
		panic("zero bitwidth constraint encountered")
	}
	// Identify target column
	column := schema.Columns().Nth(col)
	// Calculate how many bytes required.
	n := nbits / 8
	es := make([]air.Expr, n)
	fr256 := fr.NewElement(256)
	name := column.Name
	coefficient := fr.NewElement(1)
	// Add decomposition assignment
	index := schema.AddAssignment(
		assignment.NewByteDecomposition(name, column.Context, col, n))
	// Construct Columns
	for i := uint(0); i < n; i++ {
		// Create Column + Constraint
		es[i] = air.NewColumnAccess(index+i, 0).Mul(air.NewConst(coefficient))

		schema.AddRangeConstraint(index+i, fr256)
		// Update coefficient
		coefficient.Mul(&coefficient, &fr256)
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	sum := air.Sum(es...)
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := air.NewColumnAccess(col, 0)
	eq := X.Equate(sum)
	// Construct column name
	schema.AddVanishingConstraint(fmt.Sprintf("%s:u%d", name, nbits), column.Context, util.None[int](), eq)
}
