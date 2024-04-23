package hir

import (
	"github.com/consensys/go-corset/pkg/trace"
)

// For now, all constraints are vanishing constraints.
type Constraint = *VanishingConstraint

type VanishingConstraint = trace.VanishingConstraint[Expr]
