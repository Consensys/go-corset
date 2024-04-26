package hir

import (
	"github.com/consensys/go-corset/pkg/table"
)

// For now, all constraints are vanishing constraints.
type Constraint = *VanishingConstraint

type VanishingConstraint = table.VanishingConstraint[Expr]
