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
			// Sanity check assumptions
			if len(sourceRefs) != len(targetRefs) {
				panic("differing number of source / target columns in sorted permutation")
			}
			// Convert source refs into column indexes
			sources := make([]uint, len(sourceRefs))
			// Convert target refs into columns
			targets := make([]sc.Column, len(targetRefs))
			//
			for i, r := range targetRefs {
				cid, ok := sc.ColumnIndexOf(schema, sourceRefs[i])
				// Sanity check source column exists
				if !ok {
					panic(fmt.Sprintf("unknown column %s", sourceRefs[i]))
				}
				// Determine type of source column
				ith := schema.Columns().Nth(cid).Type()
				// Sanity check we have a sensible type here.
				if ith.AsUint() == nil {
					panic(fmt.Sprintf("source column %s has field type", sourceRefs[i]))
				}

				sources[i] = cid
				targets[i] = sc.NewColumn(r, ith)
			}
			// Finally, add the permutation column
			schema.AddPermutationColumns(targets, c.Sorted.Signs, sources)
		}
	}
}
