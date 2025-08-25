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
package ir

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
)

// PaddingFor determines the appropriate padding value for the given term.  This
// is done by evaluating the term against a artificially constructed trace from
// the given module, where each column has its declared padding value.
func PaddingFor[F field.Element[F], T Evaluable[F]](term T, mod schema.Module[F]) big.Int {
	var (
		num      big.Int
		val, err = term.EvalAt(-1, &traceModule[F]{mod}, mod)
	)
	// Sanity check
	if err != nil {
		panic(fmt.Sprintf("internal failure:: %s", err.Error()))
	}
	// Extract big integer
	num.SetBytes(val.Bytes())
	//
	return num
}

// Wrapper around a schema module making it look like a trace.
type traceModule[F field.Element[F]] struct {
	mod schema.Module[F]
}

// Name implementation for trace.Module interface
func (p *traceModule[F]) Name() string {
	return p.mod.Name()
}

// Column implementation for trace.Module interface
func (p *traceModule[F]) Column(index uint) trace.Column[F] {
	var (
		ith     = p.mod.Register(schema.NewRegisterId(index))
		padding F
	)
	// Convert bigint to field element
	padding = padding.SetBytes(ith.Padding.Bytes())
	//
	return &traceColumn[F]{ith.Name, padding}
}

// ColumnOf implementation for trace.Module interface
func (p *traceModule[F]) ColumnOf(string) trace.Column[F] {
	panic("unreachable")
}

// Width implementation for trace.Module interface
func (p *traceModule[F]) Width() uint {
	return p.mod.Width()
}

// Height implementation for trace.Module interface
func (p *traceModule[F]) Height() uint {
	panic("unreachable")
}

type traceColumn[F field.Element[F]] struct {
	name    string
	padding F
}

// Name implementation for trace.Column interface.
func (p *traceColumn[F]) Name() string {
	return p.name
}

// Get implementation for trace.Column interface.
func (p *traceColumn[F]) Get(_ int) F {
	return p.padding
}

// Data implementation for trace.Column interface.
func (p *traceColumn[F]) Data() array.Array[F] {
	panic("unreachable")
}

// Padding implementation for trace.Column interface.
func (p *traceColumn[F]) Padding() F {
	panic("unreachable")
}
