package hir

import (
	"encoding/gob"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ============================================================================
// Expressions
// ============================================================================

// Expr is an expression in the High-Level Intermediate Representation (HIR).
// Expressions at this level have a many-2-one correspondance with expressions
// in the AIR level.  For example, an "if" expression at this level will be
// "compiled out" into one or more expressions at the MIR level.
type Expr interface {
	util.Boundable
	sc.Contextual
	// LowerTo lowers this expression into the Mid-Level Intermediate
	// Representation.  Observe that a single expression at this
	// level can expand into *multiple* expressions at the MIR
	// level.
	LowerTo(*mir.Schema) []mir.Expr
	// EvalAt evaluates this expression in a given tabular context.
	// Observe that if this expression is *undefined* within this
	// context then it returns "nil".  An expression can be
	// undefined for several reasons: firstly, if it accesses a
	// row which does not exist (e.g. at index -1); secondly, if
	// it accesses a column which does not exist.
	EvalAllAt(int, trace.Trace) []fr.Element

	// Multiplicity returns the number of underlyg expressions that this
	// expression will expand to.
	Multiplicity() uint
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

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Add) Multiplicity() uint {
	count := uint(0)
	//
	for _, e := range p.Args {
		count += e.Multiplicity()
	}
	//
	return count
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

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Sub) Multiplicity() uint {
	count := uint(0)
	//
	for _, e := range p.Args {
		count += e.Multiplicity()
	}
	//
	return count
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

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Mul) Multiplicity() uint {
	count := uint(0)
	//
	for _, e := range p.Args {
		count += e.Multiplicity()
	}
	//
	return count
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

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Exp) Multiplicity() uint {
	return p.Arg.Multiplicity()
}

// ============================================================================
// List
// ============================================================================

// List represents a block of zero or more expressions.
type List struct{ Args []Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *List) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *List) Context(schema sc.Schema) trace.Context {
	return sc.JoinContexts[Expr](p.Args, schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *List) RequiredColumns() *util.SortedSet[uint] {
	return util.UnionSortedSets(p.Args, func(e Expr) *util.SortedSet[uint] {
		return e.RequiredColumns()
	})
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *List) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return util.UnionAnySortedSets(p.Args, func(e Expr) *util.AnySortedSet[trace.CellRef] {
		return e.RequiredCells(row, tr)
	})
}

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *List) Multiplicity() uint {
	count := uint(0)
	//
	for _, e := range p.Args {
		count += e.Multiplicity()
	}
	//
	return count
}

// ============================================================================
// Constant
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Val fr.Element }

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

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Constant) Multiplicity() uint { return 1 }

// ============================================================================
// IfZero
// ============================================================================

// IfZero returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero struct {
	// Elements contained within this list.
	Condition Expr
	// True branch (optional).
	TrueBranch Expr
	// False branch (optional).
	FalseBranch Expr
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *IfZero) Bounds() util.Bounds {
	c := p.Condition.Bounds()
	// Get bounds for true branch (if applicable)
	if p.TrueBranch != nil {
		tbounds := p.TrueBranch.Bounds()
		c.Union(&tbounds)
	}
	// Get bounds for false branch (if applicable)
	if p.FalseBranch != nil {
		fbounds := p.FalseBranch.Bounds()
		c.Union(&fbounds)
	}
	// Done
	return c
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p *IfZero) Context(schema sc.Schema) trace.Context {
	if p.TrueBranch != nil && p.FalseBranch != nil {
		args := []Expr{p.Condition, p.TrueBranch, p.FalseBranch}
		return sc.JoinContexts[Expr](args, schema)
	} else if p.TrueBranch != nil {
		// FalseBranch == nil
		args := []Expr{p.Condition, p.TrueBranch}
		return sc.JoinContexts[Expr](args, schema)
	}
	// TrueBranch == nil
	args := []Expr{p.Condition, p.FalseBranch}

	return sc.JoinContexts[Expr](args, schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p *IfZero) RequiredColumns() *util.SortedSet[uint] {
	set := p.Condition.RequiredColumns()
	// Include true branch (if applicable)
	if p.TrueBranch != nil {
		set.InsertSorted(p.TrueBranch.RequiredColumns())
	}
	// Include false branch (if applicable)
	if p.FalseBranch != nil {
		set.InsertSorted(p.FalseBranch.RequiredColumns())
	}
	// Done
	return set
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (p *IfZero) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	set := p.Condition.RequiredCells(row, tr)
	// Include true branch (if applicable)
	if p.TrueBranch != nil {
		set.InsertSorted(p.TrueBranch.RequiredCells(row, tr))
	}
	// Include false branch (if applicable)
	if p.FalseBranch != nil {
		set.InsertSorted(p.FalseBranch.RequiredCells(row, tr))
	}
	// Done
	return set
}

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *IfZero) Multiplicity() uint {
	count := uint(0)
	// TrueBranch (if applicable)
	if p.TrueBranch != nil {
		count += p.TrueBranch.Multiplicity()
	}
	// FalseBranch (if applicable)
	if p.FalseBranch != nil {
		count += p.FalseBranch.Multiplicity()
	}
	// done
	return count
}

// ============================================================================
// Normalise
// ============================================================================

// Normalise reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Normalise struct{ Arg Expr }

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *Normalise) Bounds() util.Bounds {
	return p.Arg.Bounds()
}

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

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *Normalise) Multiplicity() uint { return p.Arg.Multiplicity() }

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

// Multiplicity returns the number of underlyg expressions that this
// expression will expand to.
func (p *ColumnAccess) Multiplicity() uint { return 1 }

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(Expr(&Add{}))
	gob.Register(Expr(&Mul{}))
	gob.Register(Expr(&Sub{}))
	gob.Register(Expr(&Exp{}))
	gob.Register(Expr(&IfZero{}))
	gob.Register(Expr(&List{}))
	gob.Register(Expr(&Constant{}))
	gob.Register(Expr(&Normalise{}))
	gob.Register(Expr(&ColumnAccess{}))
}
