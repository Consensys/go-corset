package assignment

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Interleaving generates a new column by interleaving two or more existing
// colummns.  For example, say Z interleaves X and Y (in that order) and we have
// a trace X=[1,2], Y=[3,4].  Then, the interleaved column Z has the values
// Z=[1,3,2,4].
type Interleaving struct {
	// Module where this interleaving is located.
	module uint
	// The new (interleaved) column
	target schema.Column
	// The source columns
	sources []uint
}

// NewInterleaving constructs a new interleaving assignment.
func NewInterleaving(module uint, name string, multiplier uint, sources []uint) *Interleaving {
	// Update multiplier
	multiplier = multiplier * uint(len(sources))
	// Fixme: determine interleaving type
	target := schema.NewColumn(module, name, multiplier, &schema.FieldType{})

	return &Interleaving{module, target, sources}
}

// Module returns the module which encloses this interleaving.
func (p *Interleaving) Module() uint {
	return p.module
}

// Sources returns the columns used by this interleaving to define the new
// (interleaved) column.
func (p *Interleaving) Sources() []uint {
	return p.sources
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the column declared by this interleaving.
func (p *Interleaving) Columns() util.Iterator[schema.Column] {
	return util.NewUnitIterator(p.target)
}

// IsComputed Determines whether or not this declaration is computed (which an
// interleaving column is by definition).
func (p *Interleaving) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *Interleaving) RequiredSpillage() uint {
	return uint(0)
}

// ExpandTrace expands a given trace to include the columns specified by a given
// Interleaving.  This requires copying the data in the source columns to create
// the interleaved column.
func (p *Interleaving) ExpandTrace(tr tr.Trace) error {
	columns := tr.Columns()
	// Ensure target column doesn't exist
	for i := p.Columns(); i.HasNext(); {
		name := i.Next().Name()
		// Sanity check no column already exists with this name.
		if columns.HasColumn(name) {
			return fmt.Errorf("column already exists ({%s})", name)
		}
	}
	// Determine interleaving width
	width := uint(len(p.sources))
	// Following division should always produce whole value because the length
	// multiplier already includes the width as a factor.
	multiplier := p.target.LengthMultiplier() / width
	// Determine module height (as this can be used to determine the height of
	// the interleaved column)
	height := tr.Modules().Get(p.module).Height() * multiplier
	// Construct empty array
	data := make([]*fr.Element, height*width)
	// Offset just gives the column index
	offset := uint(0)
	// Copy interleaved data
	for i := uint(0); i < width; i++ {
		// Lookup source column
		col := tr.Columns().Get(p.sources[i])
		// Copy over
		for j := uint(0); j < height; j++ {
			data[offset+(j*width)] = col.Get(int(j))
		}

		offset++
	}
	// Padding for the entire column is determined by the padding for the first
	// column in the interleaving.
	padding := columns.Get(0).Padding()
	// Colunm needs to be expanded.
	columns.Add(trace.NewFieldColumn(p.module, p.target.Name(), multiplier*width, data, padding))
	//
	return nil
}
