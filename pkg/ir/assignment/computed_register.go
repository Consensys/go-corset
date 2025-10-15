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
	"fmt"
	"math"
	"slices"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
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
	Targets []schema.RegisterId
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	Expr ir.Computation[F]
	// Direction in which value is computed (true = forward, false = backward).
	// More specifically, a forwards direction means the computation starts on
	// the first row, whilst a backwards direction means it starts on the last.
	Direction bool
}

// NewComputedRegister constructs a new set of computed column(s) with a given
// determining expression.  More specifically, that expression is used to
// compute the values for the columns during trace expansion.  For each, the
// resulting value is split across the target columns.
func NewComputedRegister[F field.Element[F]](expr ir.Computation[F], dir bool, module schema.ModuleId,
	limbs ...schema.RegisterId) *ComputedRegister[F] {
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
		trModule = tr.Module(p.Module)
		scModule = schema.Module(p.Module)
		wrapper  = recursiveModule[F]{p.Targets, nil, trModule}
		err      error
	)
	// Determine multiplied height
	height := trModule.Height()
	wrapper.data = make([]array.MutArray[F], len(p.Targets))
	//
	for i := range p.Targets {
		// FIXME: using a large bitwidth here ensures the underlying data is
		// represented using a full field element, rather than e.g. some smaller
		// number of bytes.  This is needed to handle reject tests which can produce
		// values outside the range of the computed register, but which we still
		// want to check are actually rejected (i.e. since they are simulating what
		// an attacker might do).
		wrapper.data[i] = tr.Builder().NewArray(height, math.MaxUint)
	}
	// Expand the trace
	if !p.IsRecursive() {
		// Non-recursive computation
		err = fwdComputation(height, wrapper.data, p.Expr, trModule, scModule)
	} else if p.Direction {
		// Forwards recursive computation
		err = fwdComputation(height, wrapper.data, p.Expr, &wrapper, scModule)
	} else {
		// Backwards recursive computation
		err = bwdComputation(height, wrapper.data, p.Expr, &wrapper, scModule)
	}
	// Sanity check
	if err != nil {
		return nil, err
	}
	// Done
	return wrapper.data, err
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
		rid := sc.NewRegisterId(iter.Next())
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
func (p *ComputedRegister[F]) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ComputedRegister[F]) RegistersRead() []schema.RegisterRef {
	var (
		module = p.Module
		regs   = p.Expr.RequiredRegisters()
		rids   = make([]schema.RegisterRef, regs.Iter().Count())
	)
	//
	for i, iter := 0, regs.Iter(); iter.HasNext(); i++ {
		rid := sc.NewRegisterId(iter.Next())
		rids[i] = schema.NewRegisterRef(module, rid)
	}
	// Remove target to allow recursive definitions.  Observe this does not
	// guarantee they make sense!
	return array.RemoveMatching(rids, func(r schema.RegisterRef) bool {
		if r.Module() == p.Module {
			for _, id := range p.Targets {
				if id == r.Column() {
					return true
				}
			}
		}
		//
		return false
	})
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *ComputedRegister[F]) RegistersWritten() []sc.RegisterRef {
	var written = make([]schema.RegisterRef, len(p.Targets))
	//
	for i, r := range p.Targets {
		written[i] = schema.NewRegisterRef(p.Module, r)
	}
	//
	return written
}

// Subdivide implementation for the FieldAgnostic interface.
func (p *ComputedRegister[F]) Subdivide(mapping schema.LimbsMap) sc.Assignment[F] {
	//return p
	panic("got here")
}

// Substitute any matchined labelled constants within this assignment
func (p *ComputedRegister[F]) Substitute(mapping map[string]F) {
	p.Expr.Substitute(mapping)
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
		if ith.Width != math.MaxUint {
			datatype = fmt.Sprintf("u%d", ith.Width)
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

func fwdComputation[F field.Element[F]](height uint, data []array.MutArray[F], expr ir.Evaluable[F],
	trMod trace.Module[F], scMod schema.Module[F]) error {
	// Forwards computation
	for i := uint(0); i < height; i++ {
		val, err := expr.EvalAt(int(i), trMod, scMod)
		// error check
		if err != nil {
			return err
		}
		// FIXME: this is completely broken.
		data[0].Set(i, val)
	}
	//
	return nil
}

func bwdComputation[F field.Element[F]](height uint, data []array.MutArray[F], expr ir.Evaluable[F],
	trMod trace.Module[F], scMod schema.Module[F]) error {
	// Backwards computation
	for i := height; i > 0; i-- {
		val, err := expr.EvalAt(int(i-1), trMod, scMod)
		// error check
		if err != nil {
			return err
		}
		// FIXME: this is completely broken.
		data[0].Set(i-1, val)
	}
	//
	return nil
}

// RecModule is a wrapper which enables a computation to be recursive.
// Specifically, it allows the expression being evaluated to access as it is
// being generated.
type recursiveModule[F field.Element[F]] struct {
	col      []schema.RegisterId
	data     []array.MutArray[F]
	trModule trace.Module[F]
}

// Module implementation for trace.Module interface.
func (p *recursiveModule[F]) Name() string {
	return p.trModule.Name()
}

// Column implementation for trace.Module interface.
func (p *recursiveModule[F]) Column(index uint) trace.Column[F] {
	for i, cid := range p.col {
		if cid.Unwrap() == index {
			return &recursiveColumn[F]{p.data[i]}
		}
	}

	return p.trModule.Column(index)
}

// ColumnOf implementation for trace.Module interface.
func (p *recursiveModule[F]) ColumnOf(string) trace.Column[F] {
	// NOTE: this is marked unreachable because, as it stands, expression
	// evaluation never calls this method.
	panic("unreachable")
}

// Width implementation for trace.Module interface.
func (p *recursiveModule[F]) Width() uint {
	return p.trModule.Width()
}

// Height implementation for trace.Module interface.
func (p *recursiveModule[F]) Height() uint {
	return p.trModule.Height()
}

// RecColumn is a wrapper which enables the array being computed to be accessed
// during its own computation.
type recursiveColumn[F field.Element[F]] struct {
	data array.MutArray[F]
}

// Holds the name of this column
func (p *recursiveColumn[F]) Name() string {
	panic("unreachable")
}

// Get implementation for trace.Column interface.
func (p *recursiveColumn[F]) Get(row int) F {
	if row < 0 || uint(row) >= p.data.Len() {
		// out-of-bounds access
		return field.Zero[F]()
	}
	//
	return p.data.Get(uint(row))
}

// Data implementation for trace.Column interface.
func (p *recursiveColumn[F]) Data() array.Array[F] {
	panic("unreachable")
}

// Padding implementation for trace.Column interface.
func (p *recursiveColumn[F]) Padding() F {
	panic("unreachable")
}
