package assignment

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ByteDecomposition is part of a range constraint for wide columns (e.g. u32)
// implemented using a byte decomposition.
type ByteDecomposition struct {
	// The source column being decomposed
	source uint
	// Target columns needed for decomposition
	targets []sc.Column
}

// NewByteDecomposition creates a new sorted permutation
func NewByteDecomposition(prefix string, context trace.Context, source uint, width uint) *ByteDecomposition {
	if width == 0 {
		panic("zero byte decomposition encountered")
	}
	// Define type of bytes
	U8 := sc.NewUintType(8)
	// Construct target names
	targets := make([]sc.Column, width)

	for i := uint(0); i < width; i++ {
		name := fmt.Sprintf("%s:%d", prefix, i)
		targets[i] = sc.NewColumn(context, name, U8)
	}
	// Done
	return &ByteDecomposition{source, targets}
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this declaration.
func (p *ByteDecomposition) Context() trace.Context {
	return p.targets[0].Context
}

// Columns returns the columns declared by this byte decomposition (in the order
// of declaration).
func (p *ByteDecomposition) Columns() util.Iterator[sc.Column] {
	return util.NewArrayIterator[sc.Column](p.targets)
}

// IsComputed Determines whether or not this declaration is computed.
func (p *ByteDecomposition) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// ComputeColumns computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *ByteDecomposition) ComputeColumns(tr trace.Trace) ([]trace.ArrayColumn, error) {
	// Calculate how many bytes required.
	n := len(p.targets)
	// Identify source column
	source := tr.Column(p.source)
	// Determine height of column
	height := tr.Height(source.Context())
	// Determine padding values
	padding := decomposeIntoBytes(source.Padding(), n)
	// Construct byte column data
	cols := make([]trace.ArrayColumn, n)
	// Initialise columns
	for i := 0; i < n; i++ {
		ith := p.targets[i]
		// Construct a byte array for ith byte
		data := util.NewFrArray(height, 8)
		// Construct a byte column for ith byte
		cols[i] = trace.NewArrayColumn(ith.Context, ith.Name, data, padding[i])
	}
	// Decompose each row of each column
	for i := uint(0); i < height; i = i + 1 {
		ith := decomposeIntoBytes(source.Get(int(i)), n)
		for j := 0; j < n; j++ {
			cols[j].Data().Set(i, ith[j])
		}
	}
	// Done
	return cols, nil
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *ByteDecomposition) RequiredSpillage() uint {
	return uint(0)
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ByteDecomposition) Dependencies() []uint {
	return []uint{p.source}
}

// Decompose a given element into n bytes in little endian form.  For example,
// decomposing 41b into 2 bytes gives [0x1b,0x04].
func decomposeIntoBytes(val fr.Element, n int) []fr.Element {
	// Construct return array
	elements := make([]fr.Element, n)

	// Determine bytes of this value (in big endian form).
	bytes := val.Bytes()
	m := len(bytes) - 1
	// Convert each byte into a field element
	for i := 0; i < n; i++ {
		j := m - i
		ith := fr.NewElement(uint64(bytes[j]))
		elements[i] = ith
	}

	// Done
	return elements
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *ByteDecomposition) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	for _, t := range p.targets {
		targets.Append(sexp.NewSymbol(t.QualifiedName(schema)))
	}

	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("decompose"),
			targets,
			sexp.NewSymbol(sc.QualifiedName(schema, p.source)),
		})
}
