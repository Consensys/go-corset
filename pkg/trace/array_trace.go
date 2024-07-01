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
	modules []ArrayTraceModule
}

// Columns returns the set of columns in this trace.  Observe that mutating
// the returned array will mutate the trace.
func (p *ArrayTrace) Columns() ColumnSet {
	return ArrayTraceColumnSet{p}
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
	clone.modules = make([]ArrayTraceModule, len(p.modules))
	copy(clone.modules, p.modules)
	for i, c := range p.columns {
		clone.columns[i] = c.Clone()
	}
	// done
	return clone
}

// Modules returns the set of modules in this trace.  Observe that mutating the
// returned array will mutate the trace.
func (p *ArrayTrace) Modules() ModuleSet {
	return ArrayTraceModuleSet{p}
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
// ArrayTraceModule
// ============================================================================

// ArrayTraceModule describes an individual module within the trace table, and
// permits actions on the columns of this module as a whole.
type ArrayTraceModule struct {
	// Name of this module.
	name string
	// Determine the columns contained in this module by their column index.
	columns []uint
	// Height (in rows) of this module.  Specifically, every column in this
	// module must have this height.
	height uint
}

func (p *ArrayTraceModule) RegisterColumn(cid uint) {
	p.columns = append(p.columns, cid)
}

// Name of this module.
func (p *ArrayTraceModule) Name() string {
	return p.name
}

// Columns identifies the columns contained in this module by their column index.
func (p *ArrayTraceModule) Columns() []uint {
	return p.columns
}

// Height (in rows) of this module.  Specifically, every column in this
// module must have this height.
func (p *ArrayTraceModule) Height() uint {
	return p.height
}

// ============================================================================
// ArrayTraceColumnSet
// ============================================================================

// ArrayTraceColumnSet is an implementation of ColumnSet which maintains key
// invariants within an ArrayTrace.
type ArrayTraceColumnSet struct {
	trace *ArrayTrace
}

// Add a new column to this column set.
func (p ArrayTraceColumnSet) Add(column Column) uint {
	m := p.trace.modules[column.Module()]
	// Sanity check height
	if column.Height() != m.Height() {
		panic("invalid column height")
	}
	// Proceed
	index := uint(len(p.trace.columns))
	p.trace.columns = append(p.trace.columns, column)
	// Register column with enclosing module
	m.RegisterColumn(index)
	// Done
	return index
}

// Get returns the ith column in this column set.
func (p ArrayTraceColumnSet) Get(index uint) Column {
	return p.trace.columns[index]
}

// HasColumn checks whether a given column exists in this column set (or not).
func (p ArrayTraceColumnSet) HasColumn(name string) bool {
	for _, c := range p.trace.columns {
		if c.Name() == name {
			return true
		}
	}
	// Not found
	return false
}

// Len returns the number of items in this array.
func (p ArrayTraceColumnSet) Len() uint {
	return uint(len(p.trace.columns))
}

// Swap two columns in this column set.
func (p ArrayTraceColumnSet) Swap(l uint, r uint) {
	cols := p.trace.columns
	lth := cols[l]
	cols[l] = cols[r]
	cols[r] = lth
}

// ============================================================================
// ArrayTraceModuleSet
// ============================================================================

type ArrayTraceModuleSet struct {
	trace *ArrayTrace
}

func (p ArrayTraceModuleSet) Add(name string, height uint) uint {
	index := len(p.trace.modules)
	columns := make([]uint, 0)
	p.trace.modules = append(p.trace.modules, ArrayTraceModule{name, columns, height})
	return uint(index)
}

func (p ArrayTraceModuleSet) Get(index uint) Module {
	return &p.trace.modules[index]
}

// Len returns the number of items in this array.
func (p ArrayTraceModuleSet) Len() uint {
	return uint(len(p.trace.modules))
}

func (p ArrayTraceModuleSet) Pad(index uint, n uint) {
	var m *ArrayTraceModule = &p.trace.modules[index]
	m.height += n
	//
	for i := range m.columns {
		p.trace.columns[i].Pad(n)
	}
}
