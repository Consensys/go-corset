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
	"github.com/consensys/go-corset/pkg/corset/ast"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Computation currently describes a native computation which accepts a set of
// input columns, and assigns a set of output columns.
type ComputeColumn struct {
	// Column being assigned by this computation.
	Target []sc.RegisterRef
	// The formula to get the target column from the source columns.
	Formula ast.Expr
}

// NewComputeColumn defines a set of target columns which are assigned from a given expression.
func NewComputeColumn(targets []sc.RegisterRef, formula ast.Expr) *ComputeColumn {
	//
	return &ComputeColumn{targets, formula}
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *ComputeColumn) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute computes the values of columns defined by this assignment. This
// requires copying the data in the source columns, and sorting that data
// according to the permutation criteria.
func (p *ComputeColumn) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	return trace.Column(p.Target[1]).Data().Set(p.Formula.AsConstant()), nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ComputeColumn) RegistersRead() []sc.RegisterRef {
	return p.Formula.Dependencies()
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *ComputeColumn) RegistersWritten() []sc.RegisterRef {
	return p.Target
}
