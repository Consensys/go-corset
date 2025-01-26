package assignment

import (
	"encoding/gob"
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// SortedPermutation declares one or more columns as sorted permutations of
// existing columns.
type SortedPermutation struct {
	// Context where in which source and target columns are evaluated.
	ColumnContext tr.Context
	// Target columns declared by this sorted permutation (in the order
	// of declaration).
	Targets []sc.Column
	// Signs determines the sorting direction for each target column.
	Signs []bool
	// Source columns which define the new (sorted) columns.
	Sources []uint
}

// NewSortedPermutation creates a new sorted permutation
func NewSortedPermutation(context tr.Context, targets []sc.Column,
	signs []bool, sources []uint) *SortedPermutation {
	if len(targets) != len(signs) || len(signs) != len(sources) {
		panic("target and source column widths must match")
	}
	// Check modules
	for _, c := range targets {
		if c.Context != context {
			panic("inconsistent evaluation contexts")
		}
	}

	return &SortedPermutation{context, targets, signs, sources}
}

// Module returns the module which encloses this sorted permutation.
func (p *SortedPermutation) Module() uint {
	return p.ColumnContext.Module()
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this declaration.
func (p *SortedPermutation) Context() tr.Context {
	return p.ColumnContext
}

// Columns returns the columns declared by this sorted permutation (in the order
// of declaration).
func (p *SortedPermutation) Columns() util.Iterator[sc.Column] {
	return util.NewArrayIterator(p.Targets)
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
	data := make([]field.FrArray, len(p.Sources))
	// Construct target columns
	for i := 0; i < len(p.Sources); i++ {
		src := p.Sources[i]
		// Read column data
		src_data := trace.Column(src).Data()
		// Clone it to initialise permutation.
		data[i] = src_data.Clone()
	}
	// Sort target columns
	util.PermutationSort(data, p.Signs)
	// Physically construct the columns
	cols := make([]tr.ArrayColumn, len(p.Sources))
	//
	for i, iter := 0, p.Columns(); iter.HasNext(); i++ {
		ith := iter.Next()
		dstColName := ith.Name
		srcCol := trace.Column(p.Sources[i])
		cols[i] = tr.NewArrayColumn(ith.Context, dstColName, data[i], srcCol.Padding())
	}
	//
	return cols, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *SortedPermutation) Dependencies() []uint {
	return p.Sources
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *SortedPermutation) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	sources := sexp.EmptyList()

	for i := 0; i != len(p.Targets); i++ {
		ith := p.Targets[i].QualifiedName(schema)
		targets.Append(sexp.NewSymbol(ith))
	}

	for i, s := range p.Sources {
		ith := sc.QualifiedName(schema, s)
		if p.Signs[i] {
			ith = fmt.Sprintf("+%s", ith)
		} else {
			ith = fmt.Sprintf("-%s", ith)
		}
		//
		sources.Append(sexp.NewSymbol(ith))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("sort"),
		targets,
		sources,
	})
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Declaration(&SortedPermutation{}))
}
