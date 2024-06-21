package binfile

import (
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
			refs := asColumnRefs(c.Sorted.Tos)
			sources := asColumnRefs(c.Sorted.Froms)
			// Convert target refs into columns
			targets := make([]sc.Column, len(refs))

			for i, r := range refs {
				// TODO: correctly determine type
				ith := &sc.FieldType{}
				targets[i] = sc.NewColumn(r, ith)
			}
			// Finally, add the permutation column
			schema.AddPermutationColumns(targets, c.Sorted.Signs, sources)
		}
	}
}
