package trace

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// RawColumnSpec provides a specification of how a particular column should be instantiated.
type RawColumnSpec struct {
	// Module in which column exists
	Module string
	// Name of column
	Name string
	// Number of rows to instantiate
	Lines uint
}

// RawColumnEnumerator is an adaptor which surrounds an enumerator and, essentially,
// converts flat sequences of elements into arrays of raw columns.
type RawColumnEnumerator struct {
	// Column specifications
	specs []RawColumnSpec
	// Enumerate sequences of elements
	enumerator util.Enumerator[[]fr.Element]
}

// NewRawColumnEnumerator constructs an enumerator for all traces matching the
// given column specifications using elements sourced from the given pool.
func NewRawColumnEnumerator(specs []RawColumnSpec, pool []fr.Element) util.Enumerator[[]RawColumn] {
	n := uint(0)
	// Determine how many elements are required
	for _, col := range specs {
		n += col.Lines
	}
	// Construct the enumerator
	enumerator := util.EnumerateElements[fr.Element](n, pool)
	// Apply the adaptor
	return &RawColumnEnumerator{specs, enumerator}
}

// Next returns the next trace in the enumeration
func (p *RawColumnEnumerator) Next() []RawColumn {
	elems := p.enumerator.Next()
	cols := make([]RawColumn, len(p.specs))
	//
	j := 0
	// Construct each column from the sequence
	for i, c := range p.specs {
		data := util.NewFrArray(c.Lines, 256)
		// Slice nrows values from elems
		for k := uint(0); k < c.Lines; k++ {
			data.Set(k, elems[j])
			// Consume element from generated sequence
			j++
		}
		// Construct raw column
		cols[i] = RawColumn{Module: c.Module, Name: c.Name, Data: data}
	}
	// Done
	return cols
}

// HasNext checks whether the enumeration has more elements (or not).
func (p *RawColumnEnumerator) HasNext() bool {
	return p.enumerator.HasNext()
}
