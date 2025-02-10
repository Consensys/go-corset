package constraint

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// PermutationFailure provides structural information about a failing permutation constraint.
type PermutationFailure struct {
	Msg string
}

// Message provides a suitable error message
func (p *PermutationFailure) Message() string {
	return p.Msg
}

func (p *PermutationFailure) String() string {
	return p.Msg
}

// PermutationConstraint declares a constraint that one (or more) columns are a permutation
// of another.
type PermutationConstraint struct {
	Handle string
	// Targets returns the indices of the columns composing the "left" table of the
	// permutation.
	Targets []uint
	// Sources returns the indices of the columns composing the "right" table of the
	// permutation.
	Sources []uint
}

// NewPermutationConstraint creates a new permutation
func NewPermutationConstraint(handle string, targets []uint, sources []uint) *PermutationConstraint {
	if len(targets) != len(sources) {
		panic("differeng number of target / source permutation columns")
	}

	return &PermutationConstraint{handle, targets, sources}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p *PermutationConstraint) Name() string {
	return p.Handle
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *PermutationConstraint) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Accepts checks whether a permutation holds between the source and
// target columns.
func (p *PermutationConstraint) Accepts(tr trace.Trace) (sc.Coverage, sc.Failure) {
	// Coverage currently always empty for permutation constraints.
	var coverage sc.Coverage
	// Slice out data
	src := sliceColumns(p.Sources, tr)
	dst := sliceColumns(p.Targets, tr)
	// Sanity check whether column exists
	if util.ArePermutationOf(dst, src) {
		// Success
		return coverage, nil
	}
	// Prepare suitable error message
	src_names := trace.QualifiedColumnNamesToCommaSeparatedString(p.Sources, tr)
	dst_names := trace.QualifiedColumnNamesToCommaSeparatedString(p.Targets, tr)
	//
	msg := fmt.Sprintf("Target columns (%s) not permutation of source columns (%s)",
		dst_names, src_names)
	// Done
	return coverage, &PermutationFailure{msg}
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *PermutationConstraint) Lisp(schema sc.Schema) sexp.SExp {
	targets := sexp.EmptyList()
	sources := sexp.EmptyList()

	for _, tid := range p.Targets {
		target := schema.Columns().Nth(tid)
		targets.Append(sexp.NewSymbol(target.QualifiedName(schema)))
	}

	for _, sid := range p.Sources {
		source := schema.Columns().Nth(sid)
		sources.Append(sexp.NewSymbol(source.QualifiedName(schema)))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("permutation"),
		targets,
		sources,
	})
}

func sliceColumns(columns []uint, tr trace.Trace) []field.FrArray {
	// Allocate return array
	cols := make([]field.FrArray, len(columns))
	// Slice out the data
	for i, n := range columns {
		nth := tr.Column(n)
		// Copy over
		cols[i] = nth.Data()
	}
	// Done
	return cols
}
