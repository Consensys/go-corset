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
package binfile

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/trace"
)

type jsonComputationSet struct {
	Computations []jsonComputation `json:"computations"`
}

type jsonComputation struct {
	Sorted      *jsonSortedComputation
	Interleaved *jsonInterleavedComputation
}

type jsonSortedComputation struct {
	Froms []string `json:"froms"`
	Tos   []string `json:"tos"`
	Signs []bool   `json:"signs"`
}

type jsonInterleavedComputation struct {
	Froms  []string `json:"froms"`
	Target string   `json:"target"`
}

// =============================================================================
// Translation
// =============================================================================

// Iterate through the sorted permutations, allocating them as assignments.
// This is slightly tricky because we must also update the colmap allocation.  I
// believe this could fail the presence of sorted permutations of sorted
// permutations.  In such case, it can be resolved using a more complex
// allocation algorithm which considers the source dependencies.
func (e jsonComputationSet) addToSchema(columns []column, colmap map[uint]uint, schema *hir.Schema) {
	// Determine first allocation index
	index := schema.Columns().Count()
	//
	for _, c := range e.Computations {
		if c.Sorted != nil {
			index = addSortedComputation(c.Sorted, index, columns, colmap, schema)
		} else if c.Interleaved != nil {
			index = addInterleavedComputation(c.Interleaved, index, columns, colmap, schema)
		} else {
			panic("unknown computation encountered")
		}
	}
}

func addSortedComputation(sorted *jsonSortedComputation, index uint,
	columns []column, colmap map[uint]uint, schema *hir.Schema) uint {
	targetIDs := asColumns(sorted.Tos)
	// Convert source refs into column indexes
	ctx, sources := sourceColumnsFromHandles(sorted.Froms, columns, colmap, schema)
	// Sanity checks
	if len(sources) != len(targetIDs) {
		panic("differing number of source / target columns in sorted permutation")
	}
	// Convert target refs into columns
	targets := make([]sc.Column, len(targetIDs))
	//
	for i, target_id := range targetIDs {
		// Extract binfile info about target column
		dst_col := columns[target_id]
		dst_hnd := asHandle(dst_col.Handle)
		src_col := schema.Columns().Nth(sources[i])
		// Sanity check source column type
		if src_col.DataType.AsUint() == nil {
			panic(fmt.Sprintf("source column %s has field type", src_col.Name))
		}

		targets[i] = sc.NewColumn(ctx, dst_hnd.column, src_col.DataType)
		// Update allocation information.
		colmap[target_id] = index
		index++
	}
	// Finally, add the sorted permutation assignment
	schema.AddAssignment(assignment.NewSortedPermutation(ctx, targets, sorted.Signs, sources))
	//
	return index
}

func addInterleavedComputation(c *jsonInterleavedComputation, index uint,
	columns []column, colmap map[uint]uint, schema *hir.Schema) uint {
	// Convert column handles into column indices in the schema
	ctx, sources := sourceColumnsFromHandles(c.Froms, columns, colmap, schema)
	// Determine column handle
	target_id := asColumn(c.Target)
	dst_col := columns[target_id]
	dst_hnd := asHandle(dst_col.Handle)
	// Initially assume bottom type
	var dst_type sc.Type = sc.NewUintType(0)
	// Ensure each column's types included
	for i := range sources {
		src_col := schema.Columns().Nth(sources[i])
		// Update the column type
		dst_type = sc.Join(dst_type, src_col.DataType)
	}
	// Update multiplier
	ctx = ctx.Multiply(uint(len(sources)))
	// Finally, add the sorted permutation assignment
	schema.AddAssignment(assignment.NewInterleaving(ctx, dst_hnd.column, sources, dst_type))
	// Update allocation information.
	colmap[target_id] = index
	//
	return index + 1
}

func sourceColumnsFromHandles(handles []string, columns []column,
	colmap map[uint]uint, schema *hir.Schema) (trace.Context, []uint) {
	sourceIDs := asColumns(handles)
	handle := asHandle(columns[sourceIDs[0]].Handle)
	// Resolve enclosing module
	mid, ok := schema.Modules().Find(func(m sc.Module) bool {
		return m.Name == handle.module
	})
	// Sanity check assumptions
	if !ok {
		panic(fmt.Sprintf("unknown module %s", handle.module))
	}
	// Convert source refs into column indexes
	sources := make([]uint, len(sourceIDs))
	//
	ctx := trace.VoidContext[uint]()
	//
	for i, source_id := range sourceIDs {
		// Determine schema column index for ith source column.
		src_cid, ok := colmap[source_id]
		if !ok {
			h := asHandle(columns[source_id].Handle)
			panic(fmt.Sprintf("unallocated source column %s.%s", h.module, h.column))
		}
		// Extract schema info about source column
		src_col := schema.Columns().Nth(src_cid)
		// Sanity check enclosing modules match
		if src_col.Context.Module() != mid {
			panic("inconsistent enclosing module for sorted permutation (source)")
		}

		ctx = ctx.Join(src_col.Context)
		// Sanity check we have a sensible type here.
		if ctx.IsConflicted() {
			panic(fmt.Sprintf("source column %s has conflicted evaluation context", src_col.Name))
		} else if ctx.IsVoid() {
			panic(fmt.Sprintf("source column %s has void evaluation context", src_col.Name))
		}

		sources[i] = src_cid
	}
	//
	return ctx, sources
}
