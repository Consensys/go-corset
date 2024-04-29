package mir

import (
	"errors"
	"fmt"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

type Column interface {
	// hir.Column is-a Column
	table.Column
	// Lower this column to an MirColumn
	LowerTo(*air.Schema) air.Column
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

func NewDataColumn(name string, base Type) *DataColumn {
       return &DataColumn{name,base}
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

func (c *DataColumn) LowerTo(schema *air.Schema) air.Column {
	// Check whether a constraint is implied by the column's type
	if t := c.base.AsUint(); t != nil {
		// Yes, a constraint is implied.  Now, decide whether
		// to use a range constraint or just a vanishing
		// constraint.
		if t.HasBound(2) {
			// u1 => use vanishing constraint X * (X - 1)
			one := fr.NewElement(1)
			// Construct X
			X := &air.ColumnAccess{Column: c.name, Shift: 0}
			// Construct X-1
			X_m1 := &air.Sub{Arguments: []air.Expr{X, &air.Constant{Value: &one}}}
			// Construct X * (X-1)
			X_X_m1 := &air.Mul{Arguments: []air.Expr{X,X,X_m1}}
			//
			schema.AddConstraint(&air.VanishingConstraint{Handle: c.name, Expr: X_X_m1})
		} else {
			// u2+ => use range constraint
			schema.AddConstraint(&air.RangeConstraint{Handle: c.name, Bound: t.Bound})
		}

	}
	return air.NewDataColumn(c.name)
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
	AsUint() *UintType

	// Access thie type as a field element.  If this type is not a
	// field element, then this returns nil.
	AsField() *FieldType

	// Check whether a specific value is accepted by this type
	Accepts(*fr.Element) bool
}

// ===================================================================
// Unsigned Integer
// ===================================================================

// Represents an unsigned integer encoded using a given number of
// bits.  For example, for the type "u8" then "NumBits" is 8.
type UintType struct {
	Bound *fr.Element
}

func (p *UintType) AsUint() *UintType {
	return p
}

func (p *UintType) AsField() *FieldType {
	return nil
}

func (p *UintType) Accepts(val *fr.Element) bool {
	return val.Cmp(p.Bound) < 0
}

func (p *UintType) HasBound(bound uint64) bool {
	var n fr.Element = fr.NewElement(bound)
	return p.Bound.Cmp(&n) == 0
}

// ===================================================================
// Field Element
// ===================================================================

// Represents a field (which is normally prime).  Amongst other
// things, this gives access to the modulus used for this field.
type FieldType struct {

}

func (p *FieldType) AsUint() *UintType {
	return nil
}

func (p *FieldType) AsField() *FieldType {
	return p
}

func (p *FieldType) Accepts(val *fr.Element) bool {
	return true
}
