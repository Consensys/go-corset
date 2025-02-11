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
	// If we get here, then no conflicts were detected.
	return ctx
}

func requiredColumnsOfTerm(e Term) *set.SortedSet[uint] {
	switch e := e.(type) {
	case *Add:
		return requiredColumnsOfTerms(e.Args)
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

func requiredColumnsOfColumnAccess(e *ColumnAccess) *set.SortedSet[uint] {
	r := set.NewSortedSet[uint]()
	r.Insert(e.Column)
	// Done
	return r
}

func requiredCellsOfTerm(e Term, row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	switch e := e.(type) {
	case *Add:
		return requiredCellsOfTerms(e.Args, row, tr)
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
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func requiredCellsOfTerms(args []Term, row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return set.UnionAnySortedSets(args, func(e Term) *set.AnySortedSet[trace.CellRef] {
		return requiredCellsOfTerm(e, row, tr)
	})
}

func requiredCellsOfColumnAccess(e *ColumnAccess, row int) *set.AnySortedSet[trace.CellRef] {
	set := set.NewAnySortedSet[trace.CellRef]()
	set.Insert(trace.NewCellRef(e.Column, row+e.Shift))
	//
	return set
}
