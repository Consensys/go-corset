package binfile

import "github.com/consensys/go-corset/pkg/hir"

// JsonConstraint аn enumeration of constraint forms.  Exactly one of these fields
// must be non-nil to signify its form.
type jsonConstraint struct {
	Vanishes *jsonVanishingConstraint
}

// JsonVanishingConstraint corresponds to a constraint whose expression must evaluate to zero
// for every row of the table.
type jsonVanishingConstraint struct {
	Handle string        `json:"handle"`
	Domain string        `json:"domain"`
	Expr   jsonTypedExpr `json:"expr"`
}

// =============================================================================
// Translation
// =============================================================================

func (e jsonConstraint) ToHir() hir.Constraint {
	if e.Vanishes != nil {
		// Translate the vanishing expression
		expr := e.Vanishes.Expr.ToHir()
		// Construct the vanishing constraint
		return &hir.VanishingConstraint{Handle: e.Vanishes.Handle, Expr: expr}
	}

	panic("Unknown JSON constraint encountered")
}
