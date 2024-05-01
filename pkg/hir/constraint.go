package hir

import (
	"github.com/consensys/go-corset/pkg/table"
)

// Constraint for now, all constraints are vanishing constraints.
type Constraint = *VanishingConstraint

// VanishingConstraint represents a constraint which should, on every row of the
// table, evaluate to zero.  The only exception is when the constraint is
// undefined (e.g. because it references a non-existent table cell).  In such
// case, the constraint is ignored.
type VanishingConstraint = table.VanishingConstraint[Expr]
