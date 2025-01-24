package mir

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Expr represents an expression in the Mid-Level Intermediate Representation
// (MIR).  Expressions at this level have a one-2-one correspondance with
// expressions in the AIR level.  However, some expressions at this level do not
// exist at the AIR level (e.g. normalise) and are "compiled out" by introducing
// appropriate computed columns and constraints.
type Expr interface {
	util.Boundable
	sc.Evaluable

	// IntRange computes a conservative approximation for the set of possible
	// values that this expression can evaluate to.
	IntRange(schema sc.Schema) *util.Interval
}

// ============================================================================
// Addition
// ============================================================================

// Add represents the sum over zero or more expressions.
type Add struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Add) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Add) Context(schema sc.Schema) trace.Context {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *Add) RequiredColumns() *util.SortedSet[uint] {
	return util.UnionSortedSets(p.Args, func(e Expr) *util.SortedSet[uint] {
		return e.RequiredColumns()
	})
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *Add) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return util.UnionAnySortedSets(p.Args, func(e Expr) *util.AnySortedSet[trace.CellRef] {
		return e.RequiredCells(row, tr)
	})
}

// IntRange computes a conservative approximation for the set of possible
// values that this expression can evaluate to.
func (p *Add) IntRange(schema sc.Schema) *util.Interval {
	var res util.Interval

	for i, arg := range p.Args {
		ith := arg.IntRange(schema)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Add(ith)
		}
	}
	//
	return &res
}

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Sub) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Sub) Context(schema sc.Schema) trace.Context {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *Sub) RequiredColumns() *util.SortedSet[uint] {
	return util.UnionSortedSets(p.Args, func(e Expr) *util.SortedSet[uint] {
		return e.RequiredColumns()
	})
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *Sub) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return util.UnionAnySortedSets(p.Args, func(e Expr) *util.AnySortedSet[trace.CellRef] {
		return e.RequiredCells(row, tr)
	})
}

// IntRange computes a conservative approximation for the set of possible
// values that this expression can evaluate to.
func (p *Sub) IntRange(schema sc.Schema) *util.Interval {
	var res util.Interval

	for i, arg := range p.Args {
		ith := arg.IntRange(schema)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Sub(ith)
		}
	}
	//
	return &res
}

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Mul) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Mul) Context(schema sc.Schema) trace.Context {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *Mul) RequiredColumns() *util.SortedSet[uint] {
	return util.UnionSortedSets(p.Args, func(e Expr) *util.SortedSet[uint] {
		return e.RequiredColumns()
	})
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *Mul) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return util.UnionAnySortedSets(p.Args, func(e Expr) *util.AnySortedSet[trace.CellRef] {
		return e.RequiredCells(row, tr)
	})
}

// IntRange computes a conservative approximation for the set of possible
// values that this expression can evaluate to.
func (p *Mul) IntRange(schema sc.Schema) *util.Interval {
	var res util.Interval

	for i, arg := range p.Args {
		ith := arg.IntRange(schema)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Mul(ith)
		}
	}
	//
	return &res
}

// ============================================================================
// Exponentiation
// ============================================================================

// Exp represents the a given value taken to a power.
type Exp struct {
	Arg Expr
	Pow uint64
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Exp) Bounds() util.Bounds { return p.Arg.Bounds() }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Exp) Context(schema sc.Schema) trace.Context {
	return p.Arg.Context(schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *Exp) RequiredColumns() *util.SortedSet[uint] {
	return p.Arg.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *Exp) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, tr)
}

// IntRange computes a conservative approximation for the set of possible
// values that this expression can evaluate to.
func (p *Exp) IntRange(schema sc.Schema) *util.Interval {
	bounds := p.Arg.IntRange(schema)
	bounds.Exp(uint(p.Pow))
	//
	return bounds
}

// ============================================================================
// Constant
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Value fr.Element }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).  A constant has zero shift.
func (p *Constant) Bounds() util.Bounds { return util.EMPTY_BOUND }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Constant) Context(schema sc.Schema) trace.Context {
	return trace.VoidContext[uint]()
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *Constant) RequiredColumns() *util.SortedSet[uint] {
	return util.NewSortedSet[uint]()
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *Constant) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return util.NewAnySortedSet[trace.CellRef]()
}

// IntRange computes a conservative approximation for the set of possible
// values that this expression can evaluate to.
func (p *Constant) IntRange(schema sc.Schema) *util.Interval {
	var c big.Int
	// Extract big integer from field element
	p.Value.BigInt(&c)
	// Return as interval
	return util.NewInterval(&c, &c)
}

// ============================================================================
// Normalise
// ============================================================================

// Normalise reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Normalise struct{ Arg Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Normalise) Bounds() util.Bounds { return p.Arg.Bounds() }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *Normalise) Context(schema sc.Schema) trace.Context {
	return p.Arg.Context(schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *Normalise) RequiredColumns() *util.SortedSet[uint] {
	return p.Arg.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *Normalise) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, tr)
}

// IntRange computes a conservative approximation for the set of possible
// values that this expression can evaluate to.
func (p *Normalise) IntRange(schema sc.Schema) *util.Interval {
	return util.NewInterval(big.NewInt(0), big.NewInt(1))
}

// ============================================================================
// ColumnAccess
// ============================================================================

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount. Suppose we are evaluating a constraint on row k=5 which
// contains the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)
// accesses the STAMP column at row 5, whilst CT(-1) accesses the CT column at
// row 4.
type ColumnAccess struct {
	Column uint
	Shift  int
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *ColumnAccess) Bounds() util.Bounds {
	if p.Shift >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.Shift))
	}
	// Negative shift
	return util.NewBounds(uint(-p.Shift), 0)
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *ColumnAccess) Context(schema sc.Schema) trace.Context {
	col := schema.Columns().Nth(p.Column)
	return col.Context
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *ColumnAccess) RequiredColumns() *util.SortedSet[uint] {
	r := util.NewSortedSet[uint]()
	r.Insert(p.Column)
	// Done
	return r
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (p *ColumnAccess) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	set := util.NewAnySortedSet[trace.CellRef]()
	set.Insert(trace.NewCellRef(p.Column, row+p.Shift))

	return set
}

// IntRange computes a conservative approximation for the set of possible
// values that this expression can evaluate to.
func (p *ColumnAccess) IntRange(schema sc.Schema) *util.Interval {
	bound := big.NewInt(2)
	width := int64(schema.Columns().Nth(p.Column).DataType.BitWidth())
	bound.Exp(bound, big.NewInt(width), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	// Done
	return util.NewInterval(big.NewInt(0), bound)
}
