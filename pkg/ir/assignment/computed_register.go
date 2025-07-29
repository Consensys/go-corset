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
	"github.com/consensys/go-corset/pkg/util/field"
	bls12_377 "github.com/consensys/go-corset/pkg/util/field/bls12-377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
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
	Expr ir.Evaluable
}

// NewComputedRegister constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedRegister(column schema.RegisterRef, expr ir.Evaluable) *ComputedRegister {
	return &ComputedRegister{column, expr}
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
func (p *ComputedRegister) Compute(
	tr trace.Trace[bls12_377.Element],
	schema schema.AnySchema,
) ([]trace.ArrayColumn, error) {
	var (
		trModule = tr.Module(p.Target.Module())
		scModule = schema.Module(p.Target.Module())
		register = schema.Register(p.Target)
	)
	// Determine multiplied height
	height := trModule.Height()
	// FIXME: using an index array here ensures the underlying data is
	// represented using a full field element, rather than e.g. some smaller
	// number of bytes.  This is needed to handle reject tests which can produce
	// values outside the range of the computed register, but which we still
	// want to check are actually rejected (i.e. since they are simulating what
	// an attacker might do).
	data := field.NewFrIndexArray(height, register.Width)
	// Expand the trace
	for i := uint(0); i < data.Len(); i++ {
		val, err := p.Expr.EvalAt(int(i), trModule, scModule)
		// error check
		if err != nil {
			return nil, err
		}
		//
		data.Set(i, val)
	}
	// Determine padding value.  A negative row index is used here to ensure
	// that all columns return their padding value which is then used to compute
	// the padding value for *this* column.
	padding, err := p.Expr.EvalAt(-1, trModule, scModule)
	// Construct column
	col := trace.NewArrayColumn(register.Name, data, padding)
	// Done
	return []trace.ArrayColumn{col}, err
}

// Consistent performs some simple checks that the given assignment is
// consistent with its enclosing schema This provides a double check of certain
// key properties, such as that registers used for assignments are valid,
// etc.
func (p *ComputedRegister) Consistent(schema sc.AnySchema) []error {
	return nil
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
	//
	return rids
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
			sexp.NewSymbol(target.QualifiedName(module)),
			sexp.NewSymbol(datatype),
			p.Expr.Lisp(module),
		})
}

func init() {
	gob.Register(sc.Assignment(&ComputedRegister{}))
}
