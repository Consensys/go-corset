package schema

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Schema represents a schema which can be used to manipulate a trace.
type Schema interface {
	// Assignments returns an iterator over the assignments of this schema.  That
	// is, the subset of declarations whose trace values can be computed from
	// the inputs.
	Assignments() util.Iterator[Assignment]

	// Columns returns an iterator over the underlying columns of this schema.
	// Specifically, the index of a column in this array is its column index.
	Columns() util.Iterator[Column]

	// Constraints returns an iterator over the underlying constraints of this
	// schema.
	Constraints() util.Iterator[Constraint]

	// Declarations returns an iterator over the column declarations of this
	// schema.
	Declarations() util.Iterator[Declaration]

	// Modules returns an iterator over the declared set of modules within this
	// schema.
	Modules() util.Iterator[Module]
}

// Declaration represents an element which declares one (or more) columns in the
// schema.  For example, a single data column is (for now) always a column group
// of size 1. Likewise, an iterator of size n is a column group of size n, etc.
type Declaration interface {
	// Return the declared columns (in the order of declaration).
	Columns() util.Iterator[Column]

	// Determines whether or not this declaration is computed.
	IsComputed() bool
}

// Assignment represents a schema element which declares one or more columns
// whose values are "assigned" from the results of a computation.  An assignment
// is a column group which, additionally, can provide information about the
// computation (e.g. which columns it depends upon, etc).
type Assignment interface {
	Declaration

	// ExpandTrace expands a given trace to include "computed
	// columns".  These are columns which do not exist in the
	// original trace, but are added during trace expansion to
	// form the final trace.
	ExpandTrace(tr.Trace) error
	// RequiredSpillage returns the minimum amount of spillage required to ensure
	// valid traces are accepted in the presence of arbitrary padding.  Note,
	// spillage is currently assumed to be required only at the front of a
	// trace.
	RequiredSpillage() uint
}

// Constraint represents an element which can "accept" a trace, or either reject
// with an error (or eventually perhaps report a warning).
type Constraint interface {
	Accepts(tr.Trace) error
}

// Evaluable captures something which can be evaluated on a given table row to
// produce an evaluation point.  For example, expressions in the
// Mid-Level or Arithmetic-Level IR can all be evaluated at rows of a
// table.
type Evaluable interface {
	util.Boundable
	Contextual
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAt(int, tr.Trace) *fr.Element
}

// Testable captures the notion of a constraint which can be tested on a given
// row of a given trace.  It is very similar to Evaluable, except that it only
// indicates success or failure.  The reason for using this interface over
// Evaluable is that, for historical reasons, constraints at the HIR cannot be
// Evaluable (i.e. because they return multiple values, rather than a single
// value).  However, constraints at the HIR level remain testable.
type Testable interface {
	util.Boundable
	Contextual

	// TestAt evaluates this expression in a given tabular context and checks it
	// against zero. Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be undefined for
	// several reasons: firstly, if it accesses a row which does not exist (e.g.
	// at index -1); secondly, if it accesses a column which does not exist.
	TestAt(int, tr.Trace) bool
}

// Contextual captures something which requires an evaluation context (i.e. a
// single enclosing module) in order to make sense.  For example, expressions
// require a single context.  This interface is separated from Evaluable (and
// Testable) because HIR expressions do not implement Evaluable.
type Contextual interface {
	// Context returns the evaluation context (i.e. enclosing module + length
	// multiplier) for this constraint.  Every expression must have a single
	// evaluation context.  This function therefore attempts to determine what
	// that is, or return false to signal an error. There are several failure
	// modes which need to be considered.  Firstly, if the expression has no
	// enclosing module (e.g. because it is a constant expression) then it will
	// return 'math.MaxUint` to signal this.  Secondly, if the expression has
	// multiple (i.e. conflicting) enclosing modules then it will return false
	// to signal this.  Likewise, the expression could have a single enclosing
	// module but multiple conflicting length multipliers, in which case it also
	// returns false.
	Context(Schema) tr.Context
}

// ============================================================================
// Column
// ============================================================================

// Column represents a specific column in the schema that, ultimately, will
// correspond 1:1 with a column in the trace.
type Column struct {
	// Evaluation context of this column.
	context tr.Context
	// Returns the name of this column
	name string
	// Returns the expected type of data in this column
	datatype Type
}

// NewColumn constructs a new column
func NewColumn(context tr.Context, name string, datatype Type) Column {
	return Column{context, name, datatype}
}

// Context returns the evaluation context for this column access, which is
// determined by the column itself.
func (p Column) Context() tr.Context {
	return p.context
}

// Name returns the name of this column
func (p Column) Name() string {
	return p.name
}

// Type returns the expected type of data in this column
func (p Column) Type() Type {
	return p.datatype
}

func (p Column) String() string {
	return fmt.Sprintf("%s:%s", p.name, p.datatype.String())
}

// ============================================================================
// Module
// ============================================================================

// Module represents a specific module in the schema that groups columns
// together.  Modules don't serve a huge function in a schema at this time.
// They could, however, be used to iterate over the things they contain (e.g.
// their columns, their constraints, etc).  Potentially, they might also do
// things like identify input / output columns, etc.
type Module struct {
	// Returns the name of this column
	name string
}

// NewModule constructs a new column
func NewModule(name string) Module {
	return Module{name}
}

// Name returns the name of this module
func (p *Module) Name() string {
	return p.name
}
