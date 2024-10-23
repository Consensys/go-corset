package assignment

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// SortedPermutation declares one or more columns as sorted permutations of
// existing columns.
type SortedPermutation struct {
	// Context where this data column is located.
	context tr.Context
	// The new (sorted) columns
	targets []sc.Column
	// The sorting criteria
	signs []bool
	// The existing columns
	sources []uint
}

// NewSortedPermutation creates a new sorted permutation
func NewSortedPermutation(context tr.Context, targets []sc.Column,
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
func (p *SortedPermutation) Targets() []sc.Column {
	return p.targets
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this declaration.
func (p *SortedPermutation) Context() tr.Context {
	return p.context
}

// Columns returns the columns declared by this sorted permutation (in the order
// of declaration).
func (p *SortedPermutation) Columns() util.Iterator[sc.Column] {
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

// ComputeColumns computes the values of columns defined by this assignment.
// This requires copying the data in the source columns, and sorting that data
// according to the permutation criteria.
func (p *SortedPermutation) ComputeColumns(trace tr.Trace) ([]tr.ArrayColumn, error) {
	data := make([]util.FrArray, len(p.sources))
	// Construct target columns
	for i := 0; i < len(p.sources); i++ {
		src := p.sources[i]
		// Read column data
		src_data := trace.Column(src).Data()
		// Clone it to initialise permutation.
		data[i] = src_data.Clone()
	}
	// Sort target columns
	util.PermutationSort(data, p.signs)
	// Physically construct the columns
	cols := make([]tr.ArrayColumn, len(p.sources))
	//
	for i, iter := 0, p.Columns(); iter.HasNext(); i++ {
		ith := iter.Next()
		dstColName := ith.Name()
		srcCol := trace.Column(p.sources[i])
		cols[i] = tr.NewArrayColumn(ith.Context(), dstColName, data[i], srcCol.Padding())
	}
	//
	return cols, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *SortedPermutation) Dependencies() []uint {
	return p.sources
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *SortedPermutation) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	sources := sexp.EmptyList()

	for i := 0; i != len(p.targets); i++ {
		ith := p.targets[i].QualifiedName(schema)
		targets.Append(sexp.NewSymbol(ith))
	}

	for i, s := range p.sources {
		ith := sc.QualifiedName(schema, s)
		if p.signs[i] {
			ith = fmt.Sprintf("+%s", ith)
		} else {
			ith = fmt.Sprintf("-%s", ith)
		}
		//
		sources.Append(sexp.NewSymbol(ith))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("defpermutation"),
		targets,
		sources,
	})
}
