package assignment

import (
	"encoding/gob"
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Computation currently describes a native computation which accepts a set of
// input columns, and assigns a set of output columns.
type Computation struct {
	// Context where in which source and target columns exist.
	ColumnContext tr.Context
	// Name of the function being invoked.
	Name string
	// Target columns declared by this sorted permutation (in the order
	// of declaration).
	Targets []sc.Column
	// Source columns which define the new (sorted) columns.
	Sources []uint
}

// NativeComputation embeds information about a support native computation.
// This can be used, for example, to check that a native function is being
// called correctly, etc.
type NativeComputation struct {
	// Name of this computation
	Name string
	// Function which will be applied to a given set of input columns, whilst
	// writing to a given set of output columns.
	Function func(tr.Trace, []uint) []util.FrArray
}

// NATIVES map holds the supported set of native computations.
var NATIVES map[string]NativeComputation = map[string]NativeComputation{
	"id": {"id", idNativeFunction},
}

// NewComputation defines a set of target columns which are assigned from a
// given set of source columns using a function to multiplex input to output.
func NewComputation(context tr.Context, functionName string, targets []sc.Column, sources []uint) *Computation {
	// Sanity checks
	for _, c := range targets {
		if c.Context != context {
			err := fmt.Sprintf("inconsistent evaluation contexts (%s vs %s)", c.Context, context)
			panic(err)
		}
	}
	//
	return &Computation{context, functionName, targets, sources}
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this computed column.
func (p *Computation) Context() trace.Context {
	return p.ColumnContext
}

// Columns returns the columns declared by this computed column.
func (p *Computation) Columns() util.Iterator[sc.Column] {
	return util.NewArrayIterator[sc.Column](p.Targets)
}

// IsComputed Determines whether or not this declaration is computed (which it
// is).
func (p *Computation) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *Computation) RequiredSpillage() uint {
	return uint(0)
}

// ComputeColumns computes the values of columns defined by this assignment.
// This requires copying the data in the source columns, and sorting that data
// according to the permutation criteria.
func (p *Computation) ComputeColumns(trace tr.Trace) ([]tr.ArrayColumn, error) {
	targets := make([]tr.ArrayColumn, len(p.Targets))
	// Apply native function (or panic if none exists)
	data := NATIVES[p.Name].Function(trace, p.Sources)
	// Physically construct target columns
	for i, iter := 0, p.Columns(); iter.HasNext(); i++ {
		ith := iter.Next()
		dstColName := ith.Name
		srcCol := trace.Column(p.Sources[i])
		targets[i] = tr.NewArrayColumn(ith.Context, dstColName, data[i], srcCol.Padding())
	}
	//
	return targets, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *Computation) Dependencies() []uint {
	return p.Sources
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *Computation) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	sources := sexp.EmptyList()

	for i := 0; i != len(p.Targets); i++ {
		ith := p.Targets[i].QualifiedName(schema)
		targets.Append(sexp.NewSymbol(ith))
	}

	for _, s := range p.Sources {
		ith := sc.QualifiedName(schema, s)
		sources.Append(sexp.NewSymbol(ith))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("compute"),
		targets,
		sexp.NewSymbol(p.Name),
		sources,
	})
}

// ============================================================================
// Native Function Definitions
// ============================================================================

func idNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) != 1 {
		panic("incorrect number of arguments")
	}
	// Clone source column
	data := trace.Column(sources[0]).Data().Clone()
	// Done
	return []util.FrArray{data}
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Declaration(&Computation{}))
}
