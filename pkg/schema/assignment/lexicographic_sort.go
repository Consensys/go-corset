package assignment

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// LexicographicSort provides the necessary computation for filling out columns
// added to enforce lexicographic sorting constraints between one or more source
// columns.  Specifically, a delta column is required along with one selector
// column (binary) for each source column.
type LexicographicSort struct {
	// Context in which source and target columns to be located.  All target and
	// source columns should be contained within this.
	context trace.Context
	// The target columns to be filled.  The first entry is for the delta
	// column, and the remaining n entries are for the selector columns.
	targets []schema.Column
	// Source columns being sorted
	sources  []uint
	signs    []bool
	bitwidth uint
}

// NewLexicographicSort constructs a new LexicographicSorting assignment.
func NewLexicographicSort(prefix string, context trace.Context,
	sources []uint, signs []bool, bitwidth uint) *LexicographicSort {
	//
	targets := make([]schema.Column, len(sources)+1)
	// Create delta column
	targets[0] = schema.NewColumn(context, fmt.Sprintf("%s:delta", prefix), schema.NewUintType(bitwidth))
	// Create selector columns
	for i := range sources {
		ithName := fmt.Sprintf("%s:%d", prefix, i)
		targets[1+i] = schema.NewColumn(context, ithName, schema.NewUintType(1))
	}

	return &LexicographicSort{context, targets, sources, signs, bitwidth}
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the columns declared by this assignment.
func (p *LexicographicSort) Columns() util.Iterator[schema.Column] {
	return util.NewArrayIterator(p.targets)
}

// IsComputed Determines whether or not this declaration is computed (which it
// is).
func (p *LexicographicSort) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *LexicographicSort) RequiredSpillage() uint {
	return uint(0)
}

// ExpandTrace adds columns as needed to support the LexicographicSortingGadget.
// That includes the delta column, and the bit selectors.
func (p *LexicographicSort) ExpandTrace(tr trace.Trace) error {
	columns := tr.Columns()
	zero := fr.NewElement(0)
	one := fr.NewElement(1)
	// Exact number of columns involved in the sort
	ncols := len(p.sources)
	//
	multiplier := p.context.LengthMultiplier()
	// Determine how many rows to be constrained.
	nrows := tr.Modules().Get(p.context.Module()).Height() * multiplier
	// Initialise new data columns
	bit := make([]util.FrArray, ncols)
	// Byte width records the largest width of any column.
	bit_width := uint(0)

	for i := 0; i < ncols; i++ {
		// TODO: following can be optimised to use a single bit per element,
		// rather than an entire byte.
		bit[i] = util.NewFrArray(nrows, 1)
		ith := columns.Get(p.sources[i])
		bit_width = max(bit_width, ith.Data().BitWidth())
	}

	delta := util.NewFrArray(nrows, bit_width)

	for i := uint(0); i < nrows; i++ {
		set := false
		// Initialise delta to zero
		delta.Set(i, &zero)
		// Decide which row is the winner (if any)
		for j := 0; j < ncols; j++ {
			prev := columns.Get(p.sources[j]).Get(int(i - 1))
			curr := columns.Get(p.sources[j]).Get(int(i))

			if !set && prev != nil && prev.Cmp(curr) != 0 {
				var diff fr.Element

				bit[j].Set(i, &one)
				// Compute curr - prev
				if p.signs[j] {
					diff.Set(curr)
					delta.Set(i, diff.Sub(&diff, prev))
				} else {
					diff.Set(prev)
					delta.Set(i, diff.Sub(&diff, curr))
				}

				set = true
			} else {
				bit[j].Set(i, &zero)
			}
		}
	}
	// Add delta column data
	first := p.targets[0]
	columns.Add(first.Context(), first.Name(), delta, &zero)
	// Add bit column data
	for i := 0; i < ncols; i++ {
		ith := p.targets[1+i]
		columns.Add(ith.Context(), ith.Name(), bit[i], &zero)
	}
	// Done.
	return nil
}

// String returns a string representation of this constraint.  This is primarily
// used for debugging.
func (p *LexicographicSort) String() string {
	return fmt.Sprintf("(lexer (%v) (%v) :%d))", any(p.targets), p.signs, p.bitwidth)
}
