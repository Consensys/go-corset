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

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ComputedRegister describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedRegister struct {
	// Module index for computed column
	module uint
	// Target index for computed column
	target schema.RegisterId
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	expr ir.Evaluable
}

// NewComputedRegister constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedRegister(context trace.Context, column schema.RegisterId, expr ir.Evaluable) *ComputedRegister {
	return &ComputedRegister{context.ModuleId, column, expr}
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *ComputedRegister) Bounds() util.Bounds {
	return p.expr.Bounds()
}

// Compute the values of columns defined by this assignment. Specifically, this
// creates a new column which contains the result of evaluating a given
// expression on each row.
func (p *ComputedRegister) Compute(tr trace.Trace, schema schema.AnySchema) ([]trace.ArrayColumn, error) {
	var (
		module   = tr.Module(p.module)
		register = schema.Module(p.module).Register(p.target)
	)
	// Determine multiplied height
	height := module.Height()
	// FIXME: using an index array here ensures the underlying data is
	// represented using a full field element, rather than e.g. some smaller
	// number of bytes.  This is needed to handle reject tests which can produce
	// values outside the range of the computed register, but which we still
	// want to check are actually rejected (i.e. since they are simulating what
	// an attacker might do).
	data := field.NewFrIndexArray(height, register.Width)
	// Expand the trace
	for i := uint(0); i < data.Len(); i++ {
		val, err := p.expr.EvalAt(int(i), module)
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
	padding, err := p.expr.EvalAt(-1, module)
	// Construct column
	col := trace.NewArrayColumn(trace.NewContext(p.module, 1), register.Name, data, padding)
	// Done
	return []trace.ArrayColumn{col}, err
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ComputedRegister) Dependencies() []schema.RegisterId {
	var (
		regs = p.expr.RequiredRegisters()
		rids = make([]schema.RegisterId, regs.Iter().Count())
	)
	//
	for i, iter := 0, regs.Iter(); iter.HasNext(); i++ {
		rids[i] = schema.NewRegisterId(iter.Next())
	}
	//
	return rids
}

// Consistent performs some simple checks that the given assignment is
// consistent with its enclosing schema This provides a double check of certain
// key properties, such as that registers used for assignments are valid,
// etc.
func (p *ComputedRegister) Consistent(schema schema.AnySchema) []error {
	// Check target module exists
	if p.module >= schema.Width() {
		return []error{fmt.Errorf("invalid module (%d >= %d)", p.module, schema.Width())}
	}
	// Check target register exists
	var module = schema.Module(p.module)
	//
	if p.target.Unwrap() >= module.Width() {
		err := fmt.Errorf("invalid register in module %s (%d >= %d)", module.Name(), p.target, module.Width())
		return []error{err}
	}
	// Check register is supposed to be computed.
	var reg = module.Register(p.target)
	//
	if !reg.IsComputed() {
		err := fmt.Errorf("register %s in module %s is not computed", reg.Name, module.Name())
		return []error{err}
	}
	//
	return nil
}

// Module returns the enclosing register for all columns computed by this
// assignment.
func (p *ComputedRegister) Module() uint {
	return p.module
}

// Registers identifies registers assigned by this assignment.
func (p *ComputedRegister) Registers() []schema.RegisterId {
	return []schema.RegisterId{p.target}
}

// Lisp converts this constraint into an S-Expression.
//
//nolint:revive
func (p *ComputedRegister) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		module = schema.Module(p.module)
		target = module.Register(p.target)
	)
	//
	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("compute"),
			sexp.NewSymbol(target.QualifiedName(module)),
			sexp.NewSymbol(fmt.Sprintf("u%d", target.Width)),
			p.expr.Lisp(module),
		})
}
