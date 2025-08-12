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
	"math/big"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
)

// Expand converts an arbitrary expression into a specific column index.  In
// general, this means adding a computed column to hold the value of the
// arbitrary expression and returning its index.  However, this can be optimised
// in the case the given expression is a direct column access by simply
// returning the accessed column index.
func Expand(bitwidth uint, e air.Term, module *air.ModuleBuilder) schema.RegisterId {
	// Check whether this is a straightforward register access.
	if ca, ok := e.(*air.ColumnAccess); ok && ca.Shift == 0 {
		// Optimisation possible
		return ca.Register
	}
	//
	var (
		// Determine computed column name
		name = e.Lisp(true, module).String(false)
		// Look up column
		index, ok = module.HasRegister(name)
		// Default padding (for now)
		padding big.Int = ir.PaddingFor(e, module)
	)
	// Add new column (if it does not already exist)
	if !ok {
		// Add computed column
		index = module.NewRegister(schema.NewComputedRegister(name, bitwidth, padding))
		module.AddAssignment(assignment.NewComputedRegister(sc.NewRegisterRef(module.Id(), index), e, true))
		// Construct v == [e]
		v := ir.NewRegisterAccess[bls12_377.Element, air.Term](index, 0)
		// v - e
		eq_e_v := ir.Subtract(v, e)
		// Ensure (v - e) == 0, where v is value of computed column.
		module.AddConstraint(
			air.NewVanishingConstraint(name, module.Id(), util.None[int](), eq_e_v))
	}
	//
	return index
}
