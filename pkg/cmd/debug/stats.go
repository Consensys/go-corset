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
	"strings"

	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// PrintStats is used for printing summary information about a constraint set,
// such as the number and type of constraints, etc.
func PrintStats(stack cmd_util.SchemaStack) {
	schemas := stack.Schemas()
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
	summary func(sc.AnySchema) int
}

var schemaSummarisers []schemaSummariser = []schemaSummariser{
	// Constraints
	constraintCounter("Constraints", isVanishingConstraint),
	constraintCounter("Lookups", isLookupConstraint),
	constraintCounter("Permutations", isPermutationConstraint),
	constraintCounter("Range", isRangeConstraint),
	// Assignments
	assignmentCounter("Decompositions", reflect.TypeOf((*assignment.ByteDecomposition)(nil))),
	//assignmentCounter("Committed Columns", reflect.TypeOf((*assignment.DataColumn)(nil))),
	assignmentCounter("Computed Columns", reflect.TypeOf((*assignment.ComputedRegister)(nil))),
	assignmentCounter("Computation Columns", reflect.TypeOf((*assignment.Computation)(nil))),
	//assignmentCounter("Interleavings", reflect.TypeOf((*assignment.Interleaving)(nil))),
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

func isVanishingConstraint(c schema.Constraint) bool {
	return strings.Contains(reflect.TypeOf(c).Name(), "VanishingConstraint")
}

func isLookupConstraint(c schema.Constraint) bool {
	return strings.Contains(reflect.TypeOf(c).Name(), "LookupConstraint")
}

func isPermutationConstraint(c schema.Constraint) bool {
	return strings.Contains(reflect.TypeOf(c).Name(), "PermutationConstraint")
}

func isRangeConstraint(c schema.Constraint) bool {
	return strings.Contains(reflect.TypeOf(c).Name(), "RangeConstraint")
}

func constraintCounter(title string, includes func(schema.Constraint) bool) schemaSummariser {
	return schemaSummariser{
		name: title,
		summary: func(schema sc.AnySchema) int {
			sum := 0
			for iter := schema.Constraints(); iter.HasNext(); {
				if includes(iter.Next()) {
					sum++
				}
			}
			return sum
		},
	}
}

func assignmentCounter(title string, types ...reflect.Type) schemaSummariser {
	// return schemaSummariser{
	// 	name: title,
	// 	summary: func(schema sc.Schema) int {
	// 		sum := 0
	// 		for _, t := range types {
	// 			sum += typeOfCounter(schema.Declarations(), t)
	// 		}
	// 		return sum
	// 	},
	// }
	panic("todo")
}

func columnCounter() schemaSummariser {
	return schemaSummariser{
		name: "Columns (all)",
		summary: func(schema sc.AnySchema) int {
			count := 0
			for m := range schema.Width() {
				count += int(schema.Module(m).Width())
			}
			return count
		},
	}
}

func columnWidthSummariser(lowWidth uint, highWidth uint) schemaSummariser {
	return schemaSummariser{
		name: fmt.Sprintf("Columns (%d..%d bits)", lowWidth, highWidth),
		summary: func(schema sc.AnySchema) int {
			count := 0
			for i := schema.Modules(); i.HasNext(); {
				m := i.Next()
				for c := uint(0); c < m.Width(); c++ {
					ith := m.Register(c)
					ithWidth := ith.Width
					if ithWidth >= lowWidth && ithWidth <= highWidth {
						count++
					}
				}
			}
			return count
		},
	}
}
