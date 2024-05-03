package mir

import (
	"errors"
	"fmt"

	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
)

// Column captures the essence of a column at the MIR level.  Specifically, an
// MIR column can be lowered to an AIR column.
type Column interface {
	// hir.Column is-a Column.
	table.Column
	// // LowerTo lowers this column to an MirColumn.
	// LowerTo(*air.Schema) air.Column
}

// DataColumn represents a column of user-provided values.
type DataColumn struct {
	name string
	// A constraint on the range of values permitted for this column.
	base Type
}

// NewDataColumn constructs a new data column with a given name and base type.
func NewDataColumn(name string, base Type) *DataColumn {
	return &DataColumn{name, base}
}

// Name returns the name of this column.
func (c *DataColumn) Name() string {
	return c.name
}

// Computable determines whether or not this column can be computed from the
// existing columns of a trace.  That is, whether or not there is a known
// expression which determines the values for this column based on others in the
// trace.  Data columns are not computable.
func (c *DataColumn) Computable() bool {
	return false
}

// Get the value of this column at a given row in a given trace.
func (c *DataColumn) Get(row int, tr table.Trace) (*fr.Element, error) {
	return tr.GetByName(c.name, row)
}

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
func (c *DataColumn) Accepts(tr table.Trace) error {
	for i := 0; i < tr.Height(); i++ {
		val, err := tr.GetByName(c.name, i)
		if err != nil {
			return err
		}

		if !c.base.Accept(val) {
			// Construct useful error message
			msg := fmt.Sprintf("column %s value out-of-bounds (row %d, %s)", c.name, i, val)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}

// LowerTo lowers this datacolumn to the AIR level.  The main effect of this is
// that, for columns with non-trivial types, we must add appropriate range
// constraints to the enclosing schema.
func (c *DataColumn) LowerTo(schema *air.Schema) {
	// Check whether a constraint is implied by the column's type
	if t := c.base.AsUint(); t != nil {
		// Yes, a constraint is implied.  Now, decide whether to use a range
		// constraint or just a vanishing constraint.
		if t.HasBound(2) {
			// u1 => use vanishing constraint X * (X - 1)
			one := fr.NewElement(1)
			// Construct X
			X := &air.ColumnAccess{Column: c.name, Shift: 0}
			// Construct X-1
			X_m1 := &air.Sub{Args: []air.Expr{X, &air.Constant{Value: &one}}}
			// Construct X * (X-1)
			X_X_m1 := &air.Mul{Args: []air.Expr{X, X, X_m1}}
			//
			schema.AddVanishingConstraint(c.name, nil, X_X_m1)
		} else {
			// u2+ => use range constraint
			schema.AddRangeConstraint(c.name, t.bound)
		}
	}
	// Finally, add an (untyped) data column representing this
	// data column.
	schema.AddDataColumn(c.name)
}

// Type represents a _column type_ which restricts the set of values a column
// can take on.  For example, a column might be restricted to holding only byte
// values (i.e. in the range 0..255).
type Type interface {
	// AsUint accesses this type as an unsigned integer.  If this type is not an
	// unsigned integer, then this returns nil.
	AsUint() *UintType

	// AsField accesses this type as a field element.  If this type is not a
	// field element, then this returns nil.
	AsField() *FieldType

	// Accept checks whether a specific value is accepted by this type
	Accept(*fr.Element) bool

	// Produce a string representation of this type.
	String() string
}

// UintType represents an unsigned integer encoded using a given number of bits.
// For example, for the type "u8" then "nbits" is 8.
type UintType struct {
	// The number of bits this type represents (e.g. 8 for u8, etc).
	nbits uint
	// The numeric bound of all values in this type (e.g. 2^8 for u8, etc).
	bound *fr.Element
}

// NewUintType constructs a new integer type for a given bit width.
func NewUintType(nbits uint) *UintType {
	var maxBigInt big.Int
	// Compute 2^n
	maxBigInt.Exp(big.NewInt(2), big.NewInt(int64(nbits)), nil)
	// Construct bound
	bound := new(fr.Element)
	bound.SetBigInt(&maxBigInt)

	return &UintType{nbits, bound}
}

// AsUint accesses this type assuming it is a Uint.  Since this is the case,
// this just returns itself.
func (p *UintType) AsUint() *UintType {
	return p
}

// AsField accesses this type assuming it is a Field.  Since this is not the
// case, this returns nil.
func (p *UintType) AsField() *FieldType {
	return nil
}

// Accept determines whether a given value is an element of this type.  For
// example, 123 is an element of the type u8 whilst 256 is not.
func (p *UintType) Accept(val *fr.Element) bool {
	return val.Cmp(p.bound) < 0
}

// HasBound determines whether this type fits within a given bound.  For
// example, a u8 fits within a bound of 256 and also 65536.  However, it does
// not fit within a bound of 255.
func (p *UintType) HasBound(bound uint64) bool {
	var n fr.Element = fr.NewElement(bound)
	return p.bound.Cmp(&n) == 0
}

func (p *UintType) String() string {
	return fmt.Sprintf("u%d", p.nbits)
}

// FieldType is the type of raw field elements (normally for a prime field).
type FieldType struct {
}

// AsUint accesses this type assuming it is a Uint.  Since this is not the
// case, this returns nil.
func (p *FieldType) AsUint() *UintType {
	return nil
}

// AsField accesses this type assuming it is a Field.  Since this is the case,
// this just returns itself.
func (p *FieldType) AsField() *FieldType {
	return p
}

// Accept determines whether a given value is an element of this type.  In
// fact, all field elements are members of this type.
func (p *FieldType) Accept(val *fr.Element) bool {
	return true
}

func (p *FieldType) String() string {
	return "𝔽"
}
