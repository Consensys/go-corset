package binfile

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
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

func (e jsonComputationSet) addToSchema(schema *hir.Schema) {
	for _, c := range e.Computations {
		if c.Sorted != nil {
			targetRefs := asColumnRefs(c.Sorted.Tos)
			sourceRefs := asColumnRefs(c.Sorted.Froms)
			// Resolve enclosing module
			module, _ := targetRefs[0].resolve(schema)
			// Sanity check assumptions
			if len(sourceRefs) != len(targetRefs) {
				panic("differing number of source / target columns in sorted permutation")
			}
			// Convert source refs into column indexes
			sources := make([]uint, len(sourceRefs))
			// Convert target refs into columns
			targets := make([]sc.Column, len(targetRefs))
			//
			for i, targetRef := range targetRefs {
				src_cid, src_mid := sourceRefs[i].resolve(schema)
				_, dst_mid := targetRef.resolve(schema)
				// Sanity check enclosing modules match
				if src_mid != dst_mid || src_mid != module {
					panic("inconsistent enclosing module for sorted permutation")
				}
				// Determine type of source column
				ith := schema.Columns().Nth(src_cid)
				// Sanity check we have a sensible type here.
				if ith.Type().AsUint() == nil {
					panic(fmt.Sprintf("source column %s has field type", sourceRefs[i]))
				}

				sources[i] = src_cid
				targets[i] = sc.NewColumn(ith.Module(), targetRef.column, ith.Type())
			}
			// Finally, add the permutation column
			schema.AddPermutationColumns(module, targets, c.Sorted.Signs, sources)
		}
	}
}
