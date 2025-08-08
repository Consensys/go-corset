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

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ComputedRegister describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedRegister struct {
	// Target index for computed column
	Target schema.RegisterRef
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	Expr ir.Evaluable[bls12_377.Element]
	// Direction in which value is computed (true = forward, false = backward).
	// More specifically, a forwards direction means the computation starts on
	// the first row, whilst a backwards direction means it starts on the last.
	Direction bool
}

// NewComputedRegister constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedRegister(column schema.RegisterRef, expr ir.Evaluable[bls12_377.Element], dir bool) *ComputedRegister {
	return &ComputedRegister{column, expr, dir}
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *ComputedRegister) Bounds(mid sc.ModuleId) util.Bounds {
	if mid == p.Target.Module() {
		return p.Expr.Bounds()
	}
	// Not relevant
	return util.EMPTY_BOUND
}

// Compute the values of columns defined by this assignment. Specifically, this
// creates a new column which contains the result of evaluating a given
// expression on each row.
func (p *ComputedRegister) Compute(tr trace.Trace[bls12_377.Element], schema schema.AnySchema,
) ([]array.MutArray[bls12_377.Element], error) {
	var (
		trModule = tr.Module(p.Target.Module())
		scModule = schema.Module(p.Target.Module())
		wrapper  = recModule[bls12_377.Element]{p.Target.Column().Unwrap(), nil, trModule}
		register = schema.Register(p.Target)
		err      error
	)
	// Determine multiplied height
	height := trModule.Height()
	// FIXME: using an index array here ensures the underlying data is
	// represented using a full field element, rather than e.g. some smaller
	// number of bytes.  This is needed to handle reject tests which can produce
	// values outside the range of the computed register, but which we still
	// want to check are actually rejected (i.e. since they are simulating what
	// an attacker might do).
	wrapper.data = word.NewIndexArray(height, register.Width, tr.Pool())
	// Expand the trace
	if !p.IsRecursive() {
		// Non-recursive computation
		err = fwdComputation(wrapper.data, p.Expr, trModule, scModule)
	} else if p.Direction {
		// Forwards recursive computation
		err = fwdComputation(wrapper.data, p.Expr, &wrapper, scModule)
	} else {
		// Backwards recursive computation
		err = bwdComputation(wrapper.data, p.Expr, &wrapper, scModule)
	}
	// Sanity check
	if err != nil {
		return nil, err
	}
	// Done
	return []array.MutArray[bls12_377.Element]{wrapper.data}, err
}

// Consistent performs some simple checks that the given assignment is
// consistent with its enclosing schema This provides a double check of certain
// key properties, such as that registers used for assignments are valid,
// etc.
func (p *ComputedRegister) Consistent(schema sc.AnySchema) []error {
	return nil
}

// IsRecursive checks whether or not this computation is recursive (i.e. the
// target column is defined in terms of itself).
func (p *ComputedRegister) IsRecursive() bool {
	var regs = p.Expr.RequiredRegisters()
	// Walk through registers accessed by the computation and see whether target
	// register is amongst them.
	for i, iter := 0, regs.Iter(); iter.HasNext(); i++ {
		rid := sc.NewRegisterId(iter.Next())
		// Did we find it?
		if rid == p.Target.Column() {
			// Yes!
			return true
		}
	}
	//
	return false
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *ComputedRegister) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ComputedRegister) RegistersRead() []schema.RegisterRef {
	var (
		module = p.Target.Module()
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
		return r == p.Target
	})
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *ComputedRegister) RegistersWritten() []sc.RegisterRef {
	return []schema.RegisterRef{p.Target}
}

// Subdivide implementation for the FieldAgnostic interface.
func (p *ComputedRegister) Subdivide(mapping schema.LimbsMap) sc.Assignment {
	return p
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p *ComputedRegister) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		module          = schema.Module(p.Target.Module())
		target          = module.Register(p.Target.Register())
		datatype string = "𝔽"
	)
	//
	if target.Width != math.MaxUint {
		datatype = fmt.Sprintf("u%d", target.Width)
	}
	//
	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("compute"),
			sexp.NewList([]sexp.SExp{
				sexp.NewSymbol(target.QualifiedName(module)),
				sexp.NewSymbol(datatype)}),
			p.Expr.Lisp(false, module),
		})
}

func fwdComputation[F field.Element[F]](data *word.IndexArray[F, word.Pool[uint, F]], expr ir.Evaluable[F],
	trMod trace.Module[F], scMod schema.Module) error {
	// Forwards computation
	for i := uint(0); i < data.Len(); i++ {
		val, err := expr.EvalAt(int(i), trMod, scMod)
		// error check
		if err != nil {
			return err
		}
		//
		data.Set(i, val)
	}
	//
	return nil
}

func bwdComputation[F field.Element[F]](data *word.IndexArray[F, word.Pool[uint, F]], expr ir.Evaluable[F],
	trMod trace.Module[F], scMod schema.Module) error {
	// Backwards computation
	for i := data.Len(); i > 0; i-- {
		val, err := expr.EvalAt(int(i-1), trMod, scMod)
		// error check
		if err != nil {
			return err
		}
		//
		data.Set(i-1, val)
	}
	//
	return nil
}

// RecModule is a wrapper which enables a computation to be recursive.
// Specifically, it allows the expression being evaluated to access as it is
// being generated.
type recModule[F field.Element[F]] struct {
	col      uint
	data     *word.IndexArray[F, word.Pool[uint, F]]
	trModule trace.Module[F]
}

// Module implementation for trace.Module interface.
func (p *recModule[F]) Name() string {
	return p.trModule.Name()
}

// Column implementation for trace.Module interface.
func (p *recModule[F]) Column(index uint) trace.Column[F] {
	if p.col == index {
		return &recColumn[F]{p.data}
	}

	return p.trModule.Column(index)
}

// ColumnOf implementation for trace.Module interface.
func (p *recModule[F]) ColumnOf(string) trace.Column[F] {
	// NOTE: this is marked unreachable because, as it stands, expression
	// evaluation never calls this method.
	panic("unreachable")
}

// Width implementation for trace.Module interface.
func (p *recModule[F]) Width() uint {
	return p.trModule.Width()
}

// Height implementation for trace.Module interface.
func (p *recModule[F]) Height() uint {
	return p.trModule.Height()
}

// RecColumn is a wrapper which enables the array being computed to be accessed
// during its own computation.
type recColumn[F field.Element[F]] struct {
	data *word.IndexArray[F, word.Pool[uint, F]]
}

// Holds the name of this column
func (p *recColumn[F]) Name() string {
	panic("unreachable")
}

// Get implementation for trace.Column interface.
func (p *recColumn[F]) Get(row int) F {
	if row < 0 || uint(row) >= p.data.Len() {
		// out-of-bounds access
		return field.Zero[F]()
	}
	//
	return p.data.Get(uint(row))
}

// Data implementation for trace.Column interface.
func (p *recColumn[F]) Data() array.Array[F] {
	panic("unreachable")
}

// Padding implementation for trace.Column interface.
func (p *recColumn[F]) Padding() F {
	panic("unreachable")
}

func init() {
	gob.Register(sc.Assignment(&ComputedRegister{}))
}
