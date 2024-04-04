package binfile

import (
	//"encoding/json"
	// "fmt"
	// "github.com/Consensys/go-corset/pkg/constraint"
)

// An enumeration of constraint forms.  Exactly one of these fields
// must be non-nil to signify its form.
type JsonConstraint struct {
	Vanishes *JsonVanishingConstraint
}

// Corresponds to a constraint whose expression must evaluate to zero
// for every row of the table.
type JsonVanishingConstraint struct {
	Handle string `json:"handle"`
	Domain JsonDomain `json:"domain"`
	Expr JsonTypedExpr `json:"expr"`
}
