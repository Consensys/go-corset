package corset

import (
	"encoding/gob"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
)

// SourceMap is a binary file attribute which provides debugging
// information about the relationship between registers and source-level
// columns.  This is used, for example, within the inspector.
type SourceMap struct {
	// Root module correspond to the top-level HIR modules.  Thus, indicates into
	// this table correspond to HIR module indices, etc.
	Root SourceModule
}

// AttributeName returns the name of the binary file attribute that this will
// generate.  This is used, for example, when listing attributes contained
// within a binary file.
func (p *SourceMap) AttributeName() string {
	return "CorsetSourceMap"
}

// Flattern modules in this tree matching a given criteria
func (p *SourceMap) Flattern(predicate func(*SourceModule) bool) []SourceModule {
	return p.Root.Flattern(predicate)
}

// SourceModule represents an entity at the source-level which groups together
// related columns.  Modules can be either concrete (in which case they
// correspond with HIR modules) or virtual (in which case they are encoded
// within an HIR module).
type SourceModule struct {
	// Name of this submodule.
	Name string
	// Synthetic indicates whether or not this module was automatically
	// generated or not.
	Synthetic bool
	Virtual   bool
	// Selector determines when this (sub)module is active.  Specifically, when
	// it evaluates to a non-zero value the module is active.
	Selector hir.Expr
	// Submodules identifies any (virtual) submodules contained within this.
	// Currently, perspectives are the only form of submodule currently
	// supported.
	Submodules []SourceModule
	// Columns identifies any columns defined in this module.  Observe that
	// columns across modules are mapped to registers in a many-to-one fashion.
	Columns []SourceColumn
}

// Flattern modules in this tree either including (or excluding) virtual
// modules.
func (p *SourceModule) Flattern(predicate func(*SourceModule) bool) []SourceModule {
	var modules []SourceModule

	if predicate(p) {
		modules = append(modules, *p)
		for _, child := range p.Submodules {
			modules = append(modules, child.Flattern(predicate)...)
		}
	}

	return modules
}

// SourceColumn represents a source-level column which is mapped to a given HIR
// register.  Observe that multiplie source-level columns can be mapped to the
// same register.
type SourceColumn struct {
	Name string
	// Length Multiplier of source-level column.
	Multiplier uint
	// Underlying DataType of the source-level column.
	DataType sc.Type
	// Provability requirement for source-level column.
	MustProve bool
	// Determines whether this is a Computed column.
	Computed bool
	// Register at HIR level to which this column is mapped.
	Register uint
}

func init() {
	gob.Register(binfile.Attribute(&SourceMap{}))
}
