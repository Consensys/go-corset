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
package compiler

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/asm/io"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// StampAssignment is a simple assignment which assigns values to the $stamp
// column.
type StampAssignment struct {
	ref sc.RegisterRef
}

// Bounds implementation for schema.Assignment interface.
func (p StampAssignment) Bounds(module uint) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute implementation for schema.Assignment interface.
func (p StampAssignment) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	var (
		zero     = fr.NewElement(0)
		one      = fr.One()
		trModule = trace.Module(p.ref.Module())
		stampReg = schema.Register(p.ref)
		stampCol = field.NewFrArray(trModule.Height(), stampReg.Width)
		pcCol    = trModule.Column(io.PC_INDEX)
		stamp    = fr.NewElement(0)
	)
	//
	for i := range pcCol.Data().Len() {
		ith := pcCol.Get(int(i))
		// Check for incrementing stamp
		if i != 0 && ith.Cmp(&zero) == 0 {
			// Yes, so increment
			stamp.Add(&stamp, &one)
		}
		//
		stampCol.Set(i, stamp)
	}
	// Construct array column
	col := tr.NewArrayColumn(stampReg.Name, stampCol, zero)
	// Done
	return []tr.ArrayColumn{col}, nil
}

// Consistent implementation for schema.Assignment interface.
func (p StampAssignment) Consistent(sc.AnySchema) []error {
	return nil
}

// Lisp implementation for schema.Assignment interface.
func (p StampAssignment) Lisp(schema sc.AnySchema) sexp.SExp {
	var reg = schema.Register(p.ref)
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("compute"),
		sexp.NewSymbol(reg.Name),
	})
}

// RegistersRead implementation for schema.Assignment interface.
func (p StampAssignment) RegistersRead() []sc.RegisterRef {
	rid := sc.NewRegisterId(io.PC_INDEX)
	return []sc.RegisterRef{sc.NewRegisterRef(p.ref.Module(), rid)}
}

// RegistersWritten implementation for schema.Assignment interface.
func (p StampAssignment) RegistersWritten() []sc.RegisterRef {
	return []sc.RegisterRef{p.ref}
}
