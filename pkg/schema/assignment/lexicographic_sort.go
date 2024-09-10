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

// Context returns the evaluation context for this declaration.
func (p *LexicographicSort) Context() trace.Context {
	return p.context
}

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

// ComputeColumns computes the values of columns defined as needed to support
// the LexicographicSortingGadget. That includes the delta column, and the bit
// selectors.
func (p *LexicographicSort) ComputeColumns(tr trace.Trace) ([]trace.ArrayColumn, error) {
	zero := fr.NewElement(0)
	one := fr.NewElement(1)
	first := p.targets[0]
	// Exact number of columns involved in the sort
	nbits := len(p.sources)
	// Determine how many rows to be constrained.
	nrows := tr.Height(p.context)
	// Initialise new data columns
	cols := make([]trace.ArrayColumn, nbits+1)
	// Byte width records the largest width of any column.
	bit_width := uint(0)
	//
	delta := util.NewFrArray(nrows, bit_width)
	cols[0] = trace.NewArrayColumn(first.Context(), first.Name(), delta, zero)
	//
	for i := 0; i < nbits; i++ {
		target := p.targets[1+i]
		source := tr.Column(p.sources[i])
		data := util.NewFrArray(nrows, 1)
		cols[i+1] = trace.NewArrayColumn(target.Context(), target.Name(), data, zero)
		bit_width = max(bit_width, source.Data().BitWidth())
	}

	for i := uint(0); i < nrows; i++ {
		set := false
		// Initialise delta to zero
		delta.Set(i, zero)
		// Decide which row is the winner (if any)
		for j := 0; j < nbits; j++ {
			prev := tr.Column(p.sources[j]).Get(int(i - 1))
			curr := tr.Column(p.sources[j]).Get(int(i))

			if !set && prev.Cmp(&curr) != 0 {
				var diff fr.Element

				cols[j+1].Data().Set(i, one)
				// Compute curr - prev
				if p.signs[j] {
					diff.Set(&curr)
					delta.Set(i, *diff.Sub(&diff, &prev))
				} else {
					diff.Set(&prev)
					delta.Set(i, *diff.Sub(&diff, &curr))
				}

				set = true
			} else {
				cols[j+1].Data().Set(i, zero)
			}
		}
	}
	// Done.
	return cols, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *LexicographicSort) Dependencies() []uint {
	return p.sources
}

// String returns a string representation of this constraint.  This is primarily
// used for debugging.
func (p *LexicographicSort) String() string {
	return fmt.Sprintf("(lexer (%v) (%v) :%d))", any(p.targets), p.signs, p.bitwidth)
}
