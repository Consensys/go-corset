package binfile

import (
	"github.com/consensys/go-corset/pkg/hir"
)

// JsonConstraint аn enumeration of constraint forms.  Exactly one of these fields
// must be non-nil to signify its form.
type jsonConstraint struct {
	Vanishes *jsonVanishingConstraint
}

type jsonDomain struct {
	Set []int
}

// JsonVanishingConstraint corresponds to a constraint whose expression must evaluate to zero
// for every row of the table.
type jsonVanishingConstraint struct {
	Handle string        `json:"handle"`
	Domain jsonDomain    `json:"domain"`
	Expr   jsonTypedExpr `json:"expr"`
}

// =============================================================================
// Translation
// =============================================================================

func (e jsonConstraint) addToSchema(schema *hir.Schema) {
	if e.Vanishes == nil {
		panic("Unknown JSON constraint encountered")
	}

	// Translate the vanishing expression
	expr := e.Vanishes.Expr.ToHir()
	// Translate Domain
	domain := e.Vanishes.Domain.toHir()
	// Construct the vanishing constraint
	schema.AddVanishingConstraint(e.Vanishes.Handle, domain, expr)
}

func (e jsonDomain) toHir() *int {
	if len(e.Set) == 1 {
		domain := e.Set[0]
		return &domain
	} else if e.Set != nil {
		panic("Unknown domain")
	}
	// Default
	return nil
}
