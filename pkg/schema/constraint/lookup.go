package constraint

import (
	"encoding/binary"
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// LookupFailure provides structural information about a failing lookup constraint.
type LookupFailure struct {
	msg string
}

// Message provides a suitable error message
func (p *LookupFailure) Message() string {
	return p.msg
}

func (p *LookupFailure) String() string {
	return p.msg
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
	handle string
	// Evaluation context for source columns.
	source trace.Context
	// Evaluation context for target columns.
	target trace.Context
	// Source rows represent the subset of rows.
	sources []E
	// Target rows represent the set of rows.
	targets []E
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint[E schema.Evaluable](handle string, source trace.Context,
	target trace.Context, sources []E, targets []E) *LookupConstraint[E] {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}

	return &LookupConstraint[E]{handle, source, target, sources, targets}
}

// Handle returns the handle for this lookup constraint which is simply an
// identifier useful when debugging (i.e. to know which lookup failed, etc).
//
//nolint:revive
func (p *LookupConstraint[E]) Handle() string {
	return p.handle
}

// SourceContext returns the contezt in which all target expressions are evaluated.
func (p *LookupConstraint[E]) SourceContext() trace.Context {
	return p.source
}

// TargetContext returns the contezt in which all target expressions are evaluated.
func (p *LookupConstraint[E]) TargetContext() trace.Context {
	return p.target
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
func (p *LookupConstraint[E]) Accepts(tr trace.Trace) schema.Failure {
	// Determine height of enclosing module for source columns
	src_height := tr.Height(p.source)
	tgt_height := tr.Height(p.target)
	//
	rows := util.NewHashSet[util.BytesKey](tgt_height)
	// Add all target columns to the set
	for i := 0; i < int(tgt_height); i++ {
		ith_bytes := evalExprsAt(i, p.targets, tr)
		rows.Insert(util.NewBytesKey(ith_bytes))
	}
	// Check all source columns are contained
	for i := 0; i < int(src_height); i++ {
		ith_bytes := evalExprsAt(i, p.sources, tr)
		// Check whether contained.
		if !rows.Contains(util.NewBytesKey(ith_bytes)) {
			return &LookupFailure{fmt.Sprintf("lookup \"%s\" failed (row %d)", p.handle, i)}
		}
	}
	//
	return nil
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
	for i := 0; i < len(p.sources); i++ {
		sources.Append(p.sources[i].Lisp(schema))
	}
	// Iterate source expressions
	for i := 0; i < len(p.targets); i++ {
		targets.Append(p.targets[i].Lisp(schema))
	}
	// Done
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("deflookup"),
		sexp.NewSymbol(p.handle),
		targets,
		sources,
	})
}
