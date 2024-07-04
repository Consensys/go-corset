package constraint

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
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
	// Enclosing module for source columns.
	source uint
	// Length multiplier partly determines the evaluation context for source
	// expressions.
	source_multiplier uint
	// Enclosing module for target columns.
	target uint
	// Length multiplier partly determines the evaluation context for target
	// expressions.
	target_multiplier uint
	// Source rows represent the subset of rows.
	sources []E
	// Target rows represent the set of rows.
	targets []E
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint[E schema.Evaluable](handle string, source uint, source_multiplier uint,
	target uint, target_multiplier uint, sources []E, targets []E) *LookupConstraint[E] {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}

	return &LookupConstraint[E]{handle, source, source_multiplier, target, target_multiplier, sources, targets}
}

// Handle returns the handle for this lookup constraint which is simply an
// identifier useful when debugging (i.e. to know which lookup failed, etc).
//
//nolint:revive
func (p *LookupConstraint[E]) Handle() string {
	return p.handle
}

// SourceContext returns the module in which all source columns are located.
func (p *LookupConstraint[E]) SourceContext() (uint, uint) {
	return p.source, p.source_multiplier
}

// TargetContext returns the module in which all target columns are located.
func (p *LookupConstraint[E]) TargetContext() (uint, uint) {
	return p.target, p.target_multiplier
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
//
//nolint:revive
func (p *LookupConstraint[E]) Accepts(tr trace.Trace) error {
	// Determine height of enclosing module for source columns
	src_height := tr.Modules().Get(p.source).Height()
	tgt_height := tr.Modules().Get(p.target).Height()
	// Go through every row of the source columns checking they are present in
	// the target columns.
	//
	// NOTE: performance could be improved here by pre-evaluating and sorting
	// the target columns to give O(log n) lookups, or using hash map to give
	// O(1) checks.
	for i := 0; i < int(src_height); i++ {
		ith := evalExprsAt(i, p.sources, tr)
		matched := false

		for j := 0; j < int(tgt_height); j++ {
			jth := evalExprsAt(j, p.targets, tr)
			if util.Equals(ith, jth) {
				matched = true
				break
			}
		}

		if !matched {
			return fmt.Errorf("lookup \"%s\" failed (row %d, %v)", p.handle, i, ith)
		}
	}
	//
	return nil
}

func evalExprsAt[E schema.Evaluable](k int, sources []E, tr trace.Trace) []*fr.Element {
	vals := make([]*fr.Element, len(sources))
	// Evaluate each expression in turn
	for i := 0; i < len(sources); i++ {
		vals[i] = sources[i].EvalAt(k, tr)
	}
	// Done
	return vals
}

//nolint:revive
func (p *LookupConstraint[E]) String() string {
	sources := ""
	targets := ""
	// Iterate source expressions
	for i := 0; i < len(p.sources); i++ {
		if i == 0 {
			sources = fmt.Sprintf("%s", any(p.sources[i]))
		} else {
			sources = fmt.Sprintf("%s %s", sources, any(p.sources[i]))
		}
	}
	// Iterate source expressions
	for i := 0; i < len(p.targets); i++ {
		if i == 0 {
			targets = fmt.Sprintf("%s", any(p.targets[i]))
		} else {
			targets = fmt.Sprintf("%s %s", targets, any(p.targets[i]))
		}
	}
	// Done
	return fmt.Sprintf("(lookup %s (%s) (%s))", p.handle, targets, sources)
}
