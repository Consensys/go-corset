package table

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ByteDecomposition is part of a range constraint for wide columns (e.g. u32)
// implemented using a byte decomposition.
type ByteDecomposition struct {
	// The target column being decomposed
	Target string
	// The bitwidth of the target column
	BitWidth uint
}

// NewByteDecomposition creates a new sorted permutation
func NewByteDecomposition(target string, width uint) *ByteDecomposition {
	if width%8 != 0 {
		panic("asymetric byte decomposition not yet supported")
	} else if width == 0 {
		panic("zero byte decomposition encountered")
	}

	return &ByteDecomposition{target, width}
}

// Accepts checks whether a given trace has the necessary columns
func (p *ByteDecomposition) Accepts(tr Trace) error {
	n := int(p.BitWidth / 8)
	//
	for i := 0; i < n; i++ {
		colName := fmt.Sprintf("%s:%d", p.Target, i)
		if !tr.HasColumn(colName) {
			return fmt.Errorf("Trace missing byte decomposition column ({%s})", colName)
		}
	}
	// Done
	return nil
}

// ExpandTrace expands a given trace to include the columns specified by a given
// ByteDecomposition.  This requires computing the value of each byte column in
// the decomposition.
func (p *ByteDecomposition) ExpandTrace(tr Trace) error {
	// Calculate how many bytes required.
	n := int(p.BitWidth / 8)
	// Extract column data to decompose
	data := tr.ColumnByName(p.Target)
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
	// Finally, add byte columns to trace
	for i := 0; i < n; i++ {
		col := fmt.Sprintf("%s:%d", p.Target, i)
		tr.AddColumn(col, cols[i])
	}
	// Done
	return nil
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
