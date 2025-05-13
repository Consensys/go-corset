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
	"fmt"
	"math"
	"reflect"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Remap a given schema using a given criteria to determine which modules to keep.
func Remap(schema *Schema, criteria func(Module) bool) {
	var remapper Remapper
	// Remap modules
	remapper.remapModules(*schema, criteria)
	// Remap columns
	remapper.remapInputAndComputedColumns(*schema)
	// Remap constraints
	remapper.remapConstraints(*schema)
	// Remap assertions
	remapper.remapAssertions(*schema)
	// Done
	*schema = remapper.schema
}

// Remapper is a tool for remapping modules and columns in a schema after one or
// more modules have been deleted.  The key challenge is that we have to update
// all module and column ids used within the schema for the new layout.
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

func (p *Remapper) remapInputAndComputedColumns(oSchema Schema) {
	// Initialise colmap
	p.colmap = make([]uint, oSchema.Columns().Count())
	// Remap input columns
	for iter, i := oSchema.InputColumns(), 0; iter.HasNext(); i++ {
		var (
			ith = iter.Next()
			cid uint
		)
		// Add column (if applicable)
		if p.isActive(ith.Context) {
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
	// Remap computed columns
	oldColId := oSchema.InputColumns().Count()
	newColId := p.schema.InputColumns().Count()
	//
	for _, a := range oSchema.assignments {
		if p.isActive(a.Context()) {
			// Remap all active columns
			for i := uint(0); i < a.Columns().Count(); i++ {
				p.colmap[oldColId] = newColId
				oldColId++
				newColId++
			}
			//
			p.schema.AddAssignment(p.remapAssignment(a))
		} else {
			// Delete all inactive columns
			for i := uint(0); i < a.Columns().Count(); i++ {
				p.colmap[oldColId] = math.MaxUint
				oldColId++
			}
		}
	}
}

func (p *Remapper) remapConstraints(oSchema Schema) {
	for _, c := range oSchema.constraints {
		if p.isActive(c.Contexts()...) {
			nc := p.remapConstraint(c)
			p.schema.constraints = append(p.schema.constraints, nc)
		}
	}
}

func (p *Remapper) remapAssertions(oSchema Schema) {
	for _, c := range oSchema.assertions {
		if p.isActive(c.Context) {
			nc := p.remapAssertion(c)
			p.schema.assertions = append(p.schema.assertions, nc)
		}
	}
}

// isActive checks whether the given contexts all reference active modules (i.e.
// which have not been deleted).
func (p *Remapper) isActive(contexts ...tr.Context) bool {
	for _, ctx := range contexts {
		if p.modmap[ctx.ModuleId] == math.MaxUint {
			return false
		}
	}
	//
	return true
}

func (p *Remapper) remapAssignment(a sc.Assignment) sc.Assignment {
	switch a := a.(type) {
	case *assignment.Computation:
		return p.remapComputation(a)
	case *assignment.Interleaving:
		return p.remapInterleaving(a)
	case *assignment.SortedPermutation:
		return p.remapSortedPermutation(a)
	default:
		// All other cases are not used at the HIR level, only at lower levels.
		// Hence, they can be ignored.
		panic("unreachable")
	}
}

func (p *Remapper) remapComputation(c *assignment.Computation) sc.Assignment {
	c.ColumnContext = p.remapContext(c.ColumnContext)
	// Remap target context's to be safe
	p.remapColumns(c.Targets)
	// Remap source columns
	p.remapColumnIds(c.Sources)
	//
	return c
}

func (p *Remapper) remapInterleaving(c *assignment.Interleaving) sc.Assignment {
	p.remapColumn(&c.Target)
	// Remap source columns
	p.remapColumnIds(c.Sources)
	//
	return c
}

func (p *Remapper) remapSortedPermutation(c *assignment.SortedPermutation) sc.Assignment {
	c.ColumnContext = p.remapContext(c.ColumnContext)
	// Remap target context's to be safe
	p.remapColumns(c.Targets)
	// Remap source columns
	p.remapColumnIds(c.Sources)
	//
	return c
}

func (p *Remapper) remapConstraint(c sc.Constraint) sc.Constraint {
	switch c := c.(type) {
	case LookupConstraint:
		return p.remapLookup(c)
	case RangeConstraint:
		return p.remapRange(c)
	case SortedConstraint:
		return p.remapSorted(c)
	case VanishingConstraint:
		return p.remapVanishing(c)
	default:
		// should be no other cases
		panic("unreachable")
	}
}

func (p *Remapper) remapAssertion(c PropertyAssertion) PropertyAssertion {
	c.Context = p.remapContext(c.Context)
	p.remapExpression(c.Property)
	//
	return c
}

func (p *Remapper) remapLookup(c LookupConstraint) sc.Constraint {
	c.SourceContext = p.remapContext(c.SourceContext)
	c.TargetContext = p.remapContext(c.TargetContext)
	// Remap source terms
	for _, e := range c.Sources {
		p.remapExpression(e)
	}
	// Remap target terms
	for _, e := range c.Targets {
		p.remapExpression(e)
	}
	//
	return c
}

func (p *Remapper) remapRange(c RangeConstraint) sc.Constraint {
	// Remap context
	c.Context = p.remapContext(c.Context)
	// Remap expression
	p.remapExpression(c.Expr)
	//
	return c
}

func (p *Remapper) remapSorted(c SortedConstraint) sc.Constraint {
	// Remap context
	c.Context = p.remapContext(c.Context)
	// Remap selector (if applicable)
	if c.Selector.HasValue() {
		p.remapExpression(c.Selector.Unwrap())
	}
	// Remap source terms
	for _, e := range c.Sources {
		p.remapExpression(e)
	}
	//
	return c
}

func (p *Remapper) remapVanishing(c VanishingConstraint) sc.Constraint {
	// Remap context
	c.Context = p.remapContext(c.Context)
	// Remap constraint
	p.remapExpression(c.Constraint)
	//
	return c
}

func (p *Remapper) remapExpression(expr Expr) {
	p.remapTerm(expr.Term)
}

func (p *Remapper) remapTerm(e Term) {
	switch e := e.(type) {
	case *Add:
		p.remapTerms(e.Args...)
	case *Cast:
		p.remapTerm(e.Arg)
	case *Connective:
		p.remapTerms(e.Args...)
	case *Constant:
		// nothing
	case *Equation:
		p.remapTerms(e.Lhs, e.Rhs)
	case *LabelledConstant:
		// nothing
	case *ColumnAccess:
		// Remap column ID
		e.Column = p.remapColumnId(e.Column)
	case *Exp:
		p.remapTerm(e.Arg)
	case *IfZero:
		p.remapTerm(e.Condition)
		p.remapOptionalTerm(e.TrueBranch)
		p.remapOptionalTerm(e.FalseBranch)
	case *List:
		p.remapTerms(e.Args...)
	case *Mul:
		p.remapTerms(e.Args...)
	case *Norm:
		p.remapTerm(e.Arg)
	case *Not:
		p.remapTerm(e.Arg)
	case *Sub:
		p.remapTerms(e.Args...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func (p *Remapper) remapOptionalTerm(e Term) {
	if e != nil {
		p.remapTerm(e)
	}
}

func (p *Remapper) remapTerms(terms ...Term) {
	for _, term := range terms {
		p.remapTerm(term)
	}
}

func (p *Remapper) remapColumns(columns []sc.Column) {
	for i := range columns {
		p.remapColumn(&columns[i])
	}
}

func (p *Remapper) remapColumn(column *sc.Column) {
	column.Context = p.remapContext(column.Context)
}

func (p *Remapper) remapColumnIds(cids []uint) {
	for i := range cids {
		cids[i] = p.remapColumnId(cids[i])
	}
}

func (p *Remapper) remapColumnId(cid uint) uint {
	cid = p.colmap[cid]
	// sanity check
	if cid == math.MaxUint {
		panic("remapping deleted column")
	}

	return cid
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
