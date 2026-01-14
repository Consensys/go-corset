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
package assignment

import (
	"encoding/gob"
	"fmt"
	"math"
	"slices"

	"github.com/consensys/go-corset/pkg/ir/term"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ComputedRegister describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedRegister[F field.Element[F]] struct {
	// Module in which expression is evaluated
	Module schema.ModuleId
	// Target indices for computed column
	Targets []register.Id
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	Expr term.Computation[word.BigEndian]
	// Direction in which value is computed (true = forward, false = backward).
	// More specifically, a forwards direction means the computation starts on
	// the first row, whilst a backwards direction means it starts on the last.
	Direction bool
}

// NewComputedRegister constructs a new set of computed column(s) with a given
// determining expression.  More specifically, that expression is used to
// compute the values for the columns during trace expansion.  For each, the
// resulting value is split across the target columns.
func NewComputedRegister[F field.Element[F]](expr term.Computation[word.BigEndian], dir bool, module schema.ModuleId,
	limbs ...register.Id) *ComputedRegister[F] {
	//
	if len(limbs) == 0 {
		panic("computed register requires at least one limb")
	}
	//
	return &ComputedRegister[F]{module, limbs, expr, dir}
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *ComputedRegister[F]) Bounds(mid sc.ModuleId) util.Bounds {
	if mid == p.Module {
		return p.Expr.Bounds()
	}
	// Not relevant
	return util.EMPTY_BOUND
}

// Compute the values of columns defined by this assignment. Specifically, this
// creates a new column which contains the result of evaluating a given
// expression on each row.
func (p *ComputedRegister[F]) Compute(tr trace.Trace[F], schema schema.AnySchema[F],
) ([]array.MutArray[F], error) {
	var (
		trModule = trace.ModuleAdapter[F, word.BigEndian](tr.Module(p.Module))
		scModule = schema.Module(p.Module)
		wrapper  = recursiveModule{p.Targets, nil, trModule}
		err      error
	)
	// Determine multiplied height
	height := trModule.Height()
	bitwidths := make([]uint, len(p.Targets))
	wrapper.data = make([][]word.BigEndian, len(p.Targets))
	//
	for i, target := range p.Targets {
		wrapper.data[i] = make([]word.BigEndian, height)
		// Record bitwidth information
		bitwidths[i] = scModule.Register(target).Width()
	}
	// Expand the trace
	if !p.IsRecursive() {
		// Non-recursive computation
		err = fwdComputation(height, wrapper.data, bitwidths, p.Expr, trModule, scModule, p.Module)
	} else if p.Direction {
		// Forwards recursive computation
		err = fwdComputation(height, wrapper.data, bitwidths, p.Expr, &wrapper, scModule, p.Module)
	} else {
		// Backwards recursive computation
		err = bwdComputation(height, wrapper.data, bitwidths, p.Expr, &wrapper, scModule, p.Module)
	}
	// Sanity check
	if err != nil {
		return nil, err
	}
	// Done
	return concretizeColumns(wrapper.data, tr), err
}

func concretizeColumns[F field.Element[F]](data [][]word.BigEndian, tr trace.Trace[F]) []array.MutArray[F] {
	var cols = make([]array.MutArray[F], len(data))
	//
	for i, d := range data {
		cols[i] = concretizeColumn(d, tr)
	}
	//
	return cols
}

func concretizeColumn[F field.Element[F]](data []word.BigEndian, tr trace.Trace[F]) array.MutArray[F] {
	var (
		// FIXME: using a large bitwidth here ensures the underlying data is
		// represented using a full field element, rather than e.g. some smaller
		// number of bytes.  This is needed to handle reject tests which can produce
		// values outside the range of the computed register, but which we still
		// want to check are actually rejected (i.e. since they are simulating what
		// an attacker might do).
		col = tr.Builder().NewArray(uint(len(data)), math.MaxUint)
	)
	//
	for i, word := range data {
		var element F
		// Assign value
		col.Set(uint(i), element.SetBytes(word.Bytes()))
	}
	//
	return col
}

// Consistent performs some simple checks that the given assignment is
// consistent with its enclosing schema This provides a double check of certain
// key properties, such as that registers used for assignments are valid,
// etc.
func (p *ComputedRegister[F]) Consistent(schema sc.AnySchema[F]) []error {
	return nil
}

// IsRecursive checks whether or not this computation is recursive (i.e. the
// target column is defined in terms of itself).
func (p *ComputedRegister[F]) IsRecursive() bool {
	var regs = p.Expr.RequiredRegisters()
	// Walk through registers accessed by the computation and see whether target
	// register is amongst them.
	for i, iter := 0, regs.Iter(); iter.HasNext(); i++ {
		rid := register.NewId(iter.Next())
		// Did we find it?
		if slices.Contains(p.Targets, rid) {
			// Yes!
			return true
		}
	}
	//
	return false
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *ComputedRegister[F]) RegistersExpanded() []register.Ref {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ComputedRegister[F]) RegistersRead() []register.Ref {
	var (
		module = p.Module
		regs   = p.Expr.RequiredRegisters()
		rids   = make([]register.Ref, regs.Iter().Count())
	)
	//
	for i, iter := 0, regs.Iter(); iter.HasNext(); i++ {
		rid := register.NewId(iter.Next())
		rids[i] = register.NewRef(module, rid)
	}
	// Remove target to allow recursive definitions.  Observe this does not
	// guarantee they make sense!
	return array.RemoveMatching(rids, func(r register.Ref) bool {
		if r.Module() == p.Module {
			return slices.Contains(p.Targets, r.Column())
		}
		//
		return false
	})
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *ComputedRegister[F]) RegistersWritten() []register.Ref {
	var written = make([]register.Ref, len(p.Targets))
	//
	for i, r := range p.Targets {
		written[i] = register.NewRef(p.Module, r)
	}
	//
	return written
}

// Substitute any matchined labelled constants within this assignment
func (p *ComputedRegister[F]) Substitute(mapping map[string]F) {
	var tmp any = mapping
	// NOTE: this is the only scenario under which this method can be called.
	w, ok := tmp.(map[string]word.BigEndian)
	// sanity check
	if !ok {
		panic("unreachable")
	}
	//
	p.Expr.Substitute(w)
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p *ComputedRegister[F]) Lisp(schema sc.AnySchema[F]) sexp.SExp {
	var (
		module  = schema.Module(p.Module)
		targets = make([]sexp.SExp, len(p.Targets))
	)
	//
	for i, t := range p.Targets {
		var (
			datatype = "𝔽"
			ith      = module.Register(t)
		)
		if ith.Width() != math.MaxUint {
			datatype = fmt.Sprintf("u%d", ith.Width())
		}

		targets[i] = sexp.NewList([]sexp.SExp{
			sexp.NewSymbol(ith.QualifiedName(module)), sexp.NewSymbol(datatype),
		})
	}
	//
	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("compute"),
			sexp.NewList(targets),
			p.Expr.Lisp(false, module),
		})
}

