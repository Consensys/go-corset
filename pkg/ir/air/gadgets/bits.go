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
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/schema"
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
// bitwidth.  This is implemented using a *byte decomposition* which adds n
// columns and a vanishing constraint (where n*8 >= bitwidth).
func ApplyBitwidthGadget(col uint, bitwidth uint, selector air.Term, module *air.ModuleBuilder) {
	context := trace.NewContext(module.Id(), 1)
	// Identify target register name
	name := module.Register(col).Name
	// Allocated computed byte registers in the given module, and add required
	// range constraints.
	byteRegisters := allocateByteRegisters(name, bitwidth, module)
	// Build up the decomposition sum
	sum := buildDecompositionTerm(bitwidth, byteRegisters)
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := ir.NewRegisterAccess[air.Term](col, 0)
	//
	eq := ir.Product(selector, ir.Subtract(X, sum))
	// Construct column name
	module.AddConstraint(
		air.NewVanishingConstraint(fmt.Sprintf("%s:u%d", name, bitwidth), context, util.None[int](), eq))
	// Add decomposition assignment
	module.AddAssignment(
		assignment.NewByteDecomposition(name, context, col, bitwidth, byteRegisters))
}

// Allocate n byte registers, each of which requires a suitable range
// constraint.
func allocateByteRegisters(prefix string, bitwidth uint, module *air.ModuleBuilder) []uint {
	var (
		context = trace.NewContext(module.Id(), 1)
		n       = bitwidth / 8
	)
	//
	if bitwidth == 0 {
		panic("zero byte decomposition encountered")
	}
	// Account for asymetric case
	if bitwidth%8 != 0 {
		n++
	}
	// Allocate target register ids
	targets := make([]uint, n)
	// Allocate byte registers
	for i := uint(0); i < n; i++ {
		name := fmt.Sprintf("%s:%d", prefix, i)
		byteRegister := schema.NewComputedRegister(name, min(8, bitwidth))
		// Allocate byte register
		targets[i] = module.NewRegister(byteRegister)
		// Add suitable range constraint
		ith_access := ir.RawRegisterAccess[air.Term](targets[i], 0)
		//
		module.AddConstraint(
			air.NewRangeConstraint(name, context, *ith_access, byteRegister.Width))
		//
		bitwidth -= 8
	}
	//
	return targets
}

func buildDecompositionTerm(bitwidth uint, byteRegisters []uint) air.Term {
	var (
		// Determine ranges required for the give bitwidth
		ranges = splitColumnRanges(bitwidth)
		// Initialise array of terms
		terms = make([]air.Term, len(byteRegisters))
		// Initialise coefficient
		coefficient = fr.NewElement(1)
	)

	// Construct Columns
	for i, rid := range byteRegisters {
		// Create Column + Constraint
		reg := ir.NewRegisterAccess[air.Term](rid, 0)
		terms[i] = ir.Product(reg, ir.Const[air.Term](coefficient))
		// Update coefficient
		coefficient.Mul(&coefficient, &ranges[i])
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	return ir.Sum(terms...)
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
