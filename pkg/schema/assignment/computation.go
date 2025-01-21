package assignment

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
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
	var (
		fn NativeComputation
		ok bool
	)
	// Sanity check
	if fn, ok = NATIVES[p.Name]; !ok {
		panic(fmt.Sprintf("unknown native function: %s", p.Name))
	}
	// Proceed
	targets := make([]tr.ArrayColumn, len(p.Targets))
	// Apply native function (or panic if none exists)
	data := fn.Function(trace, p.Sources)
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

// NativeComputation embeds information about a support native computation.
// This can be used, for example, to check that a native function is being
// called correctly, etc.
type NativeComputation struct {
	// Function which will be applied to a given set of input columns, whilst
	// writing to a given set of output columns.
	Function func(tr.Trace, []uint) []util.FrArray
}

// NATIVES map holds the supported set of native computations.
var NATIVES map[string]NativeComputation = map[string]NativeComputation{
	"id":                   {idNativeFunction},
	"filter":               {filterNativeFunction},
	"map-if":               {mapIfNativeFunction},
	"fwd-changes-within":   {fwdChangesWithinNativeFunction},
	"fwd-unchanged-within": {fwdUnchangedWithinNativeFunction},
	"bwd-changes-within":   {bwdChangesWithinNativeFunction},
	"fwd-fill-within":      {fwdFillWithinNativeFunction},
	"bwd-fill-within":      {bwdFillWithinNativeFunction},
}

// id assigns the target column with the corresponding value of the source
// column
func idNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) != 1 {
		panic("incorrect number of arguments")
	}
	// Clone source column
	data := trace.Column(sources[0]).Data().Clone()
	// Done
	return []util.FrArray{data}
}

// filter assigns the target column with the corresponding value of the source
// column *when* a given selector column is non-zero.  Otherwise, the target
// column remains zero at the given position.
func filterNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) != 2 {
		panic("incorrect number of arguments")
	}
	// Extract input column info
	src_col := trace.Column(sources[0]).Data()
	sel_col := trace.Column(sources[1]).Data()
	// Clone source column
	data := util.NewFrArray(src_col.Len(), src_col.BitWidth())
	//
	for i := uint(0); i < data.Len(); i++ {
		selector := sel_col.Get(i)
		// Check whether selctor non-zero
		if !selector.IsZero() {
			ith_value := src_col.Get(i)
			data.Set(i, ith_value)
		}
	}
	// Done
	return []util.FrArray{data}
}

// apply a key-value map conditionally.
func mapIfNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	n := len(sources) - 3
	if n%2 != 0 {
		panic(fmt.Sprintf("map-if expects 3 + 2*n columns (given %d)", len(sources)))
	}
	//
	n = n / 2
	// Setup what we need
	source_selector := trace.Column(sources[1+n]).Data()
	source_keys := make([]util.Array[fr.Element], n)
	source_value := trace.Column(sources[2+n+n]).Data()
	source_map := util.NewHashMap[util.BytesKey, fr.Element](source_value.Len())
	target_selector := trace.Column(sources[0]).Data()
	target_keys := make([]util.Array[fr.Element], n)
	target_value := util.NewFrArray(target_selector.Len(), source_value.BitWidth())
	// Initialise source / target keys
	for i := 0; i < n; i++ {
		target_keys[i] = trace.Column(sources[1+i]).Data()
		source_keys[i] = trace.Column(sources[2+n+i]).Data()
	}
	// Build source map
	for i := uint(0); i < source_value.Len(); i++ {
		ith_selector := source_selector.Get(i)
		if !ith_selector.IsZero() {
			ith_value := source_value.Get(i)
			ith_key := extractIthKey(i, source_keys)
			//
			if val, ok := source_map.Get(ith_key); ok && val.Cmp(&ith_value) != 0 {
				// Conflicting item already in map, so fail with useful error.
				ith_row := extractIthColumns(i, source_keys)
				lhs := fmt.Sprintf("%v=>%s", ith_row, ith_value.String())
				rhs := fmt.Sprintf("%v=>%s", ith_row, val.String())
				panic(fmt.Sprintf("conflicting values in source map (row %d): %s vs %s", i, lhs, rhs))
			} else if !ok {
				// Item not previously in map
				source_map.Insert(ith_key, ith_value)
			}
		}
	}
	// Construct target value column
	for i := uint(0); i < target_value.Len(); i++ {
		ith_selector := target_selector.Get(i)
		if !ith_selector.IsZero() {
			ith_key := extractIthKey(i, target_keys)
			//nolint:revive
			if val, ok := source_map.Get(ith_key); !ok {
				// Couldn't find key in source map, so fail with useful error.
				ith_row := extractIthColumns(i, target_keys)
				panic(fmt.Sprintf("target key (%v) missing from source map", ith_row))
			} else {
				// Assign target value
				target_value.Set(i, val)
			}
		}
	}
	// Done
	return []util.FrArray{target_value}
}

func extractIthKey(index uint, cols []util.Array[fr.Element]) util.BytesKey {
	// Each fr.Element is 4 x 64bit words.
	bytes := make([]byte, 32*len(cols))
	// Slice provides an access window for writing
	slice := bytes
	// Evaluate each expression in turn
	for i := 0; i < len(cols); i++ {
		ith := cols[i].Get(index)
		// Copy over each element
		binary.BigEndian.PutUint64(slice, ith[0])
		binary.BigEndian.PutUint64(slice[8:], ith[1])
		binary.BigEndian.PutUint64(slice[16:], ith[2])
		binary.BigEndian.PutUint64(slice[24:], ith[3])
		// Move slice over
		slice = slice[32:]
	}
	// Done
	return util.NewBytesKey(bytes)
}

