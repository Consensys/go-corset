package assignment

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// SortedPermutation declares one or more columns as sorted permutations of
// existing columns.
type SortedPermutation struct {
	// The new (sorted) columns
	targets []schema.Column
	// The sorting criteria
	Signs []bool
	// The existing columns
	Sources []string
}

// NewSortedPermutation creates a new sorted permutation
func NewSortedPermutation(targets []schema.Column, signs []bool, sources []string) *SortedPermutation {
	if len(targets) != len(signs) || len(signs) != len(sources) {
		panic("target and source column widths must match")
	}

	return &SortedPermutation{targets, signs, sources}
}

// Targets returns the columns declared by this sorted permutation (in the order
// of declaration).  This is the same as Columns(), except that it avoids using
// an iterator.
func (p *SortedPermutation) Targets() []schema.Column {
	return p.targets
}

// String returns a string representation of this constraint.  This is primarily
// used for debugging.
func (p *SortedPermutation) String() string {
	targets := ""
	sources := ""

	index := 0
	for i := 0; i != len(p.targets); i++ {
		if index != 0 {
			targets += " "
		}

		targets += p.targets[i].Name()
		index++
	}

	for i, s := range p.Sources {
		if i != 0 {
			sources += " "
		}

		if p.Signs[i] {
			sources += fmt.Sprintf("+%s", s)
		} else {
			sources += fmt.Sprintf("-%s", s)
		}
	}

	return fmt.Sprintf("(permute (%s) (%s))", targets, sources)
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the columns declared by this sorted permutation (in the order
// of declaration).
func (p *SortedPermutation) Columns() util.Iterator[schema.Column] {
	return util.NewArrayIterator(p.targets)
}

// IsComputed Determines whether or not this declaration is computed.
func (p *SortedPermutation) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *SortedPermutation) RequiredSpillage() uint {
	return uint(0)
}

// ExpandTrace expands a given trace to include the columns specified by a given
// SortedPermutation.  This requires copying the data in the source columns, and
// sorting that data according to the permutation criteria.
func (p *SortedPermutation) ExpandTrace(tr tr.Trace) error {
	// Ensure target columns don't exist
	for i := p.Columns(); i.HasNext(); {
		if tr.HasColumn(i.Next().Name()) {
			panic("target column already exists")
		}
	}

	cols := make([][]*fr.Element, len(p.Sources))
	// Construct target columns
	for i := 0; i < len(p.Sources); i++ {
		src := p.Sources[i]
		// Read column data to initialise permutation.
		data := tr.ColumnByName(src).Data()
		// Copy column data to initialise permutation.
		cols[i] = make([]*fr.Element, len(data))
		copy(cols[i], data)
	}
	// Sort target columns
	util.PermutationSort(cols, p.Signs)
	// Physically add the columns
	index := 0

	for i := p.Columns(); i.HasNext(); {
		dstColName := i.Next().Name()
		srcCol := tr.ColumnByName(p.Sources[index])
		tr.AddColumn(dstColName, cols[index], srcCol.Padding())

		index++
	}
	//
	return nil
}
