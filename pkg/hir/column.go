package hir

import (
	"math/big"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

type Column interface {
	// hir.Column is-a Column
	table.Column
	// Lower this column to an MirColumn
	LowerTo() mir.Column
}

// ===================================================================
// Column
// ===================================================================

type DataColumn struct {
	name string
}

func (c *DataColumn) Name() string {
	return c.name
}

func (c *DataColumn) Computable() bool {
	return false
}

func (c *DataColumn) Get(row int, tr table.Trace) (*fr.Element,error) {
	return tr.GetByName(c.name,row)
}

func (c *DataColumn) LowerTo() mir.Column {
	// FIXME: this is only temporary.
	return c
}

// ===================================================================
// Column Type
// ===================================================================

// Represents a _column type_ which restricts the set of values a
// column can take on.  For example, a column might be restricted to
// holding only byte values (i.e. in the range 0..255).
type Type interface {
	// Access thie type as a unsigned integer.  If this type is not an
	// unsigned integer, then this returns nil.
	AsUint() *Uint

	// Access thie type as a field element.  If this type is not a
	// field element, then this returns nil.
	AsField() *Field
}

// ===================================================================
// Unsigned Integer
// ===================================================================

// Represents an unsigned integer encoded using a given number of
// bits.  For example, for the type "u8" then "NumBits" is 8.
type Uint struct {
	NumBits int
}

func (p *Uint) AsInt() *Uint {
	return p
}

func (p *Uint) AsField() *Field {
	return nil
}

// ===================================================================
// Field Element
// ===================================================================

// Represents a field (which is normally prime).  Amongst other
// things, this gives access to the modulus used for this field.
type Field struct {
	Modulus *big.Int
}

func (p *Field) AsInt() *Uint {
	return nil
}

func (p *Field) AsField() *Field {
	return p
}
