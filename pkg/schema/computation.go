package schema

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// TraceComputation represents a computation which is applied to a
// high-level trace in order to expand it to a low-level trace.  This
// typically involves adding columns, evaluating compute-only
// expressions, sorting columns, etc.
type TraceComputation interface {
	Acceptable
	// ExpandTrace expands a given trace to include "computed
	// columns".  These are columns which do not exist in the
	// original trace, but are added during trace expansion to
	// form the final trace.
	ExpandTrace(tr.Trace) error
	// RequiredSpillage returns the minimum amount of spillage required to ensure
	// valid traces are accepted in the presence of arbitrary padding.  Note,
	// spillage is currently assumed to be required only at the front of a
	// trace.
	RequiredSpillage() uint
}

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
func (p *ByteDecomposition) Accepts(tr tr.Trace) error {
	n := int(p.BitWidth / 8)
	//
	for i := 0; i < n; i++ {
		colName := fmt.Sprintf("%s:%d", p.Target, i)
		if !tr.HasColumn(colName) {
			return fmt.Errorf("tr.Trace missing byte decomposition column ({%s})", colName)
		}
	}
	// Done
	return nil
}

// ExpandTrace expands a given trace to include the columns specified by a given
// ByteDecomposition.  This requires computing the value of each byte column in
// the decomposition.
func (p *ByteDecomposition) ExpandTrace(tr tr.Trace) error {
	// Calculate how many bytes required.
	n := int(p.BitWidth / 8)
	// Identify target column
	target := tr.ColumnByName(p.Target)
	// Extract column data to decompose
	data := tr.ColumnByName(p.Target).Data()
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
	padding := decomposeIntoBytes(target.Padding(), n)
	// Finally, add byte columns to trace
	for i := 0; i < n; i++ {
		col := fmt.Sprintf("%s:%d", p.Target, i)
		tr.AddColumn(col, cols[i], padding[i])
	}
	// Done
	return nil
}

func (p *ByteDecomposition) String() string {
	return fmt.Sprintf("(decomposition %s %d)", p.Target, p.BitWidth)
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
