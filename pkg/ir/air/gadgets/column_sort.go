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

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ColumnSortGadget adds sorting constraints for a column where the
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
type ColumnSortGadget struct {
	// Prefix is used to construct the delta column name.
	prefix string
	// Identifies column being sorted
	column uint
	// Sign of sort (true = ascending, false = descending)
	sign bool
	// Bitwidth of delta column
	bitwidth uint
	// Strict implies equal values are not permitted.
	strict bool
	// Constraint active when selector is non-zero.
	selector air.Term
}

// NewColumnSortGadget constructs a new column sort gadget which can then be
// configured.
func NewColumnSortGadget(prefix string, column uint, bitwidth uint) ColumnSortGadget {
	return ColumnSortGadget{
		prefix,
		column,
		true,
		bitwidth,
		false,
		ir.Const64[air.Term](1),
	}
}

// SetSign configures the sort direction
func (p *ColumnSortGadget) SetSign(sign bool) {
	p.sign = sign
}

// SetStrict configures strictness
func (p *ColumnSortGadget) SetStrict(strict bool) {
	p.strict = strict
}

// SetSelector sets the selector for this constraint.
func (p *ColumnSortGadget) SetSelector(selector air.Term) {
	p.selector = selector
}

// Apply a given ColumnSortGadget to a given schema.
func (p *ColumnSortGadget) Apply(module *air.ModuleBuilder) {
	var (
		deltaName string
		// Construct enclosing context
		context = trace.NewContext(module.Id(), 1)
	)
	// Configure computation
	Xk := ir.NewRegisterAccess[air.Term](p.column, 0)
	Xkm1 := ir.NewRegisterAccess[air.Term](p.column, -1)
	// Account for sign
	var Xdiff air.Term
	if p.sign {
		Xdiff = ir.Subtract(Xk, Xkm1)
		deltaName = fmt.Sprintf("+%s", p.prefix)
	} else {
		Xdiff = ir.Subtract(Xkm1, Xk)
		deltaName = fmt.Sprintf("-%s", p.prefix)
	}
	// Apply strictness
	if p.strict {
		Xdiff = ir.Subtract(Xdiff, ir.Const64[air.Term](1))
	}
	// Look up register
	deltaIndex, ok := module.HasRegister(deltaName)
	// Add new column (if it does not already exist)
	if !ok {
		//
		Xcomp := ir.Product(p.selector, Xdiff)
		// Add computed column
		deltaIndex = module.NewRegister(sc.NewComputedRegister(deltaName, p.bitwidth))
		// Add corresponding assignment
		module.AddAssignment(
			assignment.NewComputedRegister(context, deltaIndex, Xcomp))
		// Add necessary bitwidth constraints
		ApplyBitwidthGadget(deltaIndex, p.bitwidth, p.selector, module)
	}
	// Configure constraint: Delta[k] = X[k] - X[k-1]
	Dk := ir.NewRegisterAccess[air.Term](deltaIndex, 0)
	// Apply selector
	e := ir.Product(p.selector, ir.Subtract(Dk, Xdiff))
	// Done
	module.AddConstraint(
		air.NewVanishingConstraint(deltaName, context, util.None[int](), e),
	)
}
