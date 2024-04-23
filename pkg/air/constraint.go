package air

import (
	"github.com/consensys/go-corset/pkg/trace"
)

// For now, all constraints are vanishing constraints.
type Constraint = *trace.VanishingConstraint[Expr]
