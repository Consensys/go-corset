package trace

import (
	"fmt"
	"math"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace struct {
	// Holds the complete set of columns in this trace.  The index of each
	// column in this array uniquely identifies it, and is referred to as the
	// "column index".
	columns []ArrayColumn
	// Holds the height of each module in this trace.  The index of each
	// module in this array uniquely identifies it, and is referred to as the
	// "module index".
	modules []ArrayModule
}

// NewArrayTrace constructs a trace from a given set of indexed modules and columns.
func NewArrayTrace(modules []ArrayModule, columns []ArrayColumn) *ArrayTrace {
	return &ArrayTrace{columns, modules}
}

// Modules returns an iterator over the modules in this trace.
func (p *ArrayTrace) Modules() util.Iterator[ArrayModule] {
	return util.NewArrayIterator(p.modules)
}

// Width returns number of columns in this trace.
func (p *ArrayTrace) Width() uint {
	return uint(len(p.columns))
}

// Height returns the height of a given context (i.e. module) in the trace.
func (p *ArrayTrace) Height(ctx Context) uint {
	return p.modules[ctx.module].height * ctx.multiplier
}

// Column returns a given column in this trace.
func (p *ArrayTrace) Column(cid uint) Column {
	return &p.columns[cid]
}

// FillColumn sets the data and padding for the given column.  This will panic
// if the data is already set.
func (p *ArrayTrace) FillColumn(cid uint, data util.FrArray, padding fr.Element) {
	// Find column to fill
	col := &p.columns[cid]
	// Find enclosing module
	mod := &p.modules[col.Context().Module()]
	// Determine appropriate length multiplier
	multiplier := col.context.multiplier
	// Sanity check this column has not already been filled.
	if data.Len()%multiplier != 0 {
		colname := QualifiedColumnName(mod.name, col.name)
		panic(fmt.Sprintf("column %s has invalid length multiplier (%d indivisible by %d)",
			colname, data.Len(), multiplier))
	} else if mod.height == math.MaxUint {
		// Initialise column height
		mod.height = data.Len() / col.context.multiplier
	} else if data.Len() != p.Height(col.Context()) {
		colname := QualifiedColumnName(mod.name, col.name)
		panic(fmt.Sprintf("column %s has invalid height (%d but expected %d)", colname, data.Len(), mod.height*multiplier))
	}
	// Fill the column
	col.fill(data, padding)
}

// Pad pads a given module with a given number of padding rows.
func (p *ArrayTrace) Pad(module uint, n uint) {
	p.modules[module].height += n
	// Padd each column contained within this module.
	for i := 0; i < len(p.columns); i++ {
		c := &p.columns[i]
		if c.context.module == module {
			c.pad(n)
		}
	}
}

func (p *ArrayTrace) String() string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder

	id.WriteString("{")

	for i := 0; i < len(p.columns); i++ {
		ith := p.columns[i]

		if i != 0 {
			id.WriteString(",")
		}

		modName := p.modules[ith.Context().Module()].name
		if modName != "" {
			id.WriteString(modName)
			id.WriteString(".")
		}

		id.WriteString(ith.Name())
		// Sanity check whether filled or not.
		if ith.Data() == nil {
			id.WriteString("=⊥")
		} else {
			id.WriteString("={")
			// Print out each element
			for j := uint(0); j < ith.Height(); j++ {
				jth := ith.Get(int(j))

				if j != 0 {
					id.WriteString(",")
				}

				id.WriteString(jth.String())
			}

			id.WriteString("}")
		}
	}

	id.WriteString("}")
	//
	return id.String()
}

// ----------------------------------------------------------------------------

// ArrayModule describes an individual module within a trace.
type ArrayModule struct {
	// Holds the name of this module
	name string
	// Holds the height of all columns within this module.
	height uint
}

// EmptyArrayModule constructs a module with the given name and an (as yet)
// unspecified height.
func EmptyArrayModule(name string) ArrayModule {
	return ArrayModule{name, math.MaxUint}
}

// Name returns the name of this module.
func (p ArrayModule) Name() string {
	return p.name
}

// Height returns the height of this module, meaning the number of assigned
// rows.
func (p ArrayModule) Height() uint {
	return p.height
}

// ----------------------------------------------------------------------------

// ArrayColumn describes an individual column of data within a trace table.
type ArrayColumn struct {
	// Evaluation context of this column
	context Context
	// Holds the name of this column
	name string
	// Holds the raw data making up this column
	data util.FrArray
	// Value to be used when padding this column
	padding fr.Element
}

// NewArrayColumn constructs a  with the give name, data and padding.
func NewArrayColumn(context Context, name string, data util.FrArray,
	padding fr.Element) ArrayColumn {
	col := EmptyArrayColumn(context, name)
	col.fill(data, padding)
	//
	return col
}

// EmptyArrayColumn constructs a  with the give name, data and padding.
func EmptyArrayColumn(context Context, name string) ArrayColumn {
	return ArrayColumn{context, name, nil, fr.NewElement(0)}
}

// Context returns the evaluation context this column provides.
func (p *ArrayColumn) Context() Context {
	return p.context
}

// Name returns the name of the given column.
func (p *ArrayColumn) Name() string {
	return p.name
}

// Height determines the height of this column.
func (p *ArrayColumn) Height() uint {
	return p.data.Len()
}

// Padding returns the value which will be used for padding this column.
func (p *ArrayColumn) Padding() fr.Element {
	return p.padding
}

// Data provides access to the underlying data of this column
func (p *ArrayColumn) Data() util.FrArray {
	return p.data
}

// Get the value at a given row in this column.  If the row is
// out-of-bounds, then the column's padding value is returned instead.
// Thus, this function always succeeds.
func (p *ArrayColumn) Get(row int) fr.Element {
	if row < 0 || uint(row) >= p.data.Len() {
		// out-of-bounds access
		return p.padding
	}
	// in-bounds access
	return p.data.Get(uint(row))
}

func (p *ArrayColumn) fill(data util.FrArray, padding fr.Element) {
	// Sanity check this column has not already been filled.
	if p.data != nil {
		panic(fmt.Sprintf("computed column %s has already been filled", p.name))
	} else if data.Len()%p.context.LengthMultiplier() != 0 {
		panic(fmt.Sprintf("computed column %s filling has invalid length multiplier", p.name))
	}
	// Fill the column
	p.data = data
	p.padding = padding
}

func (p *ArrayColumn) pad(n uint) {
	// FIXME: have to avoid attempting to pad a computed column which has not
	// yet neen computed (i.e. because we are applying spillage).  Somehow, this
	// special casing does not feel right to me.
	if p.data != nil {
		// Apply the length multiplier
		n = n * p.context.LengthMultiplier()
		// Pad front of array
		p.data = p.data.PadFront(n, p.padding)
	}
}
