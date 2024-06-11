package binfile

import (
	"github.com/consensys/go-corset/pkg/hir"
)

// JsonConstraint аn enumeration of constraint forms.  Exactly one of these fields
// must be non-nil to signify its form.
type jsonConstraint struct {
	Vanishes    *jsonVanishingConstraint
	Permutation *jsonPermutationConstraint
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

type jsonPermutationConstraint struct {
	From []string `json:"from"`
	To   []string `json:"to"`
}

// =============================================================================
// Translation
// =============================================================================

func (e jsonConstraint) addToSchema(schema *hir.Schema) {
	// NOTE: for permutation constraints, we currently ignore them as they
	// actually provide no useful information.  They are generated from
	// "defpermutation" declarations, but lack information about the direction
	// of sorting (signs).  Instead, we have to extract what we need from
	// "Sorted" computations.
	if e.Vanishes != nil {
		// Translate the vanishing expression
		expr := e.Vanishes.Expr.ToHir()
		// Translate Domain
		domain := e.Vanishes.Domain.toHir()
		// Construct the vanishing constraint
		schema.AddVanishingConstraint(e.Vanishes.Handle, domain, expr)
	} else if e.Permutation == nil {
		// Catch all
		panic("Unknown JSON constraint encountered")
	}
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