// determines changes of a given set of columns within a given region.
func fwdChangesWithinNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) < 2 {
		panic("incorrect number of arguments")
	}
	// Useful constant
	one := fr.One()
	// Extract input column info
	selector_col := trace.Column(sources[0]).Data()
	source_cols := make([]util.Array[fr.Element], len(sources)-1)
	//
	for i := 1; i < len(sources); i++ {
		source_cols[i-1] = trace.Column(sources[i]).Data()
	}
	// Construct (binary) output column
	data := util.NewFrArray(selector_col.Len(), 1)
	// Set current value
	current := make([]fr.Element, len(source_cols))
	started := false
	//
	for i := uint(0); i < selector_col.Len(); i++ {
		ith_selector := selector_col.Get(i)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			//
			row := extractIthColumns(i, source_cols)
			// Trigger required?
			if !started || !slices.Equal(current, row) {
				started = true
				current = row
				//
				data.Set(i, one)
			}
		}
	}
	// Done
	return []util.FrArray{data}
}

func fwdUnchangedWithinNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) < 2 {
		panic("incorrect number of arguments")
	}
	// Useful constant
	one := fr.One()
	zero := fr.NewElement(0)
	// Extract input column info
	selector_col := trace.Column(sources[0]).Data()
	source_cols := make([]util.Array[fr.Element], len(sources)-1)
	//
	for i := 1; i < len(sources); i++ {
		source_cols[i-1] = trace.Column(sources[i]).Data()
	}
	// Construct (binary) output column
	data := util.NewFrArray(selector_col.Len(), 1)
	// Set current value
	current := make([]fr.Element, len(source_cols))
	started := false
	//
	for i := uint(0); i < selector_col.Len(); i++ {
		ith_selector := selector_col.Get(i)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			//
			row := extractIthColumns(i, source_cols)
			// Trigger required?
			if !started || !slices.Equal(current, row) {
				started = true
				current = row
				//
				data.Set(i, zero)
			} else {
				data.Set(i, one)
			}
		}
	}
	// Done
	return []util.FrArray{data}
}

// determines changes of a given set of columns within a given region.
func bwdChangesWithinNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) < 2 {
		panic("incorrect number of arguments")
	}
	// Useful constant
	one := fr.One()
	// Extract input column info
	selector_col := trace.Column(sources[0]).Data()
	source_cols := make([]util.Array[fr.Element], len(sources)-1)
	//
	for i := 1; i < len(sources); i++ {
		source_cols[i-1] = trace.Column(sources[i]).Data()
	}
	// Construct (binary) output column
	data := util.NewFrArray(selector_col.Len(), 1)
	// Set current value
	current := make([]fr.Element, len(source_cols))
	started := false
	//
	for i := selector_col.Len(); i > 0; i-- {
		ith_selector := selector_col.Get(i - 1)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			//
			row := extractIthColumns(i-1, source_cols)
			// Trigger required?
			if !started || !slices.Equal(current, row) {
				started = true
				current = row
				//
				data.Set(i-1, one)
			}
		}
	}
	// Done
	return []util.FrArray{data}
}

func fwdFillWithinNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) != 3 {
		panic("incorrect number of arguments")
	}
	// Extract input column info
	selector_col := trace.Column(sources[0]).Data()
	first_col := trace.Column(sources[1]).Data()
	source_col := trace.Column(sources[2]).Data()
	// Construct (binary) output column
	data := util.NewFrArray(source_col.Len(), source_col.BitWidth())
	// Set current value
	current := fr.NewElement(0)
	//
	for i := uint(0); i < selector_col.Len(); i++ {
		ith_selector := selector_col.Get(i)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			ith_first := first_col.Get(i)
			//
			if !ith_first.IsZero() {
				current = source_col.Get(i)
			}
			//
			data.Set(i, current)
		}
	}
	// Done
	return []util.FrArray{data}
}

func bwdFillWithinNativeFunction(trace tr.Trace, sources []uint) []util.FrArray {
	if len(sources) != 3 {
		panic("incorrect number of arguments")
	}
	// Extract input column info
	selector_col := trace.Column(sources[0]).Data()
	first_col := trace.Column(sources[1]).Data()
	source_col := trace.Column(sources[2]).Data()
	// Construct (binary) output column
	data := util.NewFrArray(source_col.Len(), source_col.BitWidth())
	// Set current value
	current := fr.NewElement(0)
	//
	for i := selector_col.Len(); i > 0; i-- {
		ith_selector := selector_col.Get(i - 1)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			ith_first := first_col.Get(i - 1)
			//
			if !ith_first.IsZero() {
				current = source_col.Get(i - 1)
			}
			//
			data.Set(i-1, current)
		}
	}
	// Done
	return []util.FrArray{data}
}

func extractIthColumns(index uint, cols []util.Array[fr.Element]) []fr.Element {
	row := make([]fr.Element, len(cols))
	//
	for i := range row {
		row[i] = cols[i].Get(index)
	}
	//
	return row
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Declaration(&Computation{}))
}
