package trace

import (
	"strings"
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

// ColumnIndex returns the column index of the column with the given name in
// this trace, or returns false if no such column exists.
func (p *ArrayTrace) ColumnIndex(name string) (uint, bool) {
	for i := 0; i < len(p.columns); i++ {
		c := p.columns[i]
		if c.Name() == name {
			return uint(i), true
		}
	}
	// Column does not exist
	return 0, false
}

// HasColumn checks whether the trace has a given named column (or not).
func (p *ArrayTrace) HasColumn(name string) bool {
	_, ok := p.ColumnIndex(name)
	return ok
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
	m := &p.trace.modules[column.Module()]
	// Sanity check height
	if column.Height() != m.Height() {
		panic("invalid column height")
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

// Len returns the number of items in this array.
func (p arrayTraceColumnSet) Len() uint {
	return uint(len(p.trace.columns))
}

// Swap two columns in this column set.
func (p arrayTraceColumnSet) Swap(l uint, r uint) {
	cols := p.trace.columns
	lth := cols[l]
	cols[l] = cols[r]
	cols[r] = lth
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

func (p arrayTraceModuleSet) Pad(index uint, n uint) {
	var m *Module = &p.trace.modules[index]
	m.height += n
	//
	for i := range m.columns {
		p.trace.columns[i].Pad(n)
	}
}
