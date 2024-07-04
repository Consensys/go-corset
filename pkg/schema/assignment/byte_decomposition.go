package assignment

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ByteDecomposition is part of a range constraint for wide columns (e.g. u32)
// implemented using a byte decomposition.
type ByteDecomposition struct {
	// The source column being decomposed
	source uint
	// Target columns needed for decomposition
	targets []schema.Column
}

// NewByteDecomposition creates a new sorted permutation
func NewByteDecomposition(prefix string, module uint, multiplier uint, source uint, width uint) *ByteDecomposition {
	if width == 0 {
		panic("zero byte decomposition encountered")
	}
	// Define type of bytes
	U8 := schema.NewUintType(8)
	// Construct target names
	targets := make([]schema.Column, width)

	for i := uint(0); i < width; i++ {
		name := fmt.Sprintf("%s:%d", prefix, i)
		targets[i] = schema.NewColumn(module, name, multiplier, U8)
	}
	// Done
	return &ByteDecomposition{source, targets}
}

func (p *ByteDecomposition) String() string {
	return fmt.Sprintf("(decomposition #%d %d)", p.source, len(p.targets))
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the columns declared by this byte decomposition (in the order
// of declaration).
func (p *ByteDecomposition) Columns() util.Iterator[schema.Column] {
	return util.NewArrayIterator[schema.Column](p.targets)
}

// IsComputed Determines whether or not this declaration is computed.
func (p *ByteDecomposition) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// ExpandTrace expands a given trace to include the columns specified by a given
// ByteDecomposition.  This requires computing the value of each byte column in
// the decomposition.
func (p *ByteDecomposition) ExpandTrace(tr trace.Trace) error {
	columns := tr.Columns()
	// Calculate how many bytes required.
	n := len(p.targets)
	// Identify source column
	source := columns.Get(p.source)
	// Extract column data to decompose
	data := source.Data()
	// Construct byte column data
	cols := make([][]*fr.Element, n)
	// Initialise columns
	for i := 0; i < n; i++ {
		cols[i] = make([]*fr.Element, len(data))
	}
	// Decompose each row of each column
	for i := 0; i < len(data); i = i + 1 {
		ith := decomposeIntoBytes(data[i], n)
		for j := 0; j < n; j++ {
			cols[j][i] = ith[j]
		}
	}
	// Determine padding values
	padding := decomposeIntoBytes(source.Padding(), n)
	// Finally, add byte columns to trace
	for i := 0; i < n; i++ {
		ith := p.targets[i]
		columns.Add(trace.NewFieldColumn(ith.Module(), ith.Name(), ith.LengthMultiplier(), cols[i], padding[i]))
	}
	// Done
	return nil
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *ByteDecomposition) RequiredSpillage() uint {
	return uint(0)
}

// Decompose a given element into n bytes in little endian form.  For example,
// decomposing 41b into 2 bytes gives [0x1b,0x04].
func decomposeIntoBytes(val *fr.Element, n int) []*fr.Element {
	// Construct return array
	elements := make([]*fr.Element, n)

	if val == nil {
		// Special case where value being decomposed is actually undefined (i.e.
		// because its before the start of the table).  In this case, we assume
		// a default decomposition of zero.
		for i := 0; i < n; i++ {
			ith := fr.NewElement(0)
			elements[i] = &ith
		}
	} else {
		// Determine bytes of this value (in big endian form).
		bytes := val.Bytes()
		m := len(bytes) - 1
		// Convert each byte into a field element
		for i := 0; i < n; i++ {
			j := m - i
			ith := fr.NewElement(uint64(bytes[j]))
			elements[i] = &ith
		}
	}
	// Done
	return elements
}
