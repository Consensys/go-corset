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

func (e jsonComputationSet) addToSchema(columns []column, schema *hir.Schema) {
	//
	for _, c := range e.Computations {
		if c.Sorted != nil {
			targetRefs := asRegisters(c.Sorted.Tos, columns)
			sourceRefs := asRegisters(c.Sorted.Froms, columns)
			// Resolve enclosing module
			module := schema.Columns().Nth(sourceRefs[0]).Context().Module()
			// Sanity check assumptions
			if len(sourceRefs) != len(targetRefs) {
				panic("differing number of source / target columns in sorted permutation")
			}
			// Convert source refs into column indexes
			sources := make([]uint, len(sourceRefs))
			// Convert target refs into columns
			targets := make([]sc.Column, len(targetRefs))
			//
			ctx := trace.VoidContext()
			//
			for i, dst_cid := range targetRefs {
				src_cid := sourceRefs[i]
				// Determine type of source column
				src_col := schema.Columns().Nth(src_cid)
				dst_col := schema.Columns().Nth(dst_cid)
				// Sanity check enclosing modules match
				if src_col.Context().Module() != module {
					panic("inconsistent enclosing module for sorted permutation (source)")
				} else if dst_col.Context().Module() != module {
					panic("inconsistent enclosing module for sorted permutation (target)")
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
				targets[i] = sc.NewColumn(ctx, dst_col.Name(), src_col.Type())
			}
			// Finally, add the sorted permutation assignment
			schema.AddAssignment(assignment.NewSortedPermutation(ctx, targets, c.Sorted.Signs, sources))
		}
	}
}
