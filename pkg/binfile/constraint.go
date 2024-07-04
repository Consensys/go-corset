package binfile

import (
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

func (e jsonConstraint) addToSchema(schema *hir.Schema) {
	// NOTE: for permutation constraints, we currently ignore them as they
	// actually provide no useful information.  They are generated from
	// "defpermutation" declarations, but lack information about the direction
	// of sorting (signs).  Instead, we have to extract what we need from
	// "Sorted" computations.
	if e.Vanishes != nil {
		// Translate the vanishing expression
		expr := e.Vanishes.Expr.ToHir(schema)
		// Translate Domain
		domain := e.Vanishes.Domain.toHir()
		// Determine enclosing module
		module, multiplier := sc.DetermineEnclosingModuleOfExpression(expr, schema)
		// Construct the vanishing constraint
		schema.AddVanishingConstraint(e.Vanishes.Handle, module, multiplier, domain, expr)
	} else if e.Lookup != nil {
		sources := jsonExprsToHirUnit(e.Lookup.From, schema)
		targets := jsonExprsToHirUnit(e.Lookup.To, schema)
		source, source_multiplier, err1 := sc.DetermineEnclosingModuleOfExpressions(sources, schema)
		target, target_multiplier, err2 := sc.DetermineEnclosingModuleOfExpressions(targets, schema)
		// Error check
		if err1 != nil {
			panic(err1.Error)
		} else if err2 != nil {
			panic(err2.Error)
		}
		// Add constraint
		schema.AddLookupConstraint(e.Lookup.Handle, source, source_multiplier, target, target_multiplier, sources, targets)
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
