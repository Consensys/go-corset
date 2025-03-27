// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package schema

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Schema represents a schema which can be used to manipulate a trace.
type Schema interface {
	// Assertions returns an iterator over the property assertions of this
	// schema.  These are properties which should hold true for any valid trace
	// (though, of course, may not hold true for an invalid trace).
	Assertions() iter.Iterator[Constraint]

	// Assignments returns an iterator over the assignments of this schema.  That
	// is, the subset of declarations whose trace values can be computed from
	// the inputs.
	Assignments() iter.Iterator[Assignment]

	// Columns returns an iterator over the underlying columns of this schema.
	// Specifically, the index of a column in this array is its column index.
	Columns() iter.Iterator[Column]

	// Constraints returns an iterator over the underlying constraints of this
	// schema.
	Constraints() iter.Iterator[Constraint]

	// Declarations returns an iterator over the column declarations of this
	// schema.
	Declarations() iter.Iterator[Declaration]

	// Iterator over the input (i.e. non-computed) columns of the schema.
	InputColumns() iter.Iterator[Column]

	// Modules returns an iterator over the declared set of modules within this
	// schema.
	Modules() iter.Iterator[Module]
}

// Declaration represents an element which declares one (or more) columns in the
// schema.  For example, a single data column is (for now) always a column group
// of size 1. Likewise, an iterator of size n is a column group of size n, etc.
type Declaration interface {
	Lispifiable
	// Return the declared columns (in the order of declaration).
	Columns() iter.Iterator[Column]

	// Context returns the evaluation context (i.e. enclosing module + length
	// multiplier) for this declaration.  Every declaration must have a single,
	// unique context.
	Context() tr.Context

	// Determines whether or not this declaration is computed.
	IsComputed() bool
}

// Assignment represents a schema element which declares one or more columns
// whose values are "assigned" from the results of a computation.  An assignment
// is a column group which, additionally, can provide information about the
// computation (e.g. which columns it depends upon, etc).
type Assignment interface {
	Declaration
	util.Boundable

	// ComputeColumns computes the values of columns defined by this assignment.
	// In order for this computation to makes sense, all columns on which this
	// assignment depends must exist (e.g. are either inputs or have been
	// computed already).  Computed columns do not exist in the original trace,
	// but are added during trace expansion to form the final trace.
	ComputeColumns(tr.Trace) ([]trace.ArrayColumn, error)

	// Returns the set of columns that this assignment depends upon.  That can
	// include both input columns, as well as other computed columns.
	Dependencies() []uint

	// CheckConsistent performs some simple checks of consistency against the
	// given schema.
	CheckConsistency(schema Schema) error
}

// Constraint represents an element which can "accept" a trace, or either reject
// with an error (or eventually perhaps report a warning).
type Constraint interface {
	Lispifiable
	// Accepts determines whether a given constraint accepts a given trace or
	// not.  If not, a failure is produced.  Otherwise, a bitset indicating
	// branch coverage is returned.
	Accepts(tr.Trace) (bit.Set, Failure)
	// Contexts returns the evaluation contexts (i.e. enclosing module + length
	// multiplier) for this constraint.  Most constraints have only a single
	// evaluation context, though some (e.g. lookups) have more.  Note that all
	// constraints have at least one context (which we can call the "primary"
	// context).
	Contexts() []tr.Context
	// Determine the well-definedness bounds for this constraint in both the
	// negative (left) or positive (right) directions.  For example, consider an
	// expression such as "(shift X -1)".  This is technically undefined for the
	// first row of any trace and, by association, any constraint evaluating
	// this expression on that first row is also undefined (and hence must pass)
	Bounds(module uint) util.Bounds
	// Return the total number of logical branches this constraint can take
	// during evaluation.
	Branches() uint
	// Name returns a unique name and case number for a given constraint.  This
	// is useful purely for identifying constraints in reports, etc.  The case
	// number is used to differentiate different low-level constraints which are
	// extracted from the same high-level constraint.
	Name() (string, uint)
}

// Failure embodies structured information about a failing constraint.
// This includes the constraint itself, along with the row
type Failure interface {
	// Provides a suitable error message
	Message() string
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
	EvalAt(int, tr.Trace) (fr.Element, error)
	// Branches returns the number of unique evaluation paths through the given
	// constraint.
	Branches() uint
	// RequiredCells returns the set of trace cells on which evaluation of this
	// constraint element depends.
	RequiredCells(int, tr.Trace) *set.AnySortedSet[tr.CellRef]
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
	TestAt(int, tr.Trace) (bool, uint, error)
	// Branches returns the number of unique evaluation paths through the given
	// constraint.
	Branches() uint
}

// Contextual captures something which requires an evaluation context (i.e. a
// single enclosing module) in order to make sense.  For example, expressions
// require a single context.  This interface is separated from Evaluable (and
// Testable) because HIR expressions do not implement Evaluable.
type Contextual interface {
	Lispifiable
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

	// RequiredColumns returns the set of columns on which this term depends.
	// That is, columns whose values may be accessed when evaluating this term
	// on a given trace.
	RequiredColumns() *set.SortedSet[uint]
	// RequiredCells returns the set of trace cells on which evaluation of this
	// constraint element depends.
	RequiredCells(int, tr.Trace) *set.AnySortedSet[tr.CellRef]
}

// Lispifiable captures a schema element which can be turned into a stand alone
// S-expression (e.g. for printing).
type Lispifiable interface {
	// Lisp converts this schema element into a simple S-Expression, for example
	// so it can be printed.
	Lisp(sc Schema) sexp.SExp
}

// ============================================================================
// Column
// ============================================================================

// Column represents a specific column in the schema that, ultimately, will
// correspond 1:1 with a column in the trace.
type Column struct {
	// Evaluation Context of this column.
	Context tr.Context
	// Returns the Name of this column
	Name string
	// Returns the expected type of data in this column
	DataType Type
}

// NewColumn constructs a new column
func NewColumn(context tr.Context, name string, datatype Type) Column {
	return Column{context, name, datatype}
}

// QualifiedName returns the fully qualified name of this column
func (p Column) QualifiedName(schema Schema) string {
	mod := schema.Modules().Nth(p.Context.Module())
	if mod.Name != "" {
		return fmt.Sprintf("%s:%s", mod.Name, p.Name)
	}
	//
	return p.Name
}

func (p Column) String() string {
	return fmt.Sprintf("%s:%s", p.Name, p.DataType.String())
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
	Name string
}

// NewModule constructs a new column
func NewModule(name string) Module {
	return Module{name}
}

// ============================================================================
// InternalFailure
// ============================================================================

// InternalFailure is a generic mechanism for reporting failures, particularly
// as arising from evaluation of a given expression.
type InternalFailure struct {
	// Handle of the failing constraint
	Handle string
	// Row on which the constraint failed
	Row uint
	// Cells involved (if any)
	Term Contextual
	// Error message
	Error string
}

// Message provides a suitable error message
func (p *InternalFailure) Message() string {
	return p.Error
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *InternalFailure) RequiredCells(trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	if p.Term != nil {
		return p.Term.RequiredCells(int(p.Row), trace)
	}
	// Empty set
	return set.NewAnySortedSet[tr.CellRef]()
}
