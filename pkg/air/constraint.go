package air

import (
	"github.com/consensys/go-corset/pkg/table"
)

// For now, all constraints are vanishing constraints.
type Constraint = *table.VanishingConstraint[Expr]
