package trace

import (
	"errors"
	"fmt"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Builder is a helper utility for constructing new traces.  It simplifies
// the process of allocating module and column indices, and sanity checking
// certain invariants (e.g. that all columns in a given module have the same
// height).
type Builder struct {
	// Set of known modules
	modules []Module
	// Mapping from name to module index
	modmap map[string]uint
	//  Set of known columns
	columns []Column
}

// NewBuilder constructs an empty builder which can then be used to build a new
// trace.
func NewBuilder() *Builder {
	modules := make([]Module, 0)
	modmap := make(map[string]uint, 0)
	columns := make([]Column, 0)
	// Initially empty environment
	return &Builder{modules, modmap, columns}
}

// Build constructs a new ArrayTrace from the given configuration.
func (p *Builder) Build() Trace {
	return &ArrayTrace{p.columns, p.modules}
}

// Add a new column to this trace based on a fully qualified column name.  This
// splits the qualified column name and (if necessary) registers a new module
// with the given height.
func (p *Builder) Add(name string, padding *fr.Element, data []*fr.Element) error {
	var err error
	// Split qualified column name
	modname, colname := p.splitQualifiedColumnName(name)
	// Lookup module
	mid, ok := p.modmap[modname]
	// Register module (if not located)
	if !ok {
		if mid, err = p.Register(modname, uint(len(data))); err != nil {
			// Should be unreachable.
			return err
		}
	}
	// register new column
	return p.registerColumn(NewFieldColumn(mid, colname, data, padding))
}

// HasModule checks whether a given module has already been registered with this
// module.
func (p *Builder) HasModule(name string) bool {
	_, ok := p.modmap[name]
	return ok
}

// Register a new module with this builder.  This requires a module height which
// defines the height at which all columns in this module should be.
func (p *Builder) Register(name string, height uint) (uint, error) {
	// Sanity check module does not already exist
	if p.HasModule(name) {
		return 0, fmt.Errorf("module %s already exists", name)
	}
	//
	mid := uint(len(p.modules))
	cols := make([]uint, 0)
	// Create new module
	p.modules = append(p.modules, Module{name, cols, height})
	// Update cache
	p.modmap[name] = mid
	//
	return mid, nil
}

// SplitQualifiedColumnName splits a qualified column name into its module and
// column components.
func (p *Builder) splitQualifiedColumnName(name string) (string, string) {
	i := strings.Index(name, ".")
	if i >= 0 {
		// Split on "."
		return name[0:i], name[i+1:]
	}
	// No module name given, therefore its in the prelude.
	return "", name
}

// RegisterColumn registers a new column with this builder.  An error can arise
// if the column's module does not exist, or if the column's height does not
// match that of its enclosing module.
func (p *Builder) registerColumn(col Column) error {
	mid := col.Module()
	// Sanity check module exists
	if mid >= uint(len(p.modules)) {
		return errors.New("column has invalid enclosing module index")
	}
	// Determine enclosing module
	m := p.modules[mid]
	// Sanity check height
	if m.height != col.Height() {
		return fmt.Errorf("column '%s' height (%d) differs from enclosing module '%s' (%d)",
			col.Name(), col.Height(), m.name, m.height)
	}
	// Register new column
	cid := uint(len(p.columns))
	p.columns = append(p.columns, col)
	p.modules[mid].registerColumn(cid)
	// Done
	return nil
}
