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
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/trace"
)

// Expand converts an arbitrary expression into a specific column index.  In
// general, this means adding a computed column to hold the value of the
// arbitrary expression and returning its index.  However, this can be optimised
// in the case the given expression is a direct column access by simply
// returning the accessed column index.
func Expand(ctx trace.Context, bitwidth uint, e air.Term, module *air.ModuleBuilder) uint {
	if ctx.IsVoid() || ctx.IsConflicted() {
		panic("conflicting (or void) context")
	}
	// Check whether this is a straightforward register access.
	if ca, ok := e.(*air.ColumnAccess); ok && ca.Shift == 0 {
		// Optimisation possible
		return ca.Register
	}
	// Determine computed column name
	name := e.Lisp(module).String(false)
	// Look up column
	index, ok := module.HasRegister(name)
	// Add new column (if it does not already exist)
	if !ok {
		// Add computed column
		// index = schema.AddAssignment(assignment.NewComputedColumn(ctx, name, sc.NewUintType(bitwidth), e))
		// // Construct v == [e]
		// v := air.NewColumnAccess(index, 0)
		// // Construct 1 == e/e
		// eq_e_v := v.Equate(e)
		// // Ensure (e - v) == 0, where v is value of computed column.
		// schema.AddVanishingConstraint(name, 0, ctx, util.None[int](), eq_e_v)
		panic("todo")
	}
	//
	return index
}
