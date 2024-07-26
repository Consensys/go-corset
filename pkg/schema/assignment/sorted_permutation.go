package assignment

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// SortedPermutation declares one or more columns as sorted permutations of
// existing columns.
type SortedPermutation struct {
	// Context where this data column is located.
	context trace.Context
	// The new (sorted) columns
	targets []schema.Column
	// The sorting criteria
	signs []bool
	// The existing columns
	sources []uint
}

// NewSortedPermutation creates a new sorted permutation
func NewSortedPermutation(context tr.Context, targets []schema.Column,
	signs []bool, sources []uint) *SortedPermutation {
	if len(targets) != len(signs) || len(signs) != len(sources) {
		panic("target and source column widths must match")
	}
	// Check modules
	for _, c := range targets {
		if c.Context() != context {
			panic("inconsistent evaluation contexts")
		}
	}

	return &SortedPermutation{context, targets, signs, sources}
}

// Module returns the module which encloses this sorted permutation.
func (p *SortedPermutation) Module() uint {
	return p.context.Module()
}

// Sources returns the columns used by this sorted permutation to define the new
// (sorted) columns.
func (p *SortedPermutation) Sources() []uint {
	return p.sources
}

// Signs returns the sorting direction for each column defined by this sorted permutation.
func (p *SortedPermutation) Signs() []bool {
	return p.signs
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

	for i, s := range p.sources {
		if i != 0 {
			sources += " "
		}

		if p.signs[i] {
			sources += fmt.Sprintf("+#%d", s)
		} else {
			sources += fmt.Sprintf("-#%d", s)
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
	columns := tr.Columns()
	// Ensure target columns don't exist
	for i := p.Columns(); i.HasNext(); {
		name := i.Next().Name()
		// Sanity check no column already exists with this name.
		if _, ok := columns.IndexOf(p.context.Module(), name); ok {
			return fmt.Errorf("permutation column already exists ({%s})", name)
		}
	}

	cols := make([]util.FrArray, len(p.sources))
	// Construct target columns
	for i := 0; i < len(p.sources); i++ {
		src := p.sources[i]
		// Read column data
		data := columns.Get(src).Data()
		// Clone it to initialise permutation.
		cols[i] = data.Clone()
	}
	// Sort target columns
	util.PermutationSort(cols, p.signs)
	// Physically add the columns
	index := 0

	for i := p.Columns(); i.HasNext(); index++ {
		ith := i.Next()
		dstColName := ith.Name()
		srcCol := tr.Columns().Get(p.sources[index])
		columns.Add(ith.Context(), dstColName, cols[index], srcCol.Padding())
	}
	//
	return nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *SortedPermutation) Dependencies() []uint {
	return p.sources
}
