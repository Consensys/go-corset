package binfile

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
)

// JsonConstraint аn enumeration of constraint forms.  Exactly one of these fields
// must be non-nil to signify its form.
type jsonConstraint struct {
	Vanishes    *jsonVanishingConstraint
	Permutation *jsonPermutationConstraint
	Lookup      *jsonLookupConstraint
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

type jsonLookupConstraint struct {
	Handle string          `json:"handle"`
	From   []jsonTypedExpr `json:"included"`
	To     []jsonTypedExpr `json:"including"`
}

// =============================================================================
// Translation
// =============================================================================

func (e jsonConstraint) addToSchema(colmap map[uint]uint, schema *hir.Schema) {
	// NOTE: for permutation constraints, we currently ignore them as they
	// actually provide no useful information.  They are generated from
	// "defpermutation" declarations, but lack information about the direction
	// of sorting (signs).  Instead, we have to extract what we need from
	// "Sorted" computations.
	if e.Vanishes != nil {
		// Translate the vanishing expression
		expr := e.Vanishes.Expr.ToHir(colmap, schema)
		// Translate Domain
		domain := e.Vanishes.Domain.toHir()
		// Determine enclosing module
		ctx := expr.Context(schema)
		// Construct the vanishing constraint
		schema.AddVanishingConstraint(e.Vanishes.Handle, ctx, domain, expr)
	} else if e.Lookup != nil {
		sources := jsonExprsToHirUnit(e.Lookup.From, colmap, schema)
		targets := jsonExprsToHirUnit(e.Lookup.To, colmap, schema)
		sourceCtx := sc.JoinContexts(sources, schema)
		targetCtx := sc.JoinContexts(targets, schema)
		// Error check
		if sourceCtx.IsConflicted() || sourceCtx.IsVoid() {
			panic(fmt.Sprintf("lookup %s has conflicting source evaluation context", e.Lookup.Handle))
		} else if targetCtx.IsConflicted() || targetCtx.IsVoid() {
			panic(fmt.Sprintf("lookup %s has conflicting target evaluation context", e.Lookup.Handle))
		}
		// Add constraint
		schema.AddLookupConstraint(e.Lookup.Handle, sourceCtx, targetCtx, sources, targets)
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
