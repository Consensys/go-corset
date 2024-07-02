package binfile

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
)

// ColumnRef uniquely identifies a column in the system.
type ColumnRef struct {
	module string
	column string
}

// Resolve a columnRef given as a handle of the form "mod:col#reg",
// where mod is the module name, col is the column name and reg is the
// (optional) allocated register.
func (p *ColumnRef) resolve(schema *hir.Schema) (uint, uint) {
	// First, lookup module
	mid, ok := schema.Modules().Find(func(m sc.Module) bool {
		return m.Name() == p.module
	})
	if !ok {
		panic(fmt.Sprintf("unknown module %s encountered", p.module))
	}
	// Second, lookup column index
	cid, ok := sc.ColumnIndexOf(schema, mid, p.column)
	if !ok {
		panic(fmt.Sprintf("unknown column %s.%s encountered", p.module, p.column))
	}
	// Done
	return mid, cid
}

func asColumnRefs(crefs []jsonColumnRef) []ColumnRef {
	refs := make([]ColumnRef, len(crefs))
	for i := 0; i < len(refs); i++ {
		refs[i] = asColumnRef(crefs[i])
	}

	return refs
}

// Split a handle into its constituent parts.
func asColumnRef(handle string) ColumnRef {
	split := strings.Split(handle, ":")
	// Split off the allocated register as we don't use this.
	split_1 := strings.Split(split[1], "#")
	module := split[0]
	column := split_1[0]
	return ColumnRef{module, column}
}
