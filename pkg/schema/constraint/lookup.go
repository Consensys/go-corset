package constraint

import (
	"encoding/binary"
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// LookupFailure provides structural information about a failing lookup constraint.
type LookupFailure struct {
	Msg string
}

// Message provides a suitable error message
func (p *LookupFailure) Message() string {
	return p.Msg
}

func (p *LookupFailure) String() string {
	return p.Msg
}

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
	// Handle returns the handle for this lookup constraint which is simply an
	// identifier useful when debugging (i.e. to know which lookup failed, etc).
	Handle string
	// Context in which all source columns are evaluated.
	SourceContext trace.Context
	// Context in which all target columns are evaluated.
	TargetContext trace.Context
	// Sources returns the source expressions which are used to lookup into the
	// target expressions.
	Sources []E
	// Targets returns the target expressions which are used to lookup into the
	// target expressions.
	Targets []E
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint[E schema.Evaluable](handle string, source trace.Context,
	target trace.Context, sources []E, targets []E) *LookupConstraint[E] {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}

	return &LookupConstraint[E]{handle, source, target, sources, targets}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p *LookupConstraint[E]) Name() string {
	return p.Handle
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p *LookupConstraint[E]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	//
	if module == p.SourceContext.Module() {
		for _, e := range p.Sources {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	} else if module == p.TargetContext.Module() {
		for _, e := range p.Targets {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	}
	//
	return bound
}

// Accepts checks whether a lookup constraint into the target columns holds for
// all rows of the source columns.
//
//nolint:revive
func (p *LookupConstraint[E]) Accepts(tr trace.Trace) (sc.Coverage, schema.Failure) {
	var coverage sc.Coverage
	// Determine height of enclosing module for source columns
	src_height := tr.Height(p.SourceContext)
	tgt_height := tr.Height(p.TargetContext)
	//
	rows := hash.NewSet[hash.BytesKey](tgt_height)
	// Add all target columns to the set
	for i := 0; i < int(tgt_height); i++ {
		ith_bytes := evalExprsAt(i, p.Targets, tr)
		rows.Insert(hash.NewBytesKey(ith_bytes))
	}
	// Check all source columns are contained
	for i := 0; i < int(src_height); i++ {
		ith_bytes := evalExprsAt(i, p.Sources, tr)
		// Check whether contained.
		if !rows.Contains(hash.NewBytesKey(ith_bytes)) {
			return coverage, &LookupFailure{fmt.Sprintf("lookup \"%s\" failed (row %d)", p.Handle, i)}
		}
	}
	//
	return coverage, nil
}

func evalExprsAt[E schema.Evaluable](k int, sources []E, tr trace.Trace) []byte {
	// Each fr.Element is 4 x 64bit words.
	bytes := make([]byte, 32*len(sources))
	// Slice provides an access window for writing
	slice := bytes
	// Evaluate each expression in turn
	for i := 0; i < len(sources); i++ {
		ith := sources[i].EvalAt(k, tr)
		// Copy over each element
		binary.BigEndian.PutUint64(slice, ith[0])
		binary.BigEndian.PutUint64(slice[8:], ith[1])
		binary.BigEndian.PutUint64(slice[16:], ith[2])
		binary.BigEndian.PutUint64(slice[24:], ith[3])
		// Move slice over
		slice = slice[32:]
	}
	// Done
	return bytes
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p *LookupConstraint[E]) Lisp(schema sc.Schema) sexp.SExp {
	sources := sexp.EmptyList()
	targets := sexp.EmptyList()
	// Iterate source expressions
	for i := 0; i < len(p.Sources); i++ {
		sources.Append(p.Sources[i].Lisp(schema))
	}
	// Iterate source expressions
	for i := 0; i < len(p.Targets); i++ {
		targets.Append(p.Targets[i].Lisp(schema))
	}
	// Done
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("lookup"),
		sexp.NewSymbol(p.Handle),
		targets,
		sources,
	})
}
