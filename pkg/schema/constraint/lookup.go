package constraint

import (
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
)

// LookupConstraint (sometimes also called an inclusion constraint) constrains
// two sets of columns (potentially in different modules). Specifically, every
// row in the source columns must match a row in the target columns (but not
// vice-versa).  As such, the number of source columns must be the same as the
// number of target columns.  Furthermore, every source column must be in the
// same module, and likewise for target modules.  However, the source columns
// can be in a different module from the target columns.
//
// Lookup constraints are typically used to "connect" modules together.  We can
// think of them (in some ways) as being a little like function calls.  In this
// analogy, the source module is making a "function call" into the target
// module.  That is, the target module contains the set of valid input/output
// pairs (and perhaps other constraints to ensure the required relationship) and
// the source module is just checking that a given set of input/output pairs
// makes sense.
type LookupConstraint[E schema.Evaluable] struct {
	handle string
	// Source rows represent the subset of rows.
	sources []E
	// Target rows represent the set of rows.
	targets []E
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint[E schema.Evaluable](handle string, sources []E, targets []E) *LookupConstraint[E] {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}

	return &LookupConstraint[E]{handle, sources, targets}
}

// Handle returns the handle for this lookup constraint which is simply an
// identifier useful when debugging (i.e. to know which lookup failed, etc).
//
//nolint:revive
func (p *LookupConstraint[E]) Handle() string {
	return p.handle
}

// Sources returns the source expressions which are used to lookup into the
// target expressions.
func (p *LookupConstraint[E]) Sources() []E {
	return p.sources
}

// Targets returns the target expressions which are used to lookup into the
// target expressions.
func (p *LookupConstraint[E]) Targets() []E {
	return p.targets
}

// Accepts checks whether a lookup constraint into the target columns holds for
// all rows of the source columns.
func (p *LookupConstraint[E]) Accepts(tr trace.Trace) error {
	panic("TODO")
}
