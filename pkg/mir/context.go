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
package mir

import (
	"fmt"
	"reflect"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/set"
)

func contextOfTerm(e Term, schema sc.Schema) trace.Context {
	switch e := e.(type) {
	case *Add:
		return contextOfTerms(e.Args, schema)
	case *Cast:
		return contextOfTerm(e.Arg, schema)
	case *Constant:
		return trace.VoidContext[uint]()
	case *ColumnAccess:
		col := schema.Columns().Nth(e.Column)
		return col.Context
	case *Exp:
		return contextOfTerm(e.Arg, schema)
	case *Mul:
		return contextOfTerms(e.Args, schema)
	case *Norm:
		return contextOfTerm(e.Arg, schema)
	case *Sub:
		return contextOfTerms(e.Args, schema)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func contextOfTerms(args []Term, schema sc.Schema) trace.Context {
	ctx := trace.VoidContext[uint]()
	//
	for _, e := range args {
		ctx = ctx.Join(contextOfTerm(e, schema))
	}
	//
	return ctx
}

func contextOfConjunction(conjunction Constraint, schema sc.Schema) trace.Context {
	ctx := trace.VoidContext[uint]()
	//
	for _, e := range conjunction.disjuncts {
		ctx = ctx.Join(contextOfDisjunction(e, schema))
	}
	//
	return ctx
}

func contextOfDisjunction(disjunction Disjunction, schema sc.Schema) trace.Context {
	ctx := trace.VoidContext[uint]()
	//
	for _, e := range disjunction.atoms {
		ctx = ctx.Join(contextOfTerm(e.lhs, schema))
		ctx = ctx.Join(contextOfTerm(e.rhs, schema))
	}
	//
	return ctx
}

func requiredColumnsOfTerm(e Term) *set.SortedSet[uint] {
	switch e := e.(type) {
	case *Add:
		return requiredColumnsOfTerms(e.Args)
	case *Cast:
		return requiredColumnsOfTerm(e.Arg)
	case *Constant:
		return set.NewSortedSet[uint]()
	case *ColumnAccess:
		return requiredColumnsOfColumnAccess(e)
	case *Exp:
		return requiredColumnsOfTerm(e.Arg)
	case *Mul:
		return requiredColumnsOfTerms(e.Args)
	case *Norm:
		return requiredColumnsOfTerm(e.Arg)
	case *Sub:
		return requiredColumnsOfTerms(e.Args)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func requiredColumnsOfTerms(args []Term) *set.SortedSet[uint] {
	return set.UnionSortedSets(args, func(e Term) *set.SortedSet[uint] {
		return requiredColumnsOfTerm(e)
	})
}

func requiredColumnsOfConjunction(conjunction Constraint) *set.SortedSet[uint] {
	return set.UnionSortedSets(conjunction.disjuncts, func(d Disjunction) *set.SortedSet[uint] {
		return requiredColumnsOfDisjunction(d)
	})
}

func requiredColumnsOfDisjunction(disjunction Disjunction) *set.SortedSet[uint] {
	return set.UnionSortedSets(disjunction.atoms, func(e Equation) *set.SortedSet[uint] {
		cols := requiredColumnsOfTerm(e.lhs)
		cols.InsertSorted(requiredColumnsOfTerm(e.rhs))

		return cols
	})
}

func requiredColumnsOfColumnAccess(e *ColumnAccess) *set.SortedSet[uint] {
	r := set.NewSortedSet[uint]()
	r.Insert(e.Column)
	// Done
	return r
}

func requiredCellsOfTerm(t Term, row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	switch e := t.(type) {
	case *Add:
		return requiredCellsOfTerms(e.Args, row, tr)
	case *Cast:
		return requiredCellsOfTerm(e.Arg, row, tr)
	case *Constant:
		return set.NewAnySortedSet[trace.CellRef]()
	case *ColumnAccess:
		return requiredCellsOfColumnAccess(e, row)
	case *Exp:
		return requiredCellsOfTerm(e.Arg, row, tr)
	case *Mul:
		return requiredCellsOfTerms(e.Args, row, tr)
	case *Norm:
		return requiredCellsOfTerm(e.Arg, row, tr)
	case *Sub:
		return requiredCellsOfTerms(e.Args, row, tr)
	default:
		name := reflect.TypeOf(t).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func requiredCellsOfTerms(args []Term, row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return set.UnionAnySortedSets(args, func(e Term) *set.AnySortedSet[trace.CellRef] {
		return requiredCellsOfTerm(e, row, tr)
	})
}

func requiredCellsOfConjunction(conjunction Constraint, row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return set.UnionAnySortedSets(conjunction.disjuncts, func(d Disjunction) *set.AnySortedSet[trace.CellRef] {
		return requiredCellsOfDisjunction(d, row, tr)
	})
}

func requiredCellsOfDisjunction(disjunction Disjunction, row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return set.UnionAnySortedSets(disjunction.atoms, func(e Equation) *set.AnySortedSet[trace.CellRef] {
		cells := requiredCellsOfTerm(e.lhs, row, tr)
		cells.InsertSorted(requiredCellsOfTerm(e.rhs, row, tr))

		return cells
	})
}

func requiredCellsOfColumnAccess(e *ColumnAccess, row int) *set.AnySortedSet[trace.CellRef] {
	set := set.NewAnySortedSet[trace.CellRef]()
	set.Insert(trace.NewCellRef(e.Column, row+e.Shift))
	//
	return set
}
