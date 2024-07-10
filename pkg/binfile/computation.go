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
	Sorted *jsonSortedComputation
}

type jsonSortedComputation struct {
	Froms []string `json:"froms"`
	Tos   []string `json:"tos"`
	Signs []bool   `json:"signs"`
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
			targetIDs := asColumns(c.Sorted.Tos)
			sourceIDs := asColumns(c.Sorted.Froms)
			handle := asHandle(columns[sourceIDs[0]].Handle)
			// Resolve enclosing module
			mid, ok := schema.Modules().Find(func(m sc.Module) bool {
				return m.Name() == handle.module
			})
			// Sanity check assumptions
			if !ok {
				panic(fmt.Sprintf("unknown module %s", handle.module))
			} else if len(sourceIDs) != len(targetIDs) {
				panic("differing number of source / target columns in sorted permutation")
			}
			// Convert source refs into column indexes
			sources := make([]uint, len(sourceIDs))
			// Convert target refs into columns
			targets := make([]sc.Column, len(targetIDs))
			//
			ctx := trace.VoidContext()
			//
			for i, target_id := range targetIDs {
				source_id := sourceIDs[i]
				// Extract binfile info about target column
				dst_col := columns[target_id]
				dst_hnd := asHandle(dst_col.Handle)
				// Determine schema column index for ith source column.
				src_cid, ok := colmap[source_id]
				if !ok {
					h := asHandle(columns[source_id].Handle)
					panic(fmt.Sprintf("unallocated source column %s.%s", h.module, h.column))
				}
				// Extract schema info about source column
				src_col := schema.Columns().Nth(src_cid)
				// Sanity check enclosing modules match
				if src_col.Context().Module() != mid {
					panic("inconsistent enclosing module for sorted permutation (source)")
				}

				ctx = ctx.Join(src_col.Context())
				// Sanity check we have a sensible type here.
				if src_col.Type().AsUint() == nil {
					panic(fmt.Sprintf("source column %s has field type", src_col.Name()))
				} else if ctx.IsConflicted() {
					panic(fmt.Sprintf("source column %s has conflicted evaluation context", src_col.Name()))
				} else if ctx.IsVoid() {
					panic(fmt.Sprintf("source column %s has void evaluation context", src_col.Name()))
				}

				sources[i] = src_cid
				targets[i] = sc.NewColumn(ctx, dst_hnd.column, src_col.Type())
				// Update allocation information.
				colmap[target_id] = index
				index++
			}
			// Finally, add the sorted permutation assignment
			schema.AddAssignment(assignment.NewSortedPermutation(ctx, targets, c.Sorted.Signs, sources))
		}
	}
}
