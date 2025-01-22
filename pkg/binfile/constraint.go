package binfile

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// JsonConstraint аn enumeration of constraint forms.  Exactly one of these fields
// must be non-nil to signify its form.
type jsonConstraint struct {
	Vanishes    *jsonVanishingConstraint
	Permutation *jsonPermutationConstraint
	Lookup      *jsonLookupConstraint
	InRange     *jsonRangeConstraint
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

type jsonRangeConstraint struct {
	Handle string        `json:"handle"`
	Expr   jsonTypedExpr `json:"exp"`
	Max    jsonExprConst `json:"max"`
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
		// Normalise handle
		handle := asHandle(e.Vanishes.Handle)
		// Construct the vanishing constraint
		schema.AddVanishingConstraint(handle.column, ctx, domain, expr)
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
		// Normalise handle
		handle := asHandle(e.Lookup.Handle)
		// Add constraint
		schema.AddLookupConstraint(handle.column, sourceCtx, targetCtx, sources, targets)
	} else if e.InRange != nil {
		// Translate the vanishing expression
		expr := e.InRange.Expr.ToHir(colmap, schema)
		// Determine enclosing module
		ctx := expr.Context(schema)
		// Convert bound into max
		bound := e.InRange.Max.ToField()
		handle := expr.Lisp(schema).String(true)
		// Construct the vanishing constraint
		schema.AddRangeConstraint(handle, ctx, expr, bound)
	} else if e.Permutation == nil {
		// Catch all
		panic("Unknown JSON constraint encountered")
	}
}

func (e jsonDomain) toHir() util.Option[int] {
	if len(e.Set) == 1 {
		domain := e.Set[0]
		return util.Some(domain)
	} else if e.Set != nil {
		panic("Unknown domain")
	}
	// Default
	return util.None[int]()
}
