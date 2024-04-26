package hir

import (
	"errors"
	"fmt"
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
	// A constraint on the range of values permitted for
	// this column.
	base Type
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

func (c *DataColumn) Accepts(tr table.Trace) error {
	for i := 0; i < tr.Height(); i++ {
		val,err := tr.GetByName(c.name,i)
		if err != nil { return err }
		if !c.base.Accepts(val) {
			// Construct useful error message
			msg := fmt.Sprintf("column %s value out-of-bounds (row %d, %s)",c.name,i,val)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}

func (c *DataColumn) LowerTo() mir.Column {
	// FIXME: we need to add constraints here!
	return mir.NewDataColumn(c.name)
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

	// Check whether a specific value is accepted by this type
	Accepts(*fr.Element) bool
}

// ===================================================================
// Unsigned Integer
// ===================================================================

// Represents an unsigned integer encoded using a given number of
// bits.  For example, for the type "u8" then "NumBits" is 8.
type Uint struct {
	Bound *fr.Element
}

func (p *Uint) AsUint() *Uint {
	return p
}

func (p *Uint) AsField() *Field {
	return nil
}

func (p *Uint) Accepts(val *fr.Element) bool {
	return val.Cmp(p.Bound) < 0
}

// ===================================================================
// Field Element
// ===================================================================

// Represents a field (which is normally prime).  Amongst other
// things, this gives access to the modulus used for this field.
type Field struct {

}

func (p *Field) AsUint() *Uint {
	return nil
}

func (p *Field) AsField() *Field {
	return p
}

func (p *Field) Accepts(val *fr.Element) bool {
	return true
}
