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

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// PseudoInverse represents a computation which computes the multiplicative
// inverse of a given expression.
type PseudoInverse[F field.Element[F]] struct {
	// Target index for computed column
	Target schema.RegisterRef

	Expr air.Term[F]
}

// NewPseudoInverse constructs a new pseudo-inverse assignment for the given
// target register and expression.
func NewPseudoInverse[F field.Element[F]](target schema.RegisterRef, expr air.Term[F]) *PseudoInverse[F] {
	return &PseudoInverse[F]{
		Target: target,
		Expr:   expr,
	}
}

// Bounds determines the well-definedness bounds for this assignment.
// It is the same as that of the expression it is inverting.
func (e *PseudoInverse[F]) Bounds(mid schema.ModuleId) util.Bounds {
	if mid == e.Target.Module() {
		return e.Expr.Bounds()
	}
	// Not relevant
	return util.EMPTY_BOUND
}

// Compute performs the inversion.
func (e *PseudoInverse[F]) Compute(tr trace.Trace[F], schema schema.AnySchema[F]) ([]array.MutArray[F], error) {
	var (
		trModule = tr.Module(e.Target.Module())
		scModule = schema.Module(e.Target.Module())
		err      error
	)
	// Determine multiplied height
	height := trModule.Height()
	// FIXME: using a large bitwidth here ensures the underlying data is
	// represented using a full field element, rather than e.g. some smaller
	// number of bytes.  This is needed to handle reject tests which can produce
	// values outside the range of the computed register, but which we still
	// want to check are actually rejected (i.e. since they are simulating what
	// an attacker might do).
	data := tr.Builder().NewArray(height, math.MaxUint)
	// Expand the trace
	err = invert(data, e.Expr, trModule, scModule)
	// Sanity check
	if err != nil {
		return nil, err
	}
	// Done
	return []array.MutArray[F]{data}, err
}

// Consistent performs some simple checks that the given assignment is
// consistent with its enclosing schema This provides a double check of certain
// key properties, such as that registers used for assignments are valid,
// etc.
func (e *PseudoInverse[F]) Consistent(schema.AnySchema[F]) []error {
	return nil
}

// RegistersExpanded identifies registers expanded by this assignment.
func (e *PseudoInverse[F]) RegistersExpanded() []schema.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include input columns, as well as other computed columns.
func (e *PseudoInverse[F]) RegistersRead() []schema.RegisterRef {
	var (
		module = e.Target.Module()
		regs   = e.Expr.RequiredRegisters()
		rids   = make([]schema.RegisterRef, regs.Iter().Count())
	)
	//
	for i, iter := 0, regs.Iter(); iter.HasNext(); i++ {
		rid := schema.NewRegisterId(iter.Next())
		rids[i] = schema.NewRegisterRef(module, rid)
	}
	// Remove target to allow recursive definitions.  Observe this does not
	// guarantee they make sense!
	return array.RemoveMatching(rids, func(r schema.RegisterRef) bool {
		return r == e.Target
	})
}

// RegistersWritten identifies registers assigned by this assignment.
func (e *PseudoInverse[F]) RegistersWritten() []schema.RegisterRef {
	return []schema.RegisterRef{e.Target}
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (e *PseudoInverse[F]) Lisp(schema schema.AnySchema[F]) sexp.SExp {
	var (
		module   = schema.Module(e.Target.Module())
		target   = module.Register(e.Target.Register())
		datatype = "𝔽"
	)
	//
	if target.Width != math.MaxUint {
		datatype = fmt.Sprintf("u%d", target.Width)
	}
	//
	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("inv"),
			sexp.NewList([]sexp.SExp{
				sexp.NewSymbol(target.QualifiedName(module)),
				sexp.NewSymbol(datatype)}),
			e.Expr.Lisp(false, module),
		})
}

// RequiredRegisters returns the set of registers on which this term depends.
// That is, registers whose values may be accessed when evaluating this term on
// a given trace.
func (e *PseudoInverse[F]) RequiredRegisters() *set.SortedSet[uint] {
	return e.Expr.RequiredRegisters()
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (e *PseudoInverse[F]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return e.Expr.RequiredCells(row, mid)
}

// Substitute implementation for Substitutable interface.
func (e *PseudoInverse[F]) Substitute(map[string]F) {
	panic("unreachable")
}

func invert[F field.Element[F]](
	data array.MutArray[F],
	expr ir.Evaluable[F],
	trMod trace.Module[F],
	scMod schema.Module[F],
) error {
	// Forwards computation
	for i := range data.Len() {
		val, err := expr.EvalAt(int(i), trMod, scMod)
		// error check
		if err != nil {
			return err
		}
		//
		data.Set(i, val)
	}

	field.BatchInvert(data)

	//
	return nil
}
