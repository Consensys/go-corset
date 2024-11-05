package trace

import (
	"fmt"
	"math"
)

// Context represents the evaluation context in which an expression can be
// evaluated.  Firstly, every expression must have a single enclosing module
// (i.e. all columns accessed by the expression are in that module); secondly,
// the length multiplier for all columns accessed by the expression must be the
// same.  Constant expressions are something of an anomily here since they have
// neither an enclosing module, nor a length modifier.  Instead, we consider
// that constant expressions are evaluated in the empty --- or void --- context.
// This is something like a bottom type which is contained within all other
// contexts.
//
// NOTE: Whilst the evaluation context provides a general abstraction, there are
// a number of restrictions imposed on it in practice which limit its use. These
// restrictions arise from what is and is not supported by the underlying
// constraint system (i.e. the prover).  Example restrictions imposed include:
// multipliers must be powers of 2.  Likewise, non-normal expressions (i.e those
// with a multipler > 1) can only be used in a fairly limited number of
// situtions (e.g. lookups).
type Context struct {
	// Identifies the module in which this evaluation context exists.  The empty
	// module is given by the maximum index (math.MaxUint).
	module uint
	// Identifies the length multiplier required to complete this context.  In
	// essence, length multiplies divide up a given module into several disjoint
	// "subregions", such than every expression exists only in one of them.
	multiplier uint
}

// VoidContext returns the void (or empty) context.  This is the bottom type in
// the lattice, and is the context contained in all other contexts.  It is
// needed, for example, as the context for constant expressions.
func VoidContext() Context {
	return Context{math.MaxUint, 0}
}

// ConflictingContext represents the case where multiple different contexts have
// been joined together.  For example, when determining the context of an
// expression, the conflicting context is returned when no single context can be
// deteremed.  This value is generally considered to indicate an error.
func ConflictingContext() Context {
	return Context{math.MaxUint - 1, 0}
}

// NewContext returns a context representing the given module with the given
// length multiplier.
func NewContext(module uint, multiplier uint) Context {
	return Context{module, multiplier}
}

// Module returns the module for this context.  Note, however, that this is
// nonsensical in the case of either the void or the conflicted  context.  In
// this cases, this method will panic.
func (p Context) Module() uint {
	if !p.IsVoid() && !p.IsConflicted() {
		return p.module
	} else if p.IsVoid() {
		panic("void context has no module")
	}

	panic("conflicted context")
}

// LengthMultiplier returns the length multiplier for this context.  Note,
// however, that this is nonsensical in the case of either the void or the
// conflicted  context.  In this cases, this method will panic.
func (p Context) LengthMultiplier() uint {
	if !p.IsVoid() && !p.IsConflicted() {
		return p.multiplier
	} else if p.IsVoid() {
		panic("void context has no module")
	}

	panic("conflicted context has no module")
}

// IsVoid checks whether this context is the void context (or not).  This is the
// bottom element in the lattice.
func (p Context) IsVoid() bool {
	return p.module == math.MaxUint
}

// IsConflicted checks whether this context represents the conflicted context.
// This is the top element in the lattice, and is used to represent the case
// where e.g. an expression has multiple conflicting contexts.
func (p Context) IsConflicted() bool {
	return p.module == math.MaxUint-1
}

// Multiply updates the length multiplier by multiplying it by a given factor,
// producing the updated context.
func (p Context) Multiply(factor uint) Context {
	return NewContext(p.module, p.multiplier*factor)
}

// Join returns the least upper bound of the two contexts, or false if this does
// not exist (i.e. the two context's are in conflict).
func (p Context) Join(other Context) Context {
	if p.IsVoid() {
		return other
	} else if other.IsVoid() {
		return p
	} else if p != other || p.IsConflicted() || other.IsConflicted() {
		// Conflicting contexts
		return ConflictingContext()
	}
	// Matching contexts
	return p
}

func (p Context) String() string {
	return fmt.Sprintf("%d*%d", p.module, p.multiplier)
}
