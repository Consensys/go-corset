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

	"github.com/consensys/go-corset/pkg/air"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/util"
)

// ApplyColumnSortGadget adds sorting constraints for a column where the
// difference between any two rows (i.e. the delta) is constrained to fit within
// a given bitwidth.  The target column is assumed to have an appropriate
// (enforced) bitwidth to ensure overflow cannot arise.  The sorting constraint
// is either ascending (positively signed) or descending (negatively signed).  A
// delta column is added along with bitwidth constraints (where necessary) to
// ensure the delta is within the given width.
//
// This gadget does not attempt to sort the column data during trace expansion,
// and assumes the data either comes sorted or is sorted by some other
// computation.
func ApplyColumnSortGadget(col uint, sign bool, bitwidth uint, schema *air.Schema) {
	var deltaName string
	// Identify target column
	column := schema.Columns().Nth(col)
	// Determine column name
	name := column.Name
	// Configure computation
	Xk := air.NewColumnAccess(col, 0)
	Xkm1 := air.NewColumnAccess(col, -1)
	// Account for sign
	var Xdiff air.Expr
	if sign {
		Xdiff = Xk.Sub(Xkm1)
		deltaName = fmt.Sprintf("+%s", name)
	} else {
		Xdiff = Xkm1.Sub(Xk)
		deltaName = fmt.Sprintf("-%s", name)
	}
	// Look up column
	deltaIndex, ok := sc.ColumnIndexOf(schema, column.Context.Module(), deltaName)
	// Add new column (if it does not already exist)
	if !ok {
		// NOTE: require delta bitwidth is greater than 16 because, otherwise,
		// failing traces can result in a panic when the (miscomputed) delta
		// overflows the underlying column.  This works around the problem by
		// ensuring the underlying column is an instanceof FrIndexColumn (since
		// this can hold values of any width).
		deltaBitwidth := max(17, bitwidth)
		//
		deltaIndex = schema.AddAssignment(
			assignment.NewComputedColumn(column.Context, deltaName, sc.NewUintType(deltaBitwidth), Xdiff))
	}
	// Add necessary bitwidth constraints
	ApplyBitwidthGadget(deltaIndex, bitwidth, schema)
	// Configure constraint: Delta[k] = X[k] - X[k-1]
	Dk := air.NewColumnAccess(deltaIndex, 0)
	schema.AddVanishingConstraint(deltaName, 0, column.Context, util.None[int](), Dk.Equate(Xdiff))
}
