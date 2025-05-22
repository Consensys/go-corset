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
package debug

import (
	"fmt"
	"reflect"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// PrintStats is used for printing summary information about a constraint set,
// such as the number and type of constraints, etc.
func PrintStats(hirSchema *hir.Schema, hir bool, mir bool, air bool, optConfig mir.OptimisationConfig) {
	schemas := make([]sc.Schema, 0)
	mirSchema := hirSchema.LowerToMir()
	airSchema := mirSchema.LowerToAir(optConfig)
	// Construct columns
	if hir {
		schemas = append(schemas, hirSchema)
	}

	if mir {
		schemas = append(schemas, mirSchema)
	}

	if air {
		schemas = append(schemas, airSchema)
	}
	//
	n := 1 + uint(len(schemas))
	m := uint(len(schemaSummarisers))
	tbl := termio.NewTablePrinter(n, m)
	// Go!
	for i := uint(0); i < m; i++ {
		ith := schemaSummarisers[i]
		row := make([]termio.FormattedText, n)
		row[0] = termio.NewText(ith.name)

		for j := 0; j < len(schemas); j++ {
			count := ith.summary(schemas[j])
			row[j+1] = termio.NewText(fmt.Sprintf("%d", count))
		}

		tbl.SetRow(i, row...)
	}
	//
	tbl.SetMaxWidths(64)
	tbl.Print(true)
}

// ============================================================================
// Schema Summarisers
// ============================================================================

type schemaSummariser struct {
	name    string
	summary func(sc.Schema) int
}

var schemaSummarisers []schemaSummariser = []schemaSummariser{
	// Constraints
	constraintCounter("Constraints", vanishingConstraints...),
	constraintCounter("Lookups", lookupConstraints...),
	constraintCounter("Permutations", permutationConstraints...),
	constraintCounter("Range", rangeConstraints...),
	// Assignments
	assignmentCounter("Decompositions", reflect.TypeOf((*assignment.ByteDecomposition)(nil))),
	assignmentCounter("Committed Columns", reflect.TypeOf((*assignment.DataColumn)(nil))),
	assignmentCounter("Computed Columns", computedColumns...),
	assignmentCounter("Computation Columns", reflect.TypeOf((*assignment.Computation)(nil))),
	assignmentCounter("Interleavings", reflect.TypeOf((*assignment.Interleaving)(nil))),
	assignmentCounter("Lexicographic Orderings", reflect.TypeOf((*assignment.LexicographicSort)(nil))),
	assignmentCounter("Sorted Permutations", reflect.TypeOf((*assignment.SortedPermutation)(nil))),
	// Columns
	columnCounter(),
	columnWidthSummariser(1, 1),
	columnWidthSummariser(2, 4),
	columnWidthSummariser(5, 8),
	columnWidthSummariser(9, 16),
	columnWidthSummariser(17, 32),
	columnWidthSummariser(33, 64),
	columnWidthSummariser(65, 128),
	columnWidthSummariser(129, 256),
}

var vanishingConstraints = []reflect.Type{
	reflect.TypeOf((hir.VanishingConstraint)(nil)),
	reflect.TypeOf((mir.VanishingConstraint)(nil)),
	reflect.TypeOf((air.VanishingConstraint)(nil))}

var lookupConstraints = []reflect.Type{
	reflect.TypeOf((hir.LookupConstraint)(nil)),
	reflect.TypeOf((mir.LookupConstraint)(nil)),
	reflect.TypeOf((air.LookupConstraint)(nil))}

var rangeConstraints = []reflect.Type{
	reflect.TypeOf((hir.RangeConstraint)(nil)),
	reflect.TypeOf((mir.RangeConstraint)(nil)),
	reflect.TypeOf((air.RangeConstraint)(nil))}

var permutationConstraints = []reflect.Type{
	// permutation constraints only exist at AIR level
	reflect.TypeOf((air.PermutationConstraint)(nil))}

var computedColumns = []reflect.Type{
	// permutation constraints only exist at AIR level
	reflect.TypeOf((*assignment.ComputedColumn)(nil))}

func constraintCounter(title string, types ...reflect.Type) schemaSummariser {
	return schemaSummariser{
		name: title,
		summary: func(schema sc.Schema) int {
			sum := 0
			for _, t := range types {
				sum += typeOfCounter(schema.Constraints(), t)
			}
			return sum
		},
	}
}

func assignmentCounter(title string, types ...reflect.Type) schemaSummariser {
	return schemaSummariser{
		name: title,
		summary: func(schema sc.Schema) int {
			sum := 0
			for _, t := range types {
				sum += typeOfCounter(schema.Declarations(), t)
			}
			return sum
		},
	}
}

func typeOfCounter[T any](iter iter.Iterator[T], dyntype reflect.Type) int {
	count := 0

	for iter.HasNext() {
		ith := iter.Next()
		if dyntype == reflect.TypeOf(ith) {
			count++
		}
	}

	return count
}

func columnCounter() schemaSummariser {
	// return schemaSummariser{
	// 	name: "Columns (all)",
	// 	summary: func(schema sc.Schema) int {
	// 		count := 0
	// 		for i := schema.Columns(); i.HasNext(); {
	// 			i.Next()
	// 			count++
	// 		}
	// 		return count
	// 	},
	// }
	panic("todo")
}

func columnWidthSummariser(lowWidth uint, highWidth uint) schemaSummariser {
	return schemaSummariser{
		name: fmt.Sprintf("Columns (%d..%d bits)", lowWidth, highWidth),
		summary: func(schema sc.Schema) int {
			count := 0
			for i := schema.Modules(); i.HasNext(); {
				m := i.Next()
				for c := uint(0); c < m.Width(); c++ {
					ith := m.Column(c)
					ithWidth := ith.DataType.BitWidth()
					if ithWidth >= lowWidth && ithWidth <= highWidth {
						count++
					}
				}
			}
			return count
		},
	}
}
