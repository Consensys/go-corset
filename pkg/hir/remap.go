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
package hir

import (
	"math"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Remap a given schema using a given criteria to determine which modules to keep.
func Remap(schema *Schema, criteria func(Module) bool) {
	var remapper Remapper
	// Remap modules
	remapper.remapModules(*schema, criteria)
	// Remap columns
	remapper.remapColumns(*schema)
	// Remap assignments
	remapper.remapAssignments(*schema)
	// Remap constraints
	remapper.remapConstraints(*schema)
	// Remap assertions
	remapper.remapAssertions(*schema)
	// Done
	*schema = remapper.schema
}

// Remapper is a tool for remapping modules and columns in a schema after one or
// more modules have been deleted.
type Remapper struct {
	// Mapping from module identifiers before to module identifiers after.
	modmap []uint
	// Mapping from column identifiers before to column identifiers after.
	colmap []uint
	// New schema being constructed
	schema Schema
}

func (p *Remapper) remapModules(oSchema Schema, criteria func(Module) bool) {
	// Initialise modmap
	p.modmap = make([]uint, len(oSchema.modules))
	//
	for i, m := range oSchema.modules {
		var mid uint
		// Decide whether or not to keep the module.
		if criteria(m) {
			mid = p.schema.AddModule(m.Name, m.Condition)
		} else {
			// Mark module as deleted
			mid = math.MaxUint
		}
		//
		p.modmap[i] = mid
	}
}

func (p *Remapper) remapColumns(oSchema Schema) {
	// Initialise colmap
	p.colmap = make([]uint, len(oSchema.inputs))
	//
	for iter, i := oSchema.InputColumns(), 0; iter.HasNext(); i++ {
		var (
			ith = iter.Next()
			cid uint
		)
		// Add column (if applicable)
		if p.modmap[ith.Context.ModuleId] != math.MaxUint {
			ctx := p.remapContext(ith.Context)
			cid = p.schema.AddDataColumn(ctx, ith.Name, ith.DataType)
		} else {
			// Column is part of a module which has been deleted, hence this
			// column is also deleted.
			cid = math.MaxUint
		}
		//
		p.colmap[i] = cid
	}
}

func (p *Remapper) remapAssignments(oSchema Schema) {
	for _, a := range oSchema.assignments {
		p.schema.AddAssignment(p.remapAssignment(a))
	}
}

func (p *Remapper) remapConstraints(oSchema Schema) {
	for _, c := range oSchema.constraints {
		nc := p.remapConstraint(c)
		p.schema.constraints = append(p.schema.constraints, nc)
	}
}

func (p *Remapper) remapAssertions(oSchema Schema) {
	for _, c := range oSchema.assertions {
		nc := p.remapAssertion(c)
		p.schema.constraints = append(p.schema.constraints, nc)
	}
}

func (p *Remapper) remapAssignment(a sc.Assignment) sc.Assignment {
	switch a.(type) {
	case *assignment.Computation:
		panic("todo")
	case *assignment.Interleaving:
		panic("todo")
	case *assignment.SortedPermutation:
		panic("todo")
	default:
		// All other cases are not used at the HIR level, only at lower levels.
		// Hence, they can be ignored.
		panic("unreachable")
	}
}

func (p *Remapper) remapConstraint(c sc.Constraint) sc.Constraint {
	switch c := c.(type) {
	case LookupConstraint:
		panic("todo")
	case RangeConstraint:
		panic("todo")
	case SortedConstraint:
		panic("todo")
	case VanishingConstraint:
		return p.remapVanishing(c)
	default:
		// should be no other cases
		panic("unreachable")
	}
}

func (p *Remapper) remapAssertion(a PropertyAssertion) PropertyAssertion {
	panic("todo")
}

func (p *Remapper) remapVanishing(c VanishingConstraint) sc.Constraint {
	return &constraint.VanishingConstraint[Expr]{
		Handle:     c.Handle,
		Case:       c.Case,
		Context:    p.remapContext(c.Context),
		Domain:     c.Domain,
		Constraint: p.remapExpression(c.Constraint),
	}
}

func (p *Remapper) remapExpression(expr Expr) Expr {
	panic("got here")
}

func (p *Remapper) remapContext(ctx tr.Context) tr.Context {
	mid := p.modmap[ctx.ModuleId]
	// sanity check
	if mid == math.MaxUint {
		panic("remapping deleted context")
	}
	//
	return tr.NewContext(mid, ctx.LengthMultiplier())
}