func fwdComputation(height uint, data [][]word.BigEndian, widths []uint, expr term.Evaluable[word.BigEndian],
	trMod trace.Module[word.BigEndian], scMod register.Map, ctx schema.ModuleId) error {
	// Forwards computation
	for i := range height {
		val, err := expr.EvalAt(int(i), trMod, scMod)
		// error check
		if err != nil {
			e := fmt.Sprintf("%s for %s", err.Error(), expr.Lisp(false, scMod).String(true))
			return constraint.NewInternalFailure[word.BigEndian](scMod.Name().String(), ctx, i, expr, e)
		}
		// Write data across limbs
		if !write(i, val, data, widths) {
			// Generate error
			return fmt.Errorf("row %d out-of-bounds (%s not u%d) in module %s for: %s", i, val.String(),
				util_math.Sum(widths...),
				scMod.Name(),
				expr.Lisp(false, scMod).String(true))
		}
	}
	//
	return nil
}

func bwdComputation(height uint, data [][]word.BigEndian, widths []uint, expr term.Evaluable[word.BigEndian],
	trMod trace.Module[word.BigEndian], scMod register.Map, ctx schema.ModuleId) error {
	// Backwards computation
	for i := height; i > 0; i-- {
		val, err := expr.EvalAt(int(i-1), trMod, scMod)
		// error check
		if err != nil {
			e := fmt.Sprintf("%s for %s", err.Error(), expr.Lisp(false, scMod).String(true))
			return constraint.NewInternalFailure[word.BigEndian](scMod.Name().String(), ctx, i-1, expr, e)
		}
		// Write data across limbs
		if !write(i-1, val, data, widths) {
			// Generate error
			return fmt.Errorf("row %d out-of-bounds (%s not u%d) in module %s for: %s", i-1, val.String(),
				util_math.Sum(widths...),
				scMod.Name(),
				expr.Lisp(false, scMod).String(true))
		}
	}
	//
	return nil
}

func write(row uint, val word.BigEndian, data [][]word.BigEndian, bitwidths []uint) bool {
	// FIXME: following is not efficient, as it allocates memory and does quite
	// a lot of work overall.
	var elements, ok = field.SplitWord[word.BigEndian](val, bitwidths)
	//
	if ok {
		for i := range data {
			data[i][row] = elements[i]
		}
	}
	//
	return ok
}

// RecModule is a wrapper which enables a computation to be recursive.
// Specifically, it allows the expression being evaluated to access as it is
// being generated.
type recursiveModule struct {
	col      []register.Id
	data     [][]word.BigEndian
	trModule trace.Module[word.BigEndian]
}

// Module implementation for trace.Module interface.
func (p *recursiveModule) Name() module.Name {
	return p.trModule.Name()
}

// Column implementation for trace.Module interface.
func (p *recursiveModule) Column(index uint) trace.Column[word.BigEndian] {
	for i, cid := range p.col {
		if cid.Unwrap() == index {
			return &recursiveColumn{p.data[i]}
		}
	}

	return p.trModule.Column(index)
}

// ColumnOf implementation for trace.Module interface.
func (p *recursiveModule) ColumnOf(string) trace.Column[word.BigEndian] {
	// NOTE: this is marked unreachable because, as it stands, expression
	// evaluation never calls this method.
	panic("unreachable")
}

// Width implementation for trace.Module interface.
func (p *recursiveModule) Width() uint {
	return p.trModule.Width()
}

// Height implementation for trace.Module interface.
func (p *recursiveModule) Height() uint {
	return p.trModule.Height()
}

// RecColumn is a wrapper which enables the array being computed to be accessed
// during its own computation.
type recursiveColumn struct {
	data []word.BigEndian
}

// Holds the name of this column
func (p *recursiveColumn) Name() string {
	panic("unreachable")
}

// Get implementation for trace.Column interface.
func (p *recursiveColumn) Get(row int) word.BigEndian {
	if row < 0 || row >= len(p.data) {
		// out-of-bounds access
		return field.Zero[word.BigEndian]()
	}
	//
	return p.data[row]
}

// Data implementation for trace.Column interface.
func (p *recursiveColumn) Data() array.Array[word.BigEndian] {
	panic("unreachable")
}

// Padding implementation for trace.Column interface.
func (p *recursiveColumn) Padding() word.BigEndian {
	panic("unreachable")
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Assignment[word.BigEndian](&ComputedRegister[word.BigEndian]{}))
}
