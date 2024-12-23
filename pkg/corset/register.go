package corset

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// RegisterAllocation is a generic interface to support different "regsiter
// allocation" algorithms.  More specifically, register allocation is the
// process of allocating columns to their underlying HIR columns (a.k.a
// registers).  This is straightforward when there is a 1-1 mapping from a
// Corset column to an HIR column.  However, this is not always the case.  For
// example, array columns at the Corset level map to multiple columns at the HIR
// level.  Likewise, perspectives allow columns to be reused, meaning that
// multiple columns at the Corset level can be mapped down to just a single
// column at the HIR level.
//
// Notes:
//
// * Arrays.  These are allocated consecutive columns, as determined by their
// "width".  That is, the size of the array.
//
// * Perspectives.  This is where the main challenge lies.  Columns in different
// perspectives can be merged together, but this is only done when they have
// compatible underlying types.
type RegisterAllocation interface {
	Merge(string, string)
}

// Register encapsulates information about a "register" in the underlying
// constraint system.  The rough analogy is that "register allocation" is
// applied to map Corset columns down to HIR columns (a.k.a. registers).  The
// distinction between columns at the Corset level, and registers at the HIR
// level is necessary for two reasons: firstly, one corset column can expand to
// several HIR registers; secondly, register allocation is applied to columns in
// different perspectives of the same module.
type Register struct {
	// Context (i.e. module + multiplier) of this register.
	Context tr.Context
	// Name of this register
	Name string
	// Underlying datatype of this register.
	DataType sc.Type
	// Source columns of this register
	Sources []*ColumnBinding
}
