package trace

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/util"
)

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace struct {
	// Holds the complete set of columns in this trace.  The index of each
	// column in this array uniquely identifies it, and is referred to as the
	// "column index".
	columns []Column
	// Holds the complete set of modules in this trace.  The index of each
	// module in this array uniquely identifies it, and is referred to as the
	// "module index".
	modules []Module
}

// Columns returns the set of columns in this trace.  Observe that mutating
// the returned array will mutate the trace.
func (p *ArrayTrace) Columns() ColumnSet {
	return arrayTraceColumnSet{p}
}

// Clone creates an identical clone of this trace.
func (p *ArrayTrace) Clone() Trace {
	clone := new(ArrayTrace)
	clone.columns = make([]Column, len(p.columns))
	clone.modules = make([]Module, len(p.modules))
	// Clone modules
	for i, m := range p.modules {
		clone.modules[i] = m.Copy()
	}
	// Clone columns
	for i, c := range p.columns {
		clone.columns[i] = c.Clone()
	}
	// done
	return clone
}

// Modules returns the set of modules in this trace.  Observe that mutating the
// returned array will mutate the trace.
func (p *ArrayTrace) Modules() ModuleSet {
	return arrayTraceModuleSet{p}
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

		modName := p.modules[ith.Context().Module()].Name()
		if modName != "" {
			id.WriteString(modName)
			id.WriteString(".")
		}

		id.WriteString(ith.Name())
		id.WriteString("={")

		for j := uint(0); j < ith.Height(); j++ {
			jth := ith.Get(int(j))

			if j != 0 {
				id.WriteString(",")
			}

			if jth == nil {
				id.WriteString("_")
			} else {
				id.WriteString(jth.String())
			}
		}
		id.WriteString("}")
	}
	id.WriteString("}")
	//
	return id.String()
}

// ============================================================================
// arrayTraceColumnSet
// ============================================================================

// arrayTraceColumnSet is an implementation of ColumnSet which maintains key
// invariants within an ArrayTrace.
type arrayTraceColumnSet struct {
	trace *ArrayTrace
}

// Add a new column to this column set.
func (p arrayTraceColumnSet) Add(column Column) uint {
	ctx := column.Context()
	m := &p.trace.modules[ctx.Module()]
	// Sanity check effective height
	if column.Height() != (ctx.LengthMultiplier() * m.Height()) {
		panic(fmt.Sprintf("invalid column height for %s: %d vs %d*%d", column.Name(),
			column.Height(), m.Height(), ctx.LengthMultiplier()))
	}
	// Proceed
	index := uint(len(p.trace.columns))
	p.trace.columns = append(p.trace.columns, column)
	// Register column with enclosing module
	m.registerColumn(index)
	// Done
	return index
}

// Get returns the ith column in this column set.
func (p arrayTraceColumnSet) Get(index uint) Column {
	return p.trace.columns[index]
}

// HasColumn checks whether a given column exists in this column set (or not).
func (p arrayTraceColumnSet) HasColumn(name string) bool {
	for _, c := range p.trace.columns {
		if c.Name() == name {
			return true
		}
	}
	// Not found
	return false
}

// IndexOf returns the column index of the column with the given name in
// this trace, or returns false if no such column exists.
func (p arrayTraceColumnSet) IndexOf(module uint, name string) (uint, bool) {
	for i := 0; i < len(p.trace.columns); i++ {
		c := p.trace.columns[i]
		if c.Context().Module() == module && c.Name() == name {
			return uint(i), true
		}
	}
	// Column does not exist
	return 0, false
}

// Len returns the number of items in this array.
func (p arrayTraceColumnSet) Len() uint {
	return uint(len(p.trace.columns))
}

// Swap two columns in this column set.
func (p arrayTraceColumnSet) Swap(l uint, r uint) {
	if l == r {
		panic("invalid column swap")
	}

	cols := p.trace.columns
	modules := p.trace.modules
	lth := cols[l]
	rth := cols[r]
	cols[l] = rth
	cols[r] = lth
	// Update modules notion of which columns they own.  Observe that this only
	// makes sense when the modules for each column differ.  Otherwise, this
	// leads to broken results.
	if lth.Context().Module() != rth.Context().Module() {
		// Extract modules being swapped
		lthmod := &modules[lth.Context().Module()]
		rthmod := &modules[rth.Context().Module()]
		// Update their columns caches
		util.ReplaceFirstOrPanic(lthmod.columns, l, r)
		util.ReplaceFirstOrPanic(rthmod.columns, r, l)
	}
}

// ============================================================================
// arrayTraceModuleSet
// ============================================================================

type arrayTraceModuleSet struct {
	trace *ArrayTrace
}

func (p arrayTraceModuleSet) Add(name string, height uint) uint {
	index := len(p.trace.modules)
	columns := make([]uint, 0)
	p.trace.modules = append(p.trace.modules, Module{name, columns, height})
	// Return module index
	return uint(index)
}

func (p arrayTraceModuleSet) Get(index uint) *Module {
	return &p.trace.modules[index]
}

// Len returns the number of items in this array.
func (p arrayTraceModuleSet) Len() uint {
	return uint(len(p.trace.modules))
}

// IndexOf returns the module index of the module with the given name in
// this trace, or returns false if no such module exists.
func (p arrayTraceModuleSet) IndexOf(name string) (uint, bool) {
	for i := 0; i < len(p.trace.modules); i++ {
		m := p.trace.modules[i]
		if m.Name() == name {
			return uint(i), true
		}
	}
	// MOdule does not exist
	return 0, false
}

func (p arrayTraceModuleSet) Swap(l uint, r uint) {
	// Swap the modules
	lth := p.trace.modules[l]
	rth := p.trace.modules[r]
	p.trace.modules[l] = rth
	p.trace.modules[r] = lth
	// Update enclosed columns
	p.reseatColumns(r, lth.columns)
	p.reseatColumns(l, rth.columns)
}

func (p arrayTraceModuleSet) Pad(index uint, n uint) {
	var m *Module = &p.trace.modules[index]
	m.height += n
	//
	for _, c := range m.columns {
		p.trace.columns[c].Pad(n)
	}
}

func (p arrayTraceModuleSet) reseatColumns(mid uint, columns []uint) {
	for _, c := range columns {
		p.trace.columns[c].Reseat(mid)
	}
}
