package assignment

import (
	"encoding/gob"
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Interleaving generates a new column by interleaving two or more existing
// colummns.  For example, say Z interleaves X and Y (in that order) and we have
// a trace X=[1,2], Y=[3,4].  Then, the interleaved column Z has the values
// Z=[1,3,2,4].
type Interleaving struct {
	// The new (interleaved) column
	Target sc.Column
	// Sources are the columns used by this interleaving to define the new
	// (interleaved) column.
	Sources []uint
}

// NewInterleaving constructs a new interleaving assignment.
func NewInterleaving(context tr.Context, name string, sources []uint, datatype sc.Type) *Interleaving {
	if context.LengthMultiplier()%uint(len(sources)) != 0 {
		panic(fmt.Sprintf("length multiplier (%d) for column %s not divisible by number of columns (%d)",
			context.LengthMultiplier(), name, len(sources)))
	}
	// Fixme: determine interleaving type
	target := sc.NewColumn(context, name, datatype)

	return &Interleaving{target, sources}
}

// Module returns the module which encloses this interleaving.
func (p *Interleaving) Module() uint {
	return p.Target.Context.Module()
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this interleaving.
func (p *Interleaving) Context() tr.Context {
	return p.Target.Context
}

// Columns returns the column declared by this interleaving.
func (p *Interleaving) Columns() util.Iterator[sc.Column] {
	return util.NewUnitIterator(p.Target)
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

// ComputeColumns computes the values of columns defined by this assignment.
// This requires copying the data in the source columns to create the
// interleaved column.
func (p *Interleaving) ComputeColumns(trace tr.Trace) ([]tr.ArrayColumn, error) {
	ctx := p.Target.Context
	// Byte width records the largest width of any column.
	bit_width := uint(0)
	// Ensure target column doesn't exist
	for i := p.Columns(); i.HasNext(); {
		ith := i.Next()
		// Update byte width
		bit_width = max(bit_width, ith.DataType.BitWidth())
	}
	// Determine interleaving width
	width := uint(len(p.Sources))
	// Following division should always produce whole value because the length
	// multiplier already includes the width as a factor.
	height := trace.Height(ctx) / width
	// Construct empty array
	data := util.NewFrArray(height*width, bit_width)
	// Offset just gives the column index
	offset := uint(0)
	// Copy interleaved data
	for i := uint(0); i < width; i++ {
		// Lookup source column
		col := trace.Column(p.Sources[i])
		// Copy over
		for j := uint(0); j < height; j++ {
			data.Set(offset+(j*width), col.Get(int(j)))
		}

		offset++
	}
	// Padding for the entire column is determined by the padding for the first
	// column in the interleaving.
	padding := trace.Column(0).Padding()
	// Colunm needs to be expanded.
	col := tr.NewArrayColumn(ctx, p.Target.Name, data, padding)
	//
	return []tr.ArrayColumn{col}, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *Interleaving) Dependencies() []uint {
	return p.Sources
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *Interleaving) Lisp(schema sc.Schema) sexp.SExp {
	target := sexp.NewSymbol(p.Target.QualifiedName(schema))
	sources := sexp.EmptyList()
	// Convert source columns
	for _, src := range p.Sources {
		sources.Append(sexp.NewSymbol(sc.QualifiedName(schema, src)))
	}
	// Add datatype (if non-field)
	datatype := sexp.NewSymbol(p.Target.DataType.String())
	multiplier := sexp.NewSymbol(fmt.Sprintf("x%d", p.Target.Context.LengthMultiplier()))
	def := sexp.NewList([]sexp.SExp{target, datatype, multiplier})
	// Construct S-Expression
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("interleaved"),
		def,
		sources,
	})
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Declaration(&Interleaving{}))
}
